#!/bin/sh

# This script creates a release tag on main/master branch and pushes both the branch and the tag.
# Tags must have the 'vXYZ' format. The script doesn't work if the repo has uncommitted changes.

set -eu

cd "$(dirname "${0}")/../"

if [ $# -ne 1 ]; then
    echo "Missing tag version. Usage: $0 <tag version>"

    exit 1
fi

if git rev-list "$1" > /dev/null 2>&1; then
    echo "Provided tag already exists. List of existing tags:"
    git tag

    exit 1
fi

if ! git diff-index --quiet HEAD --; then
    echo "There are uncommitted changes; please commit or stash before releasing. Quitting"

    exit 1
fi

BRANCH_NAME=$(git rev-parse --abbrev-ref HEAD)

if [ "$BRANCH_NAME" != "main" ] && [ "$BRANCH_NAME" != "master" ]; then
    echo "Not on main or master branch, quitting."

    exit 1
fi

git tag "$1"
git push origin HEAD --tags
