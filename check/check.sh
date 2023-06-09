#!/bin/bash
set -e
set -x

umask 0077

cd $(dirname $(readlink -f $0))

tag=$1
key1=$2
key2=$3

input=../db/$tag.json.aes.mut.tsv 
output=$tag/output.out.Tier1.xlsx.filter_variants.txt

sh anno.sh $tag $input

sh compare.sh $tag $input $output $key1 $key2
