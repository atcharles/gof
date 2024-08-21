#!/usr/bin/env bash
#获取git地址
remote_url=$(git config --local --get remote.origin.url)

# 强制获取远端master分支
function git_pull() {
    set -vue
    rm -rf ./.git && git init
    git remote add origin "$remote_url"
    git checkout --orphan latest_branch
    git add -A
    git commit -am "Init commit" && git fetch
    git checkout -b master origin/master
    git branch -D latest_branch
    #git config pull.rebase false
    #git config --global pull.rebase false
    git pull
}

function git_clean() {
    set -vue
    rm -rf ./.git
    git init
    git checkout --orphan master
    git add -A .
    git commit -am "Initialization commit"
    git remote add origin "$remote_url"
}

# 重置git,强制推送master
function git_reset() {
    set -vue
    rm -rf ./.git
    git init
    git checkout --orphan master
    git add -A .
    git commit -am "Initialization commit"
    git remote add origin "$remote_url"
    git push -f origin master
    git branch --set-upstream-to=origin/master master
    git pull
}

function run() {
    case $1 in
    push) git_reset ;;
    pull) git_pull ;;
    clean) git_clean ;;
    esac
}

run "$@"
