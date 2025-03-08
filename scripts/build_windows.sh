#!/usr/bin/env bash

set -e

workdir=/tmp/work_$(date "+%Y%m%d%H%M%S")
htmldir=$workdir/html
dist=./dist/windows

mkdir -p $htmldir
mkdir -p $dist

./scripts/build.sh
cp ./build/* $htmldir

GOOS=windows GOARCH=amd64 go build -o $workdir/colortress_server ./localserver/for_dist/run_server.go

tar -czvf $dist/colortress_windows.tar.gz -C $workdir/ .

rm -rf $workdir
