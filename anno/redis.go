package anno

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
	"github.com/liserjrqlxue/simple-util"
)

var (
	isIndel = regexp.MustCompile(`ins|del`)
)

func nm2Ensp(item map[string]string, db *redis.Client) error {
	nm := item["Transcript"]
	var v = db.HGet("nm2ensp", nm)
	ensp, err := v.Result()
	if err == nil || err.Error() == "redis: nil" {
		item["ENSP"] = ensp
		return nil
	}
	return err
}

func getNativeSnpField(item map[string]string) string {
	return strings.Join(
		[]string{
			item["#Chr"],
			item["Stop"],
			item["Ref"],
			item["Call"],
		},
		"_",
	)
}

func getNativeIndelField(item map[string]string) string {
	zygo := strings.Split(item["Zygosity"], ";")[0]
	if zygo == "Hemi" {
		zygo = "Hom"
	}
	start, err := strconv.Atoi(item["Start"])
	simple_util.CheckErr(err, "Start")
	stop, err := strconv.Atoi(item["Stop"])
	simple_util.CheckErr(err, "Stop")
	if item["VarType"] == "ins" {
		return strings.Join(
			[]string{
				item["#Chr"],
				item["Start"] + ".." + strconv.Itoa(stop+1),
				"ins" + item["Call"],
				zygo,
			},
			"_",
		)
	} else if item["VarType"] == "del" {
		if start+1 == stop {
			return strings.Join(
				[]string{
					item["#Chr"],
					item["Stop"],
					"del" + item["Ref"],
					zygo,
				},
				"_",
			)
		}
		return strings.Join(
			[]string{
				item["#Chr"],
				strconv.Itoa(start+1) + ".." + item["Stop"],
				"del" + item["Ref"],
				zygo,
			},
			"_",
		)
	}
	/*
		fmt.Printf("[%s] can not get native indel field\n",item["MutationName"])
		for _,key:=range []string{"VarType"}{
			fmt.Printf("\t%s[%s]\n",key,item[key])
		}
	*/
	return ""
}

func redisNativeSnpAF(item map[string]string, db *redis.Client, key, field string) error {
	item["mut"] = field
	var v = db.HGet(key, field)
	r, err := v.Result()
	if err == nil {
		var rs []string
		err = json.Unmarshal([]byte(r), &rs)
		if err == nil {
			item["frequency"] = rs[1]
			item["sampleMut"] = rs[7]
			item["sampleAll"] = rs[8]
			item["sampleInformation"] = rs[9]
		}
	}
	return err
}

func redisNativeIndelAF(item map[string]string, db *redis.Client, key, field string) error {
	item["mut"] = field
	var v = db.HGet(key, field)
	r, err := v.Result()
	if err == nil {
		var rs []string
		err = json.Unmarshal([]byte(r), &rs)
		if err == nil {
			item["frequency"] = rs[4]
			item["sampleMut"] = rs[9]
			item["sampleAll"] = rs[10]
			item["sampleInformation"] = rs[11]
		}
	}
	return err
}

// UpdateRedis update item from redis
func UpdateRedis(item map[string]string, db *redis.Client, keyPrefix string) {
	nm2Ensp(item, db)

	if item["VarType"] == "snv" {
		key := keyPrefix + "_all_native_snp"
		field := getNativeSnpField(item)
		redisNativeSnpAF(item, db, key, field)
	} else if isIndel.MatchString(item["VarType"]) {
		key := keyPrefix + "_all_native_indel"
		field := getNativeIndelField(item)
		redisNativeIndelAF(item, db, key, field)
	}
}
