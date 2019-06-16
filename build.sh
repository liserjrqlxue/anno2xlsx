#!/usr/bin/env bash
go build -ldflags "-X main.buildStamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.gitHash=`git rev-parse HEAD` -X 'main.goVersion=`go version`'"
