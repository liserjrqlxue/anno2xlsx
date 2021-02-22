package anno

import "strings"

// UpdateGeneDb annotate 突变频谱
func UpdateGeneDb(geneList string, item, geneDb map[string]string) {
	genes := strings.Split(geneList, ";")
	// 突变频谱
	var vals []string
	for _, gene := range genes {
		vals = append(vals, geneDb[gene])
	}
	item["突变频谱"] = strings.Join(vals, "\n")
}
