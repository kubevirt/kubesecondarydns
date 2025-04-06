#!/usr/bin/env bash

set -e

if [[ -n "$(git status --porcelain)" ]] ; then
    echo "You have Uncommitted changes. Please commit the changes"
    git status --porcelain
    git --no-pager diff
    exit 1
fi
