#!/bin/bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to display usage
usage() {
    echo "Usage: $0 <json_file>"
    echo ""
    echo "JSON format:"
    echo "{"
    echo "  \"test_cmd\": \"make test\","
    echo "  \"mutations\": ["
    echo "    {\"path\": \"src/foo.go\", \"search\": \"if a == b {\", \"replace\": \"if false {\"},"
    echo "    {\"path\": \"src/bar.go\", \"search\": \"return true\", \"replace\": \"return false\"}"
    echo "  ]"
    echo "}"
    echo ""
    echo "The script will apply each mutation, run tests, and verify at least one test fails."
    echo ""
    echo "Requirements: jq, sd"
    exit 1
}

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq not found. Please install jq:${NC}"
    exit 1
fi

# Check if sd is available
if ! command -v sd &> /dev/null; then
    echo -e "${RED}Error: sd not found. Please install sd:${NC}"
    exit 1
fi

# Track the current file being mutated (for cleanup on interrupt)
CURRENT_MUTATED_FILE=""

# Function to restore a backup file
restore_backup() {
    local file="$1"
    local backup="${file}.mutation_backup"
    if [ -f "$backup" ]; then
        mv "$backup" "$file"
    fi
}

# Cleanup function called on script exit or interrupt
cleanup() {
    if [ -n "$CURRENT_MUTATED_FILE" ]; then
        echo ""
        echo -e "${YELLOW}Interrupted! Cleaning up...${NC}"
        restore_backup "$CURRENT_MUTATED_FILE"
        CURRENT_MUTATED_FILE=""
    fi
}

# Set up trap to call cleanup on EXIT, SIGINT (Ctrl+C), and SIGTERM
trap cleanup EXIT INT TERM

