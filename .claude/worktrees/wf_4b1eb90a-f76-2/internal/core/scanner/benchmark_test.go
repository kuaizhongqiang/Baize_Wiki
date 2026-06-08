package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkScan(b *testing.B) {
	dir := b.TempDir()

	// Create test files
	for i := 0; i < 100; i++ {
		content := []byte("# File " + string(rune('A'+i%26)) + "\n\nContent\n")
		os.WriteFile(filepath.Join(dir, "file" + string(rune('0'+i%10)) + ".md"), content, 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Scan(context.Background(), dir, ScanConfig{})
	}
}

func BenchmarkBinaryDetection(b *testing.B) {
	data := []byte("The quick brown fox jumps over the lazy dog. This is a test of binary detection.")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isBinaryData(data)
	}
}
