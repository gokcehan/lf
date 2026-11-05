package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func benchmarkSort(b *testing.B, n int) {
	tmp := b.TempDir()
	for i := 0; i < n; i++ {
		var name string
		switch i % 4 {
		case 0:
			name = fmt.Sprintf("IMG_%06d.JPG", i)
		case 1:
			name = fmt.Sprintf("report-%d-final.pdf", i)
		case 2:
			name = fmt.Sprintf("Été %d.txt", i)
		default:
			name = fmt.Sprintf(".cache-%d", i)
		}
		if err := os.WriteFile(filepath.Join(tmp, name), nil, 0o644); err != nil {
			b.Fatal(err)
		}
	}
	dir := newDir(tmp)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dir.sort()
	}
}

func BenchmarkSort1k(b *testing.B)   { benchmarkSort(b, 1000) }
func BenchmarkSort10k(b *testing.B)  { benchmarkSort(b, 10000) }
func BenchmarkSort100k(b *testing.B) { benchmarkSort(b, 100000) }
