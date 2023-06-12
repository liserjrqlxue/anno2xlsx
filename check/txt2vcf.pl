#!/usr/bin/env perl
#
use strict;
use warnings;

open HD,"< vcf.header.txt" or die$!;
while(<HD>){
	print;
}
close HD;

my%long2short=(
	"Benign"                 => "B",
	"Likely_Benign"          => "LB",
	"Likely Benign"          => "LB",
	"Uncertain Significance" => "VUS",
	"Likely Pathogenic"      => "LP",
	"Pathogenic"             => "P",
);

chomp(my$header=<>);
my@header=split /\t/,$header;
while(<>){
	chomp;
	my%info;
	@info{@header}=split /\t/,$_;
	my@mut=split /-/,$info{variantId};
	$info{criteria}=~s/\s+//g;
	print join("\t",$mut[0],$mut[1],".",$mut[2],$mut[3],".",".","HLID=$info{variantId};HLintepretation=$long2short{$info{intepretation}};HLcriteria=$info{criteria}"),"\n";
}

__END__
#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO
1	861332	1019397	G	A	.	.	ALLELEID=1003021;CLNDISDB=MedGen:CN517202;CLNDN=not_provided;CLNHGVS=NC_000001.10:g.861332G>A;CLNREVSTAT=criteria_provided,_single_submitter;CLNSIG=Uncertain_significance;CLNVC=single_nucleotide_variant;CLNVCSO=SO:0001483;GENEINFO=SAMD11:148398;MC=SO:0001583|missense_variant;ORIGIN=1;RS=1640863258
1	861336	1543320	C	T	.	.	ALLELEID=1632777;CLNDISDB=MedGen:CN517202;CLNDN=not_provided;CLNHGVS=NC_000001.10:g.861336C>T;CLNREVSTAT=criteria_provided,_single_submitter;CLNSIG=Likely_benign;CLNVC=single_nucleotide_variant;CLNVCSO=SO:0001483;GENEINFO=SAMD11:148398;MC=SO:0001819|synonymous_variant;ORIGIN=1
variantId	gene	chgvs	intepretation	criteria
1-100818534-A-G	CDC14A	NM_003672.4:c.24A>G	Likely Benign	PM2, BP4, BP7
1-100818535-A-G	CDC14A	NM_003672.4:c.25A>G	Uncertain Significance	PM2, BP4
1-100818575-C-T	CDC14A	NM_003672.4:c.49+16C>T	Uncertain Significance	PM2, BP4
1-100818579-C-T	CDC14A	NM_003672.4:c.49+20C>T	Uncertain Significance	PM2, BP4
1-100856267-CTTT-C	CDC14A	NM_003672.4:c.217-18_217-16del	Uncertain Significance	PM2, BP4
1-100856274-T-A	CDC14A	NM_003672.4:c.217-14T>A	Uncertain Significance	PM2, BP4
1-100856325-C-G	CDC14A	NM_003672.4:c.254C>G	Uncertain Significance	PM2
1-100856345-AGAGCAAATGCAGCATTTTT-A	CDC14A	NM_003672.4:c.277_295del	Likely Pathogenic	PVS1, PM2
1-100889825-G-C	CDC14A	NM_003672.4:c.357G>C	Likely Benign	PM2_Supporting, BP4, BP7
##fileformat=VCFv4.1
##fileDate=2022-09-17
##source=ClinVar
##reference=GRCh37
##ID=<Description="ClinVar Variation ID">
##INFO=<ID=AF_ESP,Number=1,Type=Float,Description="allele frequencies from GO-ESP">
##INFO=<ID=AF_EXAC,Number=1,Type=Float,Description="allele frequencies from ExAC">
##INFO=<ID=AF_TGP,Number=1,Type=Float,Description="allele frequencies from TGP">
##INFO=<ID=ALLELEID,Number=1,Type=Integer,Description="the ClinVar Allele ID">
##INFO=<ID=CLNDN,Number=.,Type=String,Description="ClinVar's preferred disease name for the concept specified by disease identifiers in CLNDISDB">
##INFO=<ID=CLNDNINCL,Number=.,Type=String,Description="For included Variant : ClinVar's preferred disease name for the concept specified by disease identifiers in CLNDISDB">
##INFO=<ID=CLNDISDB,Number=.,Type=String,Description="Tag-value pairs of disease database name and identifier, e.g. OMIM:NNNNNN">
##INFO=<ID=CLNDISDBINCL,Number=.,Type=String,Description="For included Variant: Tag-value pairs of disease database name and identifier, e.g. OMIM:NNNNNN">
##INFO=<ID=CLNHGVS,Number=.,Type=String,Description="Top-level (primary assembly, alt, or patch) HGVS expression.">
##INFO=<ID=CLNREVSTAT,Number=.,Type=String,Description="ClinVar review status for the Variation ID">
##INFO=<ID=CLNSIG,Number=.,Type=String,Description="Clinical significance for this single variant; multiple values are separated by a vertical bar">
##INFO=<ID=CLNSIGCONF,Number=.,Type=String,Description="Conflicting clinical significance for this single variant; multiple values are separated by a vertical bar">
##INFO=<ID=CLNSIGINCL,Number=.,Type=String,Description="Clinical significance for a haplotype or genotype that includes this variant. Reported as pairs of VariationID:clinical significance; multiple values are separated by a vertical bar">
##INFO=<ID=CLNVC,Number=1,Type=String,Description="Variant type">
##INFO=<ID=CLNVCSO,Number=1,Type=String,Description="Sequence Ontology id for variant type">
##INFO=<ID=CLNVI,Number=.,Type=String,Description="the variant's clinical sources reported as tag-value pairs of database and variant identifier">
##INFO=<ID=DBVARID,Number=.,Type=String,Description="nsv accessions from dbVar for the variant">
##INFO=<ID=GENEINFO,Number=1,Type=String,Description="Gene(s) for the variant reported as gene symbol:gene id. The gene symbol and id are delimited by a colon (:) and each pair is delimited by a vertical bar (|)">
##INFO=<ID=MC,Number=.,Type=String,Description="comma separated list of molecular consequence in the form of Sequence Ontology ID|molecular_consequence">
##INFO=<ID=ORIGIN,Number=.,Type=String,Description="Allele origin. One or more of the following values may be added: 0 - unknown; 1 - germline; 2 - somatic; 4 - inherited; 8 - paternal; 16 - maternal; 32 - de-novo; 64 - biparental; 128 - uniparental; 256 - not-tested; 512 - tested-inconclusive; 1073741824 - other">
##INFO=<ID=RS,Number=.,Type=String,Description="dbSNP ID (i.e. rs number)">
##INFO=<ID=SSR,Number=1,Type=Integer,Description="Variant Suspect Reason Codes. One or more of the following values may be added: 0 - unspecified, 1 - Paralog, 2 - byEST, 4 - oldAlign, 8 - Para_EST, 16 - 1kg_failed, 1024 - other">
#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO
1	861332	1019397	G	A	.	.	ALLELEID=1003021;CLNDISDB=MedGen:CN517202;CLNDN=not_provided;CLNHGVS=NC_000001.10:g.861332G>A;CLNREVSTAT=criteria_provided,_single_submitter;CLNSIG=Uncertain_significance;CLNVC=single_nucleotide_variant;CLNVCSO=SO:0001483;GENEINFO=SAMD11:148398;MC=SO:0001583|missense_variant;ORIGIN=1;RS=1640863258
1	861336	1543320	C	T	.	.	ALLELEID=1632777;CLNDISDB=MedGen:CN517202;CLNDN=not_provided;CLNHGVS=NC_000001.10:g.861336C>T;CLNREVSTAT=criteria_provided,_single_submitter;CLNSIG=Likely_benign;CLNVC=single_nucleotide_variant;CLNVCSO=SO:0001483;GENEINFO=SAMD11:148398;MC=SO:0001819|synonymous_variant;ORIGIN=1
      1 intepretation
     15 Likely_Benign
   1106 Pathogenic
   1344 Benign
   2461 Likely Pathogenic
   5624 Likely Benign
  20402 Uncertain Significance
