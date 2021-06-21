#!/usr/bin/env bash
#获取git地址
remote_url=$(git config --local --get remote.origin.url)

# 强制获取远端master分支
function git_pull() {
	rm -rf ./.git && git init &&
		git remote add origin "$remote_url" &&
		git checkout --orphan latest_branch &&
		git add -A &&
		git commit -am "Init commit" && git fetch &&
		git checkout -b master origin/master &&
		git branch -D latest_branch &&
		git config pull.rebase false
	git config --global pull.rebase false
	git pull
}

function git_clean() {
	git checkout --orphan latest_branch &&
		git add -A &&
		git commit -am "Init commit" &&
		git branch -D master &&
		git branch -m master &&
		git push -f origin master &&
		git branch --set-upstream-to=origin/master master &&
		git config pull.rebase false
	git config --global pull.rebase false
	git pull
}

# 重置git,强制推送master
function git_reset() {
	rm -rf ./.git &&
		git init &&
		git remote add origin "$remote_url" &&
		git checkout --orphan master &&
		git add -A . &&
		git commit -am "Initialization commit" &&
		git push -f origin master &&
		git branch --set-upstream-to=origin/master master &&
		git config pull.rebase false
	git config --global pull.rebase false
	git pull
}

function run() {
	case $1 in
	git_reset) git_reset ;;
	git_pull) git_pull ;;
	esac
}

run "$@"
