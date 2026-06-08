package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractBasicLink(t *testing.T) {
	content := "参考 [[数据模型]] 了解详情"
	refs := ExtractWikiLinks(content)
	assert.Len(t, refs, 1)
	assert.Equal(t, "数据模型", refs[0].Target)
	assert.Empty(t, refs[0].Text)
}

func TestExtractLinkWithText(t *testing.T) {
	content := "参考 [[数据模型|查看详情]]"
	refs := ExtractWikiLinks(content)
	assert.Len(t, refs, 1)
	assert.Equal(t, "数据模型", refs[0].Target)
	assert.Equal(t, "查看详情", refs[0].Text)
}

func TestExtractMultipleLinks(t *testing.T) {
	content := "参见 [[文档1]] 和 [[文档2|说明]]"
	refs := ExtractWikiLinks(content)
	assert.Len(t, refs, 2)
	assert.Equal(t, "文档1", refs[0].Target)
	assert.Equal(t, "文档2", refs[1].Target)
	assert.Equal(t, "说明", refs[1].Text)
}

func TestExtractNoLinks(t *testing.T) {
	content := "普通文本，没有任何链接"
	refs := ExtractWikiLinks(content)
	assert.Empty(t, refs)
}

func TestExtractLinkInCodeBlock(t *testing.T) {
	content := "普通文本 [[有效链接]]\n\n```go\n// 代码块中的 [[无效链接]] 不应被提取\nfunc main() {\n    // [[另一个链接]]\n}\n```\n\n结尾文本"
	refs := ExtractWikiLinks(content)
	assert.Len(t, refs, 1, "should only extract link outside code block")
	assert.Equal(t, "有效链接", refs[0].Target)
}

func TestExtractExternalLinkClassification(t *testing.T) {
	// External links should still be extracted as LinkRefs
	content := "参考 [[https://example.com]]"
	refs := ExtractWikiLinks(content)
	assert.Len(t, refs, 1)
	assert.Equal(t, "https://example.com", refs[0].Target)
}

func TestExtractLinkWithPath(t *testing.T) {
	content := "参见 [[目录/页面名]]"
	refs := ExtractWikiLinks(content)
	assert.Len(t, refs, 1)
	assert.Equal(t, "目录/页面名", refs[0].Target)
}

func TestExtractAnchorLink(t *testing.T) {
	content := "详见 [[#安装步骤]]"
	refs := ExtractWikiLinks(content)
	assert.Len(t, refs, 1)
	assert.Equal(t, "#安装步骤", refs[0].Target)
}

func TestExtractDeduplicateLinks(t *testing.T) {
	content := "参考 [[数据模型]] 和 [[数据模型]]"
	refs := ExtractWikiLinks(content)
	assert.Len(t, refs, 1, "duplicate links should be deduplicated")
}

func TestExtractEmptyTarget(t *testing.T) {
	content := "空链接 [[]]"
	refs := ExtractWikiLinks(content)
	assert.Empty(t, refs)
}

func TestRemoveFencedCodeBlocks(t *testing.T) {
	content := "before\n```\ninside [[link]]\n```\nafter"
	result := removeFencedCodeBlocks(content)

	// The function should replace [[ with placeholder inside code blocks
	assert.Contains(t, result, "﹝﹝link﹞﹞", "brackets inside code blocks should be replaced with placeholders")
	assert.NotContains(t, result, "inside [[link]]",
		"should not contain original [[link]] inside code block")
}

func TestExtractLinksInMixedContent(t *testing.T) {
	content := `# 标题

这是 [[链接A]] 的引用，还有 [[链接B|文本]]。

> 引用中的 [[链接C]] 也应该被提取

- [[链接D]] 在列表中`
	refs := ExtractWikiLinks(content)
	assert.Len(t, refs, 4)
}