# Check arguments
if [ $# -ne 1 ]; then
    usage
fi

JSON_FILE="$1"

# Check if JSON file exists
if [ ! -f "$JSON_FILE" ]; then
    echo -e "${RED}Error: JSON file '$JSON_FILE' not found${NC}"
    exit 1
fi

# Validate JSON format
if ! jq empty "$JSON_FILE" 2>/dev/null; then
    echo -e "${RED}Error: Invalid JSON file${NC}"
    exit 1
fi

# Extract test command
TEST_CMD=$(jq -r '.test_cmd // empty' "$JSON_FILE")

# Check if test command is empty
if [ -z "$TEST_CMD" ]; then
    echo -e "${RED}Error: 'test_cmd' field is missing or empty in JSON${NC}"
    exit 1
fi

# Extract mutations array
MUTATIONS=$(jq -c '.mutations // []' "$JSON_FILE")
TOTAL_MUTATIONS=$(echo "$MUTATIONS" | jq -c '. | length')

# Count total mutations
CURRENT=0
PASSED=0
FAILED=0

# Process each mutation
while read -r row; do
    CURRENT=$((CURRENT + 1))

    # Extract fields using jq
    path=$(echo "$row" | jq -r '.path // empty')
    search=$(echo "$row" | jq -r '.search // empty')
    replace=$(echo "$row" | jq -r '.replace // empty')

    # Check for parsing errors or missing fields
    if [ -z "$path" ] || [ -z "$search" ]; then
        # Output error as failed mutation
        TEST_OUTPUT="Error: Missing required fields (path or search)"
        echo -n "MUTATION|"
        jq -nc --arg name "mutation#$CURRENT" \
               --arg result "failed" \
               --arg path "$path" \
               --arg search "$search" \
               --arg replace "$replace" \
               --arg output "$TEST_OUTPUT" \
               '{"test_name": $name, "result": $result, "path": $path, "search": $search, "replace": $replace, "output": $output}'
        FAILED=$((FAILED + 1))
        exit 10
    fi

    # Check if file exists
    if [ ! -f "$path" ]; then
        TEST_OUTPUT="Error: File '$path' not found"
        echo -n "MUTATION|"
        jq -nc --arg name "mutation#$CURRENT" \
               --arg result "failed" \
               --arg path "$path" \
               --arg search "$search" \
               --arg replace "$replace" \
               --arg output "$TEST_OUTPUT" \
               '{"test_name": $name, "result": $result, "path": $path, "search": $search, "replace": $replace, "output": $output}'
        FAILED=$((FAILED + 1))
        exit 11
    fi

    # Check if search string exists in file using sd preview mode
    if ! sd --preview --string-mode "$search" "" "$path" 2>/dev/null | grep -q .; then
        TEST_OUTPUT="Error: Search string not found in file"
        echo -n "MUTATION|"
        jq -nc --arg name "mutation#$CURRENT" \
               --arg result "failed" \
               --arg path "$path" \
               --arg search "$search" \
               --arg replace "$replace" \
               --arg output "$TEST_OUTPUT" \
               '{"test_name": $name, "result": $result, "path": $path, "search": $search, "replace": $replace, "output": $output}'
        FAILED=$((FAILED + 1))
        continue
    fi

    # Track this file for cleanup on interrupt
    CURRENT_MUTATED_FILE="$path"

    # Create backup
    BACKUP_FILE="${path}.mutation_backup"
    cp "$path" "$BACKUP_FILE"

    # Apply mutation using sd (string mode for literal replacement)
    if ! sd --string-mode "$search" "$replace" "$path" 2>/dev/null; then
        TEST_OUTPUT="Error: Failed to apply mutation"
        echo -n "MUTATION|"
        jq -nc --arg name "mutation#$CURRENT" \
               --arg result "failed" \
               --arg path "$path" \
               --arg search "$search" \
               --arg replace "$replace" \
               --arg output "$TEST_OUTPUT" \
               '{"test_name": $name, "result": $result, "path": $path, "search": $search, "replace": $replace, "output": $output}'
        restore_backup "$path"
        CURRENT_MUTATED_FILE=""
        FAILED=$((FAILED + 1))
        exit 12
    fi

    eval "($TEST_CMD)" > /dev/null 2>&1 || TEST_EXIT_CODE=$?
    TEST_EXIT_CODE=${TEST_EXIT_CODE:-0}
    TEST_OUTPUT="Test exit code: $TEST_EXIT_CODE"
    
    if [ $TEST_EXIT_CODE -eq 0 ]; then
        # Tests passed - mutation survived (FAILED)
        echo -n "MUTATION|"
        jq -nc --arg name "mutation#$CURRENT" \
               --arg result "failed" \
               --arg path "$path" \
               --arg search "$search" \
               --arg replace "$replace" \
               --arg output "$TEST_OUTPUT" \
               '{"test_name": $name, "result": $result, "path": $path, "search": $search, "replace": $replace, "output": $output}'
        restore_backup "$path"
        CURRENT_MUTATED_FILE=""
        FAILED=$((FAILED + 1))
    else
        # Tests failed - mutation killed (PASSED)
        echo -n "MUTATION|"
        jq -nc --arg name "mutation#$CURRENT" \
               --arg result "passed" \
               --arg path "$path" \
               --arg search "$search" \
               --arg replace "$replace" \
               --arg output "$TEST_OUTPUT" \
               '{"test_name": $name, "result": $result, "path": $path, "search": $search, "replace": $replace, "output": $output}'
        restore_backup "$path"
        CURRENT_MUTATED_FILE=""
        PASSED=$((PASSED + 1))
    fi
done < <(echo "$MUTATIONS" | jq -c '.[]')

echo "==================================================================="
echo "Mutation Testing Complete"
echo "==================================================================="
echo -e "Total mutations:  $TOTAL_MUTATIONS"
echo -e "${GREEN}Killed:           $PASSED${NC}"
echo -e "${RED}Survived/Errors:  $FAILED${NC}"
echo ""

if [ $FAILED -gt 0 ]; then
    echo -e "${RED}Some mutations survived or encountered errors!${NC}"
    exit 1
else
    echo -e "${GREEN}All mutations were killed - excellent test coverage!${NC}"
    exit 0
fi
