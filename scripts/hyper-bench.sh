#!/usr/bin/env bash
set -e
shopt -s expand_aliases
alias time='date; time'

scriptdir=$(cd $(dirname $0); pwd -P)
sourcedir=$(cd $scriptdir/..; pwd -P)

# Micro benchmarks without OpenAll and Commit
filepath="$sourcedir/json-parse/micro-macro-1024txn.json"
time go test -v ./vcs -bench=BenchmarkPrunedVCS -run=BenchmarkPrunedVCS -benchtime= 1x -benchmem -timeout 10800m -json | tee $filepath

# Benchmarks of Hyperproofs aggregation
filepath="$sourcedir/json-parse/hyper-agg.json"
time go test -v ./vcs -bench=BenchmarkVCSAgg -run=BenchmarkVCSAgg -benchtime 2x -benchmem -timeout 360m -json | tee $filepath
outpath="$sourcedir/plots/hyperproofs-agg.csv"
time python3 "$sourcedir/json-parse/parse-agg.py" $filepath $outpath

# This computes the estimate verification time of SNARK based Merkle aggregation.
go build && time ./hyperproofs-go 1

# # WARNING: Benchmarking OpenAll and Commit takes around 6.5 hours.
# go build && time ./hyperproofs-go 2 # This benchmarks OpenAll and Commit
