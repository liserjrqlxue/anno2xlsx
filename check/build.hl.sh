#!/bin/bash
set -e
set -x
set -o pipefail

umask 0077

input=$(readlink -f $1) # viphl位点唯一结果统计_total_20230509.xlsx

cd $(dirname $(readlink -f $0))

tag=VIPHL
sheet=Result
ref=/zfsyt1/B2C_RD_P2/USER/wangyaoshen/wes/wes-annotation/src/db/homo_sapiens_refseq/104_GRCh37/Homo_sapiens.GRCh37.75.dna.primary_assembly.fa.gz


chmod 700 $input

mkdir -p $tag

xlsx2txt -input $input -prefix $tag/input

# zcat ~/wes/wes/wes-annotation/src/db/clinvar/vcf_GRCh37/clinvar.vcf.gz|head -30 > vcf.header.txt

head -n1 $tag/input.$sheet.txt > $tag/input.txt
awk 'NR>1' $tag/input.$sheet.txt | sort -V >> $tag/input.txt

perl txt2vcf.pl $tag/input.txt  | bgzip > $tag/input.vcf.gz

tabix -f -p vcf $tag/input.vcf.gz

bcftools norm -c w --force -m- -f $ref -o $tag/input.norm.vcf.gz -O z $tag/input.vcf.gz

perl ../src/bgi_anno/bin/bgicg_anno.pl ../src/bgi_anno/etc/config_WES.Roche.pl -b 1000 -n 10 -o $tag/output.txt $tag/input.norm.vcf.gz

cut -f 1,2,3,7,9,22,28,30,74,75,76 $tag/output.txt | awk '$7!="."&&$8!="."' > $tag/output.lite.txt
txt2excel -input $tag/output.lite.txt -sheet $sheet -output $tag/$tag
