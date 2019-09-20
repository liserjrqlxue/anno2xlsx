package anno

import "strings"

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
