package gsqli

import (
	"testing"
)

func Test_SqlCheck(t *testing.T) {
	err := SQLInject("asdf asd ; -1' and 1=1 union/* foo */select load_file('/etc/passwd')--")
	if err == nil {
		t.Fail()
		return
	}
}

func BenchmarkSQLInject(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SQLInject("asdf asd ; -1' and 1=1 union/* foo */select load_file('/etc/passwd')--")
	}
}

func BenchmarkSQLInjectParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			SQLInject("asdf asd ; -1' and 1=1 union/* foo */select load_file('/etc/passwd')--")
		}
	})
}
