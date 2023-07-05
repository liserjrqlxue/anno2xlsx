#!/bin/bash
set -e
set -x

umask 0077

cd $(dirname $(readlink -f $0))

tag=ACMGSF
key1=10
key2=74

sh check.sh $tag $key1 $key2
