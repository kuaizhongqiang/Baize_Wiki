package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kuaizhongqiang/baize-wiki/internal/app"
)

func BenchmarkFullPipeline(b *testing.B) {
	dir := b.TempDir()

	// Create 100 small .md files
	for i := 0; i < 100; i++ {
		content := []byte(fmt.Sprintf("---\ntitle: Page %d\n---\n\n# Page %d\n\nContent %d", i, i, i))
		os.WriteFile(filepath.Join(dir, fmt.Sprintf("doc%d.md", i)), content, 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outDir := b.TempDir()
		result := app.RunBuild(context.Background(), dir, outDir, "", 1, false, true, false)
		if !result.Success {
			b.Fatalf("build failed: %v", result.Errors)
		}
	}
}

func BenchmarkLargeMerge(b *testing.B) {
	dir := b.TempDir()

	// Create 50 pages in the same category to trigger merge/split
	for i := 0; i < 50; i++ {
		content := []byte(fmt.Sprintf("---\ntitle: Doc %d\ncategory: docs\n---\n\n# Doc %d\n\n%s", i, i, string(make([]byte, 2000))))
		os.WriteFile(filepath.Join(dir, fmt.Sprintf("doc%d.md", i)), content, 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outDir := b.TempDir()
		result := app.RunBuild(context.Background(), dir, outDir, "", 1, false, true, false)
		if !result.Success {
			b.Fatalf("build failed: %v", result.Errors)
		}
	}
}

func BenchmarkDeepTree(b *testing.B) {
	dir := b.TempDir()

	// Create 200 files distributed across 20+ directories
	for i := 0; i < 20; i++ {
		subDir := filepath.Join(dir, fmt.Sprintf("dir%d", i), "sub")
		os.MkdirAll(subDir, 0755)
		for j := 0; j < 10; j++ {
			content := []byte(fmt.Sprintf("# Page %d-%d", i, j))
			os.WriteFile(filepath.Join(subDir, fmt.Sprintf("p%d.md", j)), content, 0644)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outDir := b.TempDir()
		result := app.RunBuild(context.Background(), dir, outDir, "", 3, false, true, false)
		if !result.Success {
			b.Fatalf("build failed: %v", result.Errors)
		}
	}
}
