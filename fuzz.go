//go:build gofuzz
// +build gofuzz

package golibinjection

func Fuzz(data []byte) int {
	//XSSParser(string(data))
	SQLInject(string(data))
	return 0
}
