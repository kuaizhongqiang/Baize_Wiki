package parser

import (
	"testing"
)

func BenchmarkParseMarkdown(b *testing.B) {
	content := "# Title\n\nSome intro.\n\n## Section 1\n\nContent A.\n\n### Sub 1.1\n\nDeep content.\n\n## Section 2\n\nContent B.\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseMarkdown(content)
	}
}

func BenchmarkFrontmatterExtraction(b *testing.B) {
	content := "---\ntitle: Benchmark Page\ndescription: A test page for benchmarking\ntags: [go, test, benchmark]\nweight: 1\ncategory: test\n---\n\n# Benchmark\n\nContent here\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractFrontmatter(content)
	}
}
