#!/usr/bin/env bash
set -e

origin_url=https://github.com/devopsfaith/krakend.git
branch_name="sync-upstream-"$(date +"%Y%M%d")

function add_remote {
	echo "* Adding the origin remote: $origin_url"
	git remote add origin $origin_url
}

function main {
	origin=$(git remote | grep origin )

	if [[ -z $origin ]]; then
		add_remote
	fi

	git checkout master

	echo "* Fetching from upstream"
	git pull --tags origin master

	echo "* Creating new branch $branch_name"
	git checkout -b $branch_name

	echo "* Pushing to internal github"
	git push --tags schibsted $branch_name
}

main


