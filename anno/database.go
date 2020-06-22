package anno

import "github.com/liserjrqlxue/goUtil/textUtil"

var (
	Gene2trans = make(map[string]string)
	Trans2gene = make(map[string]string)
)

// LoadGeneTrans read geneSymbol.transcript.txt to two map
func LoadGeneTrans(fileName string) {
	for _, array := range textUtil.File2Slice(fileName, "\t") {
		var gene = array[0]
		var trans = array[1]
		Gene2trans[gene] = trans
		Trans2gene[trans] = gene
	}
}
