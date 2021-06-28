#!/usr/bin/env bash
set -e
shopt -s expand_aliases
alias time='date; time'

scriptdir=$(cd $(dirname $0); pwd -P)
sourcedir=$(cd $scriptdir/..; pwd -P)

time go test -v ./vcs -run=TestVCSPruned
time go test -v ./vcs -run=TestVCS
