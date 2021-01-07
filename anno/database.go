package anno

import "github.com/liserjrqlxue/goUtil/textUtil"

var (
	gene2trans = make(map[string]string)
	trans2gene = make(map[string]string)
)

// LoadGeneTrans read geneSymbol.transcript.txt to two map
func LoadGeneTrans(fileName string) {
	for _, array := range textUtil.File2Slice(fileName, "\t") {
		var gene = array[0]
		var trans = array[1]
		gene2trans[gene] = trans
		trans2gene[trans] = gene
	}
}
