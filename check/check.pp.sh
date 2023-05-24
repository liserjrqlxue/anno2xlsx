#!/bin/bash
set -e
set -x

umask 0077

cd $(dirname $(readlink -f $0))

tag=PrePregnancy

input=../db/$tag.json.aes.mut.tsv 

mkdir -p $tag

awk 'NR>1' $input | sort > $tag/input.sort

perl ../src/bgi_anno/bin/bgicg_anno.pl ../src/bgi_anno/etc/config_WES.Roche.pl -t tsv -n 13 -b 1000 -q -o $tag/output.out $tag/input.sort

../anno2xlsx -snv $tag/output.out -pp -allTier1

xlsx2txt -input $tag/output.out.Tier1.xlsx 

cut -f 12  $input | sort |uniq -c | sort -n
cut -f 163 $tag/output.out.Tier1.xlsx.filter_variants.txt | sort |uniq -c | sort -n
