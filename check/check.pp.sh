#!/bin/bash
set -e
set -x

umask 0077

cd $(dirname $(readlink -f $0))

tag=PrePregnancy
key1=12
key2=163

sh check.sh $tag $key1 $key2
