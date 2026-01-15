#!/bin/bash
set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <tag>"
    echo "Example: $0 v0.1.0"
    exit 1
fi

TAG="$1"

if [[ ! "$TAG" =~ ^v[0-9] ]]; then
    echo "Error: Tag must start with 'v' followed by a number (e.g., v0.1.0)"
    exit 1
fi

echo "Creating tag $TAG..."
git tag "$TAG"

echo "Pushing tag to origin..."
git push origin "$TAG"

echo "Done! The release workflow will now build and publish binaries."
echo "Watch progress at: https://github.com/bsamek/wt/actions"
