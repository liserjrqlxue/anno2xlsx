#!/usr/bin/env bash
gitHash=$(git rev-parse HEAD)
gitBranch=$(git branch --show-current)
gitTag=$(git describe --tags)
goVersion=$(go version)
buildStamp=$(date -u '+%Y-%m-%d_%I:%M:%S%p')
go build -ldflags "-X 'main.buildStamp=$buildStamp' -X 'main.gitHash=$gitBranch:$gitTag:$gitHash' -X 'main.goVersion=$goVersion'"
