#!/usr/bin/env bash
set -e

branch_name="sync-upstream-"$(date +"%Y%M%d")

function main {

	git checkout t-master

	echo "* Fetching from upstream"
	git pull --tags upstream master

	echo "* Creating new branch $branch_name"
	git checkout -b $branch_name

	echo "* Pushing to internal github"
	git push --tags origin $branch_name
}

main


