#!/usr/bin/env bash
set -x

upstream_url=https://github.com/devopsfaith/krakend.git
branch_name="sync-upstream-"$(date +"%Y%M%d")

function add_remote {
	echo "* Adding the upstream remote: $upstream_url"
	git remote add upstream $origin_url
}

function main {
	upstream=$(git remote | grep upstream )
	if [[ -z $upstream ]]; then
		add_remote
	fi

	git checkout t-master

	echo "* Fetching from upstream"
	git pull --tags upstream master

	echo "* Creating new branch $branch_name"
	git checkout -b $branch_name

	echo "* Pushing to internal github"
	git push --tags origin $branch_name
}

main


