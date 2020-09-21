package gsqli

import (
	"encoding/json"
	"io/ioutil"
	"sort"
	"strings"
)

func UnmarshalSqlifingerprint(data []byte) (Sqlifingerprint, error) {
	var r Sqlifingerprint
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Sqlifingerprint) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Sqlifingerprint struct {
	Charmap      []string           `json:"charmap"`
	Fingerprints []string           `json:"fingerprints"`
	Keywords     map[string]Keyword `json:"keywords"`
}

type Keyword string

const (
	A        Keyword = "A"
	B        Keyword = "B"
	E        Keyword = "E"
	Empty    Keyword = "&"
	F        Keyword = "f"
	K        Keyword = "k"
	KeywordT Keyword = "T"
	N        Keyword = "n"
	O        Keyword = "o"
	T        Keyword = "t"
	The1     Keyword = "1"
	U        Keyword = "U"
	V        Keyword = "v"
)

type keyword_t struct {
	word  string
	vtype byte
}

var (
	fingerprints Sqlifingerprint
	szKeywordMap map[string]keyword_t = map[string]keyword_t{}
)

func init() {
	for _, vk := range szkeywords {
		szKeywordMap[strings.ToUpper(vk.word)] = vk
	}
}

func initData(body []byte) error {
	sf, err := UnmarshalSqlifingerprint(body)
	if err != nil {
		return err
	}

	fingerprints = sf

	// 做个排序
	sort.Strings(fingerprints.Fingerprints)
	return nil
}

// 在二分查找中，数据必须做足够的排序，且排序必须是升序
func LoadData(filename string) error {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return initData(body)
}

func Lookup(key string) int {
	if len(fingerprints.Fingerprints) == 0 {
		return 0
	}

	// 由于SORT内的SEARCH就是二分查找法，所以不必单独编写
	upKey := strings.ToUpper(key)
	pos := sort.SearchStrings(fingerprints.Fingerprints, upKey)
	if pos == -1 {
		return 0
	}

	if strings.Compare(fingerprints.Fingerprints[pos], upKey) == 0 {
		return pos
	}

	return 0
}
