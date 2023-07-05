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
func GetKeys(transcript, cHGVS string) (keys []string) {
	var (
		cAlt  = cHgvsAlt(cHGVS)
		cStd  = cHgvsStd(cHGVS)
		cAlt1 = hgvsDelDup(cAlt)
		cStd1 = hgvsDelDup(cStd)
		sep   = []string{":", "\t"}
		hgvs  = []string{cHGVS, cAlt, cStd, cAlt1, cStd1}
	)
	for _, s := range sep {
		for _, h := range hgvs {
			keys = append(keys, transcript+s+h)
		}

	}
	return
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
