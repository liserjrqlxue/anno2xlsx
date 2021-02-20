package hgvs

import "fmt"

var MT = "NC_012920.1"

func GetMhgvs(pos int, ref, alt []byte) (mHGVS string) {
	if string(ref) == string(alt) {
		return fmt.Sprintf("%s:m.%d=", MT, pos)
	}
	for len(ref) > 0 && len(alt) > 0 && ref[0] == alt[0] {
		ref = ref[1:]
		alt = alt[1:]
		pos++
	}
	mHGVS = fmt.Sprintf("%s:m.%d%s>%s", MT, pos, ref, alt)
	if len(ref) == 0 || string(ref) == "" {
		mHGVS = fmt.Sprintf("%s:m.%d_%dins%s", MT, pos-1, pos, alt)
	} else if len(alt) == 0 || string(ref) == "" {
		if len(ref) == 1 {
			mHGVS = fmt.Sprintf("%s:m.%ddel", MT, pos)
		} else {
			mHGVS = fmt.Sprintf("%s:m.%d_%ddel", MT, pos, pos+len(ref)-1)
		}
	}
	return
}
