package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
)

var (
	groupSyncClean bool
)

var groupSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync the current group (metadata + all subgroups)",
	Long: `Sync the current group's metadata and recursively sync all its subgroups.
This means:
1. Fetch the latest list of subgroups from GitLab.
2. Update the local .ash/group.json file (handle additions, removals, renames).
3. If --clean is used, delete local folders of subgroups that no longer exist on GitLab.
4. Recursively run 'sync' on every subgroup (which pulls code for all projects).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Context Check
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		ashDir := filepath.Join(wd, ".ash")
		groupMetaPath := filepath.Join(ashDir, "group.json")

		if !fileExists(groupMetaPath) {
			return fmt.Errorf("not in a group root (.ash/group.json missing)")
		}

		var meta rootGroupMeta
		if err := readJSON(groupMetaPath, &meta); err != nil {
			return err
		}

		fmt.Printf("Syncing Group: %s (ID: %d)\n", meta.Group.Name, meta.Group.ID)

		// 2. Fetch remote subgroups
		sgs, err := apiListSubgroups(meta.Group.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch subgroups: %w", err)
		}

		// Filter out Marked For Deletion groups
		var validRemoteSGs []glGroup
		for _, sg := range sgs {
			// DEBUG: Print info to verify API response
			// fmt.Printf("[DEBUG] Subgroup %s, MarkedForDeletionOn: '%s'\n", sg.Name, sg.MarkedForDeletionOn)

			if sg.MarkedForDeletionOn == "" {
				validRemoteSGs = append(validRemoteSGs, sg)
			} else {
				fmt.Printf("%s[INFO] Ignoring soft-deleted subgroup: %s%s\n", Yellow, sg.Name, Reset)
			}
		}

		// 3. Diff & Update Metadata
		// Map ID -> Remote Group
		remoteMap := make(map[int64]glGroup)
		for _, sg := range validRemoteSGs {
			remoteMap[sg.ID] = sg
		}

		// Identify removed (Legacy check against old meta)
		// var removedPaths []string // Keep for compatibility if needed, but orphan scan handles deletion better.
		// We still need this loop to detect Renames.

		var newSubgroupIdents []subgroupIdent
		for _, sg := range validRemoteSGs {
			newSubgroupIdents = append(newSubgroupIdents, subgroupIdent{
				ID:   sg.ID,
				Name: sg.Name,
				Path: sg.Path,
			})
		}

		// Check what was removed / Renamed
		for _, oldSg := range meta.Subgroups {
			if _, ok := remoteMap[oldSg.ID]; !ok {
				// Removed or Soft-Deleted
				// Ideally we do nothing here and let orphan scanner kill it.
				// But we can verify.
			} else {
				// Still exists, check for RENAME
				newSg := remoteMap[oldSg.ID]
				if newSg.Name != oldSg.Name {
					fmt.Printf("%s[INFO] Subgroup renamed: %s -> %s%s\n", Yellow, oldSg.Name, newSg.Name, Reset)
					oldPath := filepath.Join(wd, oldSg.Name)
					newPath := filepath.Join(wd, newSg.Name)
					if fileExists(oldPath) {
						if err := os.Rename(oldPath, newPath); err != nil {
							fmt.Printf("%s[WARN] Failed to rename local folder: %v%s\n", Yellow, err, Reset)
						} else {
							fmt.Printf("%s[OK] Renamed local folder.%s\n", Green, Reset)
						}
					}
				}
			}
		}

		// Save new meta
		meta.Subgroups = newSubgroupIdents

		// FIX: Check if root group details are missing and update them
		if meta.Group.Name == "" || meta.Group.Path == "" {
			url := fmt.Sprintf("groups/%d", meta.Group.ID)
			out, err := exec.Command("glab", "api", url).Output()
			if err == nil {
				var info glGroup
				if json.Unmarshal(out, &info) == nil {
					meta.Group.Name = info.Name
					meta.Group.Path = info.Path
				}
			}
		}
		if err := writeGroupJSON(ashDir, meta); err != nil {
			return fmt.Errorf("failed to write updated metadata: %w", err)
		}
		fmt.Printf("%s[OK] Metadata updated. %d subgroups found.%s\n", Green, len(newSubgroupIdents), Reset)

		// 4. Handle Cleanup (ROBUST ORPHAN SCAN)
		// Scan local folders that are NOT in the new list.
		if groupSyncClean {
			validNames := make(map[string]bool)
			for _, sg := range newSubgroupIdents {
				validNames[sg.Name] = true
			}

			entries, _ := os.ReadDir(wd)
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				name := e.Name()
				if name == ".ash" || name == ".git" || name == "." || name == ".." {
					continue
				}

				if !validNames[name] {
					// Orphan
					if err := os.RemoveAll(filepath.Join(wd, name)); err != nil {
						fmt.Printf("%s[ERR] Failed to remove orphan subgroup %s: %v%s\n", Red, name, err, Reset)
					} else {
						fmt.Printf("%s[DEL] Removed orphan subgroup: %s%s\n", Red, name, Reset)
					}
				}
			}
		}

		// 5. Recursive Sync (Projects inside subgroups)
		// For each subgroup, we need to run the equivalent of "ash subgroup sync"
		// We can't easily import `subgroupSyncCmd.Run`, but we can extract a helper function.
		// For now, let's implement the logic or call a reusable function.
		// Since we haven't refactored Subgroup Sync yet, let's create a shared helper key logic.

		fmt.Println("Recursively syncing subgroups...")
		var wg sync.WaitGroup
		sem := make(chan struct{}, 3) // Limit concurrency

		for _, sgIdent := range newSubgroupIdents {
			// Compute local path (Name is used for folder)
			sgDir := filepath.Join(wd, sgIdent.Name)

			// If folder doesn't exist, Create it (Scaffold)
			if !fileExists(sgDir) {
				if err := os.MkdirAll(sgDir, 0o755); err != nil {
					fmt.Printf("[ERR] Failed to create folder %s\n", sgIdent.Name)
					continue
				}
				// create .ash/subgroup.json
				emptyMeta := subgroupMeta{
					Group:    groupIdent{ID: sgIdent.ID, Path: sgIdent.Path, Name: sgIdent.Name},
					Projects: []projectIdent{},
				}
				writeSubgroupJSON(filepath.Join(sgDir, ".ash"), emptyMeta)
			}

			wg.Add(1)
			go func(dir string, id int64) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				// Execute Sync logic for this subgroup
				// We can re-use the logic that will be in `subgroup_sync`
				// For now calling a placeholder `syncSubgroupContent(dir, id)`
				if err := syncSubgroupContent(dir, id, groupSyncClean); err != nil {
					fmt.Printf("[ERR] Sync %s failed: %v\n", dir, err)
				}
			}(sgDir, sgIdent.ID)
		}
		wg.Wait()

		return nil
	},
}

func init() {
	groupCmd.AddCommand(groupSyncCmd)
	groupSyncCmd.Flags().BoolVar(&groupSyncClean, "clean", false, "Delete local folders of removed subgroups")
}

// syncSubgroupContent thực hiện logic sync cho 1 subgroup (fetch projects -> clone/pull)
// Hàm này sẽ được dùng chung bởi `ash group sync` và `ash subgroup sync`
func syncSubgroupContent(wd string, subgroupID int64, clean bool) error {
	// 1. Fetch Projects
	prjs, err := apiListProjects(subgroupID)
	if err != nil {
		return err
	}

	// 2. Read/Update Local Meta
	ashDir := filepath.Join(wd, ".ash")
	subMetaPath := filepath.Join(ashDir, "subgroup.json")

	var meta subgroupMeta
	// Try read existing
	_ = readJSON(subMetaPath, &meta)

	// Update Group info (just in case ID matched but passed explicitly)
	// meta.Group.ID = subgroupID (keep what we have or update?)

	oldPrjs := meta.Projects
	remoteMap := make(map[int64]glProject)
	for _, p := range prjs {
		remoteMap[p.ID] = p
	}

	// Detect Removed
	var removedNames []string
	for _, old := range oldPrjs {
		if _, ok := remoteMap[old.ID]; !ok {
			removedNames = append(removedNames, old.Name)
		} else {
			// Check rename?
			// Doing rename for projects is tricky if folder names change.
			newP := remoteMap[old.ID]
			if newP.Name != old.Name {
				// Try rename folder
				oldPath := filepath.Join(wd, old.Name)
				newPath := filepath.Join(wd, newP.Name)
				if fileExists(oldPath) {
					os.Rename(oldPath, newPath)
				}
			}
		}
	}

	// Save new List
	newidents := []projectIdent{}
	for _, p := range prjs {
		newidents = append(newidents, projectIdent{ID: p.ID, Name: p.Name, Path: p.Path})
	}
	meta.Projects = newidents

	// FIX: Update group details from fetch (since we only have ID sometimes)
	// Theoretically we could fetch the Group info again, but here we only listed Projects.
	// However, we should at least preserve existing Name/Path if they are not empty.
	// If they are empty (which caused the issue), we might need to fetch them.
	// BUT, `syncSubgroupContent` takes `subgroupID`.
	// We didn't fetch the Group Info here!
	// We only fetched projects.
	// To fix the issue "path and name ... are empty", we need to fetch the Group info if missing.
	if meta.Group.Name == "" || meta.Group.Path == "" {
		// Fetch Group Details
		// glab api groups/:id
		url := fmt.Sprintf("groups/%d", subgroupID)
		out, err := exec.Command("glab", "api", url).Output()
		if err == nil {
			var gDetails glGroup
			if json.Unmarshal(out, &gDetails) == nil {
				meta.Group.Name = gDetails.Name
				meta.Group.Path = gDetails.Path
			}
		}
	}
	// Also ensure ID is set.
	if meta.Group.ID == 0 {
		meta.Group.ID = subgroupID
	}

	writeSubgroupJSON(ashDir, meta)

	// Clean / Orphan Check
	// Instead of just relying on `removedNames` (which depends on old metadata),
	// we scan the local directory for ANY folder that is NOT in the new project list.
	// This covers cases where metadata was already updated but folders weren't deleted.

	validNames := make(map[string]bool)
	for _, p := range newidents {
		validNames[p.Name] = true
	}

	entries, _ := os.ReadDir(wd)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if name == ".ash" || name == ".git" || name == "." || name == ".." {
			continue
		}

		// If this folder is NOT in the valid list of projects
		if !validNames[name] {
			// It's an orphan (or a leftover from deletion)
			if clean {
				if err := os.RemoveAll(filepath.Join(wd, name)); err != nil {
					fmt.Printf("%s[ERR] Failed to remove orphan %s: %v%s\n", Red, name, err, Reset)
				} else {
					fmt.Printf("%s[DEL] Removed orphan folder: %s%s\n", Red, name, Reset)
				}
			} else {
				fmt.Printf("%s[INFO] Found orphan folder: %s (use --clean to remove)%s\n", Yellow, name, Reset)
			}
		}
	}

	// 3. Sync Code (Clone/Pull)
	// Reuse `syncOneProject` from project_sync.go?
	// Note: syncOneProject needs `projectIdent`
	// We might need to make `syncOneProject` public or move strict logic.
	// `syncOneProject` is in cmd/project_sync.go, it is currently exported (lowercase s? No, Uppercase S? No it's lower.)
	// Need to export `SyncOneProject` or copy logic.
	// Providing a Helper in a separate file is better.

	// We will use a simplified inline logic or call the tool's run command?
	// Better to have pure Go function.
	// Let's assume we rename `syncOneProject` to `SyncOneProject` in project_sync.go later.
	// For now, I will duplicate/adapt the simple clone/pull logic here to avoid circular dependencies if any,
	// or ideally refactor `syncOneProject` into `helpers.go` or keep it in `project_sync.go` and export it.

	// 3. Sync Code (Clone/Pull)
	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)

	for _, p := range newidents {
		wg.Add(1)
		go func(proj projectIdent) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Use the remote info for URL
			remoteP := remoteMap[proj.ID]
			url := remoteP.HTTPURLToRepo

			targetDir := filepath.Join(wd, proj.Name)
			if !fileExists(targetDir) {
				// Clone
				out, err := exec.Command("git", "clone", "--quiet", url, targetDir).CombinedOutput()
				if err != nil {
					fmt.Printf("%s[ERR] Clone %s failed: %v\n%s%s", Red, proj.Name, err, string(out), Reset)
				} else {
					fmt.Printf("%s[NEW] Cloned: %s%s\n", Cyan, proj.Name, Reset)
				}
			} else {
				// Pull
				if fileExists(filepath.Join(targetDir, ".git")) {
					// Update remote URL just in case path changed
					exec.Command("git", "-C", targetDir, "remote", "set-url", "origin", url).Run()

					// Check if pull is needed? "git pull" usually prints "Already up to date" if nothing changed.
					// We can capture output.
					out, err := exec.Command("git", "-C", targetDir, "pull", "--quiet").CombinedOutput()
					if err != nil {
						fmt.Printf("%s[ERR] Pull %s failed: %v\n%s%s", Red, proj.Name, err, string(out), Reset)
					} else {
						// Quiet mode doesn't print anything on success.
						// If we want to know if it updated, we might need to check HEAD before/after or remove --quiet.
						// But for user "sync complete" feeling, maybe just print OK is enough.
						// OR print only if updated?
						// Let's print [OK] Updated or [OK] Checked.
						fmt.Printf("%s[OK] Checked: %s%s\n", Green, proj.Name, Reset)
					}
				} else {
					fmt.Printf("%s[SKIP] %s (folder exists but not git repo)%s\n", Yellow, proj.Name, Reset)
				}
			}
		}(p)
	}
	wg.Wait()

	return nil
}
