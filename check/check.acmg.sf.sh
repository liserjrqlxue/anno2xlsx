#!/bin/bash
set -e
set -x

umask 0077

cd $(dirname $(readlink -f $0))

input=../db/ACMGSF.json.aes.mut.tsv 


awk 'NR>1' $input | sort > input.sort

perl ../src/bgi_anno/bin/bgicg_anno.pl ../src/bgi_anno/etc/config_WES.Roche.pl -t tsv -n 13 -b 1000 -q -o output.out input.sort

../anno2xlsx -snv output.out -sf -allTier1

xlsx2txt -input output.out.Tier1.xlsx 

cut -f 10 $input | sort |uniq -c | sort -n
cut -f 74 output.out.Tier1.xlsx.filter_variants.txt|sort |uniq -c | sort -n
