#!/bin/bash
set -e
set -x

tag=$1
input=$2

chmod 700 $input

mkdir -p $tag

awk 'NR>1' $input | sort > $tag/input.sort

perl ../src/bgi_anno/bin/bgicg_anno.pl ../src/bgi_anno/etc/config_WES.Roche.pl -t tsv -n 13 -b 1000 -q -o $tag/output.out $tag/input.sort

../anno2xlsx -snv $tag/output.out -sf -pp -nb -hl -allTier1

xlsx2txt -input $tag/output.out.Tier1.xlsx 
