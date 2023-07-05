#!/bin/bash
set -e
set -x

xlsx=$1
sheetName=$2
key=$3
rowCount=$4
keyCount=$5
prefix=$6

date

wget -N https://ftp.ebi.ac.uk/pub/databases/genenames/hgnc/tsv/non_alt_loci_set.txt

stat non_alt_loci_set.txt

util/buildDb/buildDb \
  -prefix $prefix \
  -key "$key" -rowCount "$rowCount" -keyCount "$keyCount" \
  -input "$xlsx" -sheet "$sheetName"
