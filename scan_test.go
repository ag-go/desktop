package desktop

import (
	"testing"
)

func TestScan(t *testing.T) {
	dirs := DataDirs()

	_, err := Scan(dirs)
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkScan(b *testing.B) {
	var (
		dirs = DataDirs()
		err  error
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = Scan(dirs)
		if err != nil {
			b.Fatal(err)
		}
	}
}
