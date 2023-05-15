package anno

import (
	"regexp"
	"strings"
)

var (
	cHGVSAlt = regexp.MustCompile(`alt: (\S+) \)`)
	cHGVSStd = regexp.MustCompile(`std: (\S+) `)
	ins      = regexp.MustCompile(`ins`)
	del      = regexp.MustCompile(`del([ACGT]+)`)
	dup      = regexp.MustCompile(`dup([ACGT]+)`)
)

func cHgvsAlt(cHgvs string) string {
	if m := cHGVSAlt.FindStringSubmatch(cHgvs); m != nil {
		return m[1]
	}
	return cHgvs
}

func cHgvsStd(cHgvs string) string {
	if m := cHGVSStd.FindStringSubmatch(cHgvs); m != nil {
		return m[1]
	}
	return cHgvs
}

func hgvsDelDup(cHgvs string) string {
	if ins.MatchString(cHgvs) {
		return cHgvs
	}
	if m := del.FindStringSubmatch(cHgvs); m != nil {
		cHgvs = strings.TrimSuffix(cHgvs, m[1])
	}
	if m := dup.FindStringSubmatch(cHgvs); m != nil {
		cHgvs = strings.TrimSuffix(cHgvs, m[1])
	}
	return cHgvs
}

// GetKeys get keys from transcript and cHGVS
func GetKeys(transcript, cHGVS string) []string {
	var cAlt = cHgvsAlt(cHGVS)
	var cStd = cHgvsStd(cHGVS)
	var cAlt1 = hgvsDelDup(cAlt)
	var cStd1 = hgvsDelDup(cStd)
	var key1 = transcript + ":" + cHGVS
	var key2 = transcript + ":" + cAlt
	var key3 = transcript + ":" + cStd
	var key4 = transcript + ":" + cAlt1
	var key5 = transcript + ":" + cStd1
	return []string{key1, key2, key3, key4, key5}
}

// GetFromMultiKeys loop keys, return info,hit
func GetFromMultiKeys(db map[string]map[string]string, keys []string) (info map[string]string, hit bool) {
	for _, key := range keys {
		info, hit = db[key]
		if hit {
			return info, hit
		}
	}
	return info, hit
}
