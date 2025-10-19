package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/theckman/yacspin"
)

// CloneResult represents the outcome of cloning one repo.
type CloneResult struct {
	Name     string
	URL      string
	Dest     string
	Duration time.Duration
	Err      error
	Stderr   string
}

// runWithSpinner runs fn while showing a spinner (if stderr is a TTY).
// title is displayed while the spinner runs. On success, a green "done" is shown.
// On failure, a red fail message is shown.
func runWithSpinner(title string, fn func() error) error {
	useSpinner := isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())
	if !useSpinner {
		// No spinner in non-interactive env; just run.
		return fn()
	}

	cfg := yacspin.Config{
		Frequency:         100 * time.Millisecond,
		CharSet:           yacspin.CharSets[14], // dots
		Suffix:            " " + title,
		Message:           "Working",
		Colors:            []string{"fgHiCyan"},
		StopCharacter:     "✔",
		StopColors:        []string{"fgHiGreen"},
		StopFailCharacter: "✖",
		StopFailColors:    []string{"fgHiRed"},
		Writer:            os.Stderr,
	}

	sp, err := yacspin.New(cfg)
	if err != nil {
		return fn()
	}
	_ = sp.Start()
	defer func() {
		_ = sp.Stop()
	}()

	if err := fn(); err != nil {
		sp.StopFailMessage(fmt.Sprintf("Done with error: %v", err))
		_ = sp.StopFail()
		return err
	}
	sp.StopMessage("Done")
	_ = sp.Stop()
	return nil
}

// cloneOneRepo runs `git clone` quietly and captures stderr (truncated).
func cloneOneRepo(url, dest, name string) CloneResult {
	start := time.Now()
	var stderr bytes.Buffer
	cmd := exec.Command("git", "clone", "--quiet", url, dest)
	cmd.Stdout = nil
	cmd.Stderr = &stderr

	err := cmd.Run()
	dur := time.Since(start)

	s := stderr.String()
	if len(s) > 600 {
		s = s[:600] + "…"
	}

	return CloneResult{
		Name:     name,
		URL:      url,
		Dest:     dest,
		Duration: dur,
		Err:      err,
		Stderr:   s,
	}
}

// printCloneSummary prints a clean report after spinner stops.
func printCloneSummary(okay, failed []CloneResult) {
	fmt.Println()
	if len(okay) > 0 {
		fmt.Printf("✔ Success (%d)\n", len(okay))
		for _, r := range okay {
			fmt.Printf("   ✔ %s  →  %s  (%s)\n", r.Name, r.Dest, r.Duration.Truncate(time.Millisecond))
		}
	}
	if len(failed) > 0 {
		fmt.Printf("\n✖ Failed (%d)\n", len(failed))
		for _, r := range failed {
			fmt.Printf("   ✖ %s  →  %s\n", r.Name, r.Dest)
			if r.Err != nil {
				fmt.Printf("     Error: %v\n", r.Err)
			}
			if r.Stderr != "" {
				fmt.Printf("     Stderr: %s\n", indent(r.Stderr, "       "))
			}
		}
	}
	fmt.Println()
}

// indent helper (shared)
func indent(s, pad string) string {
	var out bytes.Buffer
	for i, ln := range bytes.Split([]byte(s), []byte{'\n'}) {
		if len(ln) == 0 && i == len(bytes.Split([]byte(s), []byte{'\n'}))-1 {
			continue
		}
		out.WriteString(pad)
		out.Write(ln)
		out.WriteByte('\n')
	}
	return out.String()
}
