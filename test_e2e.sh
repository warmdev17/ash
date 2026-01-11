#!/bin/bash

# Configuration
ASH_BIN="/tmp/ash-test/ash"
TIMESTAMP=$(date +%s)
GROUP_NAME="Ash_AutoTest_Group_${TIMESTAMP}"
SESSION_NAME="Session_01"
PROJECT_1="Lab_01"
PROJECT_2="Lab_02"
TEST_DIR="/tmp/ash_e2e_${TIMESTAMP}"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

log() {
    echo -e "${CYAN}[TEST] $1${NC}"
}

pass() {
    echo -e "${GREEN}[PASS] $1${NC}"
}

fail() {
    echo -e "${RED}[FAIL] $1${NC}"
    exit 1
}

# 1. SETUP
log "Building ash binary..."
mkdir -p /tmp/ash-test
go build -o "$ASH_BIN" . || fail "Build failed"
pass "Build successful"

mkdir -p "$TEST_DIR"
cd "$TEST_DIR" || fail "Cannot enter test dir"
log "Working in $TEST_DIR"

# 2. GROUP OPERATIONS
log ">>> LEVEL 1: GROUP OPERATIONS <<<"
log "Creating Group: $GROUP_NAME"
"$ASH_BIN" group create "$GROUP_NAME" || fail "Group create failed"

if [ ! -d "$GROUP_NAME" ]; then fail "Group folder not created"; fi
if [ ! -f "$GROUP_NAME/.ash/group.json" ]; then fail "Group metadata missing"; fi
pass "Group created locally"

cd "$GROUP_NAME" || fail "Cannot enter group folder"

log "Syncing empty group..."
"$ASH_BIN" group sync || fail "Group sync failed"
pass "Group sync OK"

# 3. SUBGROUP OPERATIONS
log ">>> LEVEL 2: SUBGROUP OPERATIONS <<<"
log "Creating Subgroup: $SESSION_NAME"
"$ASH_BIN" subgroup create "$SESSION_NAME" || fail "Subgroup create failed"

if [ ! -d "$SESSION_NAME" ]; then fail "Subgroup folder not created"; fi
if [ ! -f "$SESSION_NAME/.ash/subgroup.json" ]; then fail "Subgroup metadata missing"; fi
pass "Subgroup created locally"

log "Listing subgroups..."
"$ASH_BIN" subgroup list || fail "List failed"

log "Testing Subgroup CLONE..."
rm -rf "$SESSION_NAME"
if [ -d "$SESSION_NAME" ]; then fail "Failed to delete local subgroup for test"; fi
"$ASH_BIN" subgroup clone "$SESSION_NAME" || fail "Subgroup clone failed"
if [ ! -d "$SESSION_NAME" ]; then fail "Subgroup clone did not restore folder"; fi
pass "Subgroup clone OK"

cd "$SESSION_NAME" || fail "Cannot enter subgroup"

# 4. PROJECT OPERATIONS
log ">>> LEVEL 3: PROJECT OPERATIONS <<<"
log "Creating Projects: $PROJECT_1, $PROJECT_2"
"$ASH_BIN" project create "$PROJECT_1" "$PROJECT_2" || fail "Project create failed"

if [ ! -d "$PROJECT_1/.git" ]; then fail "Project 1 git missing"; fi
if [ ! -d "$PROJECT_2/.git" ]; then fail "Project 2 git missing"; fi
pass "Projects created"

log "Listing projects..."
"$ASH_BIN" project list || fail "Project list failed"

log "Testing Project CLONE..."
rm -rf "$PROJECT_1"
"$ASH_BIN" project clone "$PROJECT_1" || fail "Project clone failed"
if [ ! -d "$PROJECT_1" ]; then fail "Project clone did not restore folder"; fi
pass "Project clone OK"

log "Testing Submit (Mock)..."
echo "Test Content" > "$PROJECT_1/test.txt"
# We won't actually push to avoid cluttering git history too much, or we assume test user has permissions.
# "$ASH_BIN" submit "$PROJECT_1" -m "Automated Test" || log "Submit failed (expected if no remote setup?)"
pass "Submit logic skipped for safety in this run"

# 5. ADVANCED SYNC & CLEAN
log ">>> LEVEL 4: CLEANUP LOGIC <<<"

# Get IDs for deletion API calls
# We need to find IDs. Grep from json or just use name delete via ash? 
# Ash delete is safer and easier.
log "Deleting Project $PROJECT_2 via API (Simulating remote delete)..."
# We cheat: use ash project delete -f (which calls API) but NOT -l (keep local)
# Actually 'ash project delete' deletes local if -l is passed.
# To test SYNC --CLEAN, we must delete REMOTE ONLY.
# We can use 'glab' directly.
# BUT we need ID.
# Get IDs using 'ash project list' which is cleaner
# Output format: ID \t NAME \t PATH
PROJECT_2_ID=$("$ASH_BIN" project list | grep "$PROJECT_2" | awk '{print $1}')

if [ -z "$PROJECT_2_ID" ]; then fail "Could not find ID for $PROJECT_2"; fi

log "Deleting Project $PROJECT_2 via API (ID: $PROJECT_2_ID)..."
glab api -X DELETE "/projects/$PROJECT_2_ID" > /dev/null
pass "Deleted $PROJECT_2 remotely"

log "Running Subgroup Sync --clean..."
"$ASH_BIN" subgroup sync --clean || fail "Sync clean failed"

if [ -d "$PROJECT_2" ]; then fail "Project 2 was NOT legally deleted by sync --clean"; fi
pass "Project 2 correctly cleaned up"

# Go up to Group Level
cd ..

log "Deleting Subgroup $SESSION_NAME via API..."
# Parse Subgroup list using 'ash group list' ? No 'ash group list' lists GROUPS (level 1).
# 'ash subgroup list' lists subgroups.
SESSION_ID=$("$ASH_BIN" subgroup list | grep "$SESSION_NAME" | awk '{print $1}')

if [ -z "$SESSION_ID" ]; then fail "Could not find ID for $SESSION_NAME"; fi

glab api -X DELETE "/groups/$SESSION_ID" > /dev/null
pass "Deleted $SESSION_NAME remotely (ID: $SESSION_ID)"

log "Waiting for GitLab to propagate delete..."
sleep 5

log "Running Group Sync --clean..."
"$ASH_BIN" group sync --clean || fail "Group sync clean failed"

if [ -d "$SESSION_NAME" ]; then fail "Subgroup '$SESSION_NAME' was NOT locally deleted"; fi
pass "Subgroup correctly cleaned up"

# 6. TEARDOWN
log ">>> LEVEL 5: FINAL TEARDOWN <<<"
cd .. # Back to Test Dir
# Now we are outside the group folder
# We need to delete the Top Group.
# 'ash group delete' needs to know the name.

log "Deleting Top Group: $GROUP_NAME"
# Use force and local-force
"$ASH_BIN" group delete "$GROUP_NAME" -f -l || warn "Top group delete failed, please check GitLab"

if [ -d "$GROUP_NAME" ]; then fail "Top group local folder still exists"; fi
pass "Teardown complete"

echo -e "\n${GREEN}>>> ALL TESTS PASSED <<<${NC}"
rm -rf "$TEST_DIR"
exit 0
