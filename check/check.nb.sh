#!/bin/bash
set -e
set -x

umask 0077

cd $(dirname $(readlink -f $0))

tag=NBSP
key1=9
key2=166

sh check.sh $tag $key1 $key2
