#!/bin/bash
set -e
set -x

umask 0077

tag=$1
input=$2
output=$3
key1=$4
key2=$5

cut -f $key1 $input  | sort | uniq -c | sort -n
cut -f $key2 $output | sort |uniq -c | sort -n

cut -f 1-5,$key1 $input  | sort | uniq |sed 's|^chr||' > $tag/input.1
cut -f 2-6,$key2 $output | sort | uniq |sed 's|^chr||' > $tag/output.1

comm -12 $tag/input.1 $tag/output.1 > $tag/diff.3
comm -23 $tag/input.1 $tag/output.1 > $tag/diff.1
comm -13 $tag/input.1 $tag/output.1 > $tag/diff.2

wc -l $input $output $tag/input.1 $tag/output.1 $tag/diff.*

head $tag/diff.1 $tag/diff.2
