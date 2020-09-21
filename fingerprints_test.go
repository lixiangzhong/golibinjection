package gsqli

import (
	"testing"
)

func Test_SearchFile(t *testing.T) {
	err := LoadData(".\\sqlparse_data.json")
	if err != nil {
		t.Error(err)
		return
	}

	keyData := []string{
		"v;T(1",
		"v)&s&",
		"&(1)U",
		"1&1ov",
	}

	for _, v := range keyData {
		pos := Lookup(v)
		if pos == -1 {
			t.Fail()
			return
		}

		if fingerprints.Fingerprints[pos] != v {
			t.Fail()
			return
		}
	}
}

func Benchmark_Search(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pos := Lookup("v;T(1")
		if pos == -1 {
			b.Fail()
			return
		}
	}
}

func Benchmark_SearchParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pos := Lookup("v;T(1")
			if pos == -1 {
				b.Fail()
				return
			}
		}
	})
}
