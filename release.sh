#!/bin/bash -xe

# usage: ./release.sh linux x64
# usage: ./release.sh win x64

VER=$(cat constants/constants.go | grep VERSION | awk '{print $3}' | tr -d '"')
OS="${1:-linux}"
ARCH="${2:-x64}"

echo "${VER}"

go get -v
go build -v

mkdir -p releases

cp finance-planner "releases/finance-planner_${VER}_${OS}_${ARCH}"
