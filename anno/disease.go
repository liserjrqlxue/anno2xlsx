package anno

import (
	"encoding/json"
	"strings"

	"github.com/liserjrqlxue/goUtil/simpleUtil"
)

func UpdateDiseaseMultiGene(sep string, genes []string, item, geneDiseaseDbColumn map[string]string, geneDiseaseDb map[string]map[string]string) {
	// 基因-疾病
	for key, value := range geneDiseaseDbColumn {
		var vals []string
		for _, gene := range genes {
			geneDb, ok := geneDiseaseDb[gene]
			if ok {
				vals = append(vals, geneDb[key])
			}
		}
		if len(vals) > 0 {
			item[value] = strings.Join(vals, sep)
		}
	}
}

type GeneInfo struct {
	OmimGene     string        `json:"omimGene"`
	Transcript   string        `json:"transcript"`
	Exon         string        `json:"exon"`
	EffectType   string        `json:"effecttype"`
	Chr          string        `json:"chr"`
	Start        string        `json:"start"`
	Stop         string        `json:"stop"`
	Primer       string        `json:"primer"`
	OmimGeneId   string        `json:"omimGeneId"`
	Location     string        `json:"location"`
	OmimDiseases []OmimDisease `json:"omimDiseases"`
}
type OmimDisease struct {
	DiseaseCnName    string `json:"diseaseCnName"`
	DiseaseEnName    string `json:"diseaseEnName"`
	GeneralizationCn string `json:"generalizationCn"`
	GeneralizationEn string `json:"generalizationEn"`
	OmimDiseaseId    string `json:"omimDiseaseId"`
	OmimGeneId       string `json:"omimGeneId"`
	SystemSort       string `json:"systemSort"`
	HeredityModel    string `json:"heredityModel"`
}

//UpdateDiseMultiGene anno disease info for multil gene
func UpdateDiseMultiGene(geneLst string, item, geneDisDbCol map[string]string, geneDisDb map[string]map[string]string) {
	genes := strings.Split(geneLst, ";")
	var geneLocus []string
	// 基因-疾病
	for key, value := range geneDisDbCol {
		var vals []string
		for _, gene := range genes {
			singelGeneDb, ok := geneDisDb[gene]
			if ok {
				vals = append(vals, singelGeneDb[key])
				geneLocus = append(geneLocus, singelGeneDb["Gene/Locus"])
			}
		}
		if len(vals) > 0 {
			item[value] = strings.Join(vals, "\n")
			item["Gene/Locus"] = strings.Join(geneLocus, "\n")
		}
	}

}

func UpdateCnvAnnot(geneLst string, item map[string]string, geneDisDb map[string]map[string]string) {
	genes := strings.Split(geneLst, ";")
	var exonMap = getExonMap(item)
	var cnvAnnots []GeneInfo
	for _, gene := range genes {
		singleGeneDb, ok := geneDisDb[gene]
		if !ok {
			continue
		}
		var cnvAnnot = GeneInfo{
			OmimGene:     gene,
			Transcript:   item["Transcript"],
			Exon:         exonMap[gene],
			EffectType:   item["type"],
			Chr:          item["chromosome"],
			Start:        item["start"],
			Stop:         item["end"],
			Primer:       item["Primer"],
			OmimGeneId:   strings.Split(singleGeneDb["Gene/Locus MIM number"], "\n")[0],
			Location:     strings.Split(singleGeneDb["Location"], "\n")[0],
			OmimDiseases: nil,
		}
		cnvAnnot.OmimDiseases = singelGeneDb2OmimDiseases(singleGeneDb)
		cnvAnnots = append(cnvAnnots, cnvAnnot)
	}

	var jsonBytes, e = json.MarshalIndent(cnvAnnots, "", "  ")
	simpleUtil.CheckErr(e)
	item["CNV_annot"] = string(jsonBytes)
}

func getExonMap(item map[string]string) map[string]string {
	var exonMap = make(map[string]string)
	var genes = strings.Split(item["OMIM_Gene"], ";")
	var exons = strings.Split(item["OMIM_exon"], ";")
	for i, gene := range genes {
		var exon, ok = exonMap[gene]
		if ok {
			exon = exon + "," + exons[i]
		}
		exonMap[gene] = exon
	}
	return exonMap
}

func singelGeneDb2OmimDiseases(item map[string]string) (omimDiseases []OmimDisease) {
	var DiseaseCnName = strings.Split(item["Disease NameCH"], "\n")
	var DiseaseEnName = strings.Split(item["Disease NameEN"], "\n")
	var GeneralizationCn = strings.Split(item["GeneralizationCH"], "\n")
	var GeneralizationEn = strings.Split(item["GeneralizationEN"], "\n")
	var OmimDiseaseID = strings.Split(item["Phenotype MIM number"], "\n")
	var OmimGeneID = strings.Split(item["Gene/Locus MIM number"], "\n")
	var SystemSort = strings.Split(item["SystemSort"], "\n")
	var HeredityModel = strings.Split(item["Inheritance"], "\n")
	for i := 0; i < len(DiseaseCnName); i++ {
		var omimDisease = OmimDisease{
			DiseaseCnName:    DiseaseCnName[i],
			DiseaseEnName:    DiseaseEnName[i],
			GeneralizationCn: GeneralizationCn[i],
			GeneralizationEn: GeneralizationEn[i],
			OmimDiseaseId:    OmimDiseaseID[i],
			OmimGeneId:       OmimGeneID[i],
			SystemSort:       SystemSort[i],
			HeredityModel:    HeredityModel[i],
		}
		omimDiseases = append(omimDiseases, omimDisease)
	}
	return
}
