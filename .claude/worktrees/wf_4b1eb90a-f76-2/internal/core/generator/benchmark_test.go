package generator

import (
	"fmt"
	"testing"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
)

func BenchmarkLevelBuilderFlat(b *testing.B) {
	pages := make([]*model.Page, 100)
	for i := 0; i < 100; i++ {
		w := i % 10
		pages[i] = makePage(
			fmt.Sprintf("category%d/file%d.md", i/10, i),
			"Page",
			fmt.Sprintf("category%d", i/10),
			w,
		)
	}

	builder := NewLevelBuilder()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.Flat(pages)
	}
}

func BenchmarkLevelBuilderStructured(b *testing.B) {
	pages := make([]*model.Page, 100)
	for i := 0; i < 100; i++ {
		pages[i] = makePage(
			fmt.Sprintf("category%d/file%d.md", i/10, i),
			"Page",
			"",
			0,
		)
	}

	builder := NewLevelBuilder()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.Structured(pages)
	}
}
