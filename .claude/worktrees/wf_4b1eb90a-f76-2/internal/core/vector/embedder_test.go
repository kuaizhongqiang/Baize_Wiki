package vector

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalEmbedDeterministic(t *testing.T) {
	e := NewLocalEmbedder(256)
	ctx := context.Background()

	v1, err := e.Embed(ctx, "hello world")
	require.NoError(t, err)

	v2, err := e.Embed(ctx, "hello world")
	require.NoError(t, err)

	assert.Equal(t, v1, v2, "same input should produce same vector")
}

func TestLocalEmbedEmpty(t *testing.T) {
	e := NewLocalEmbedder(256)
	ctx := context.Background()

	v, err := e.Embed(ctx, "")
	require.NoError(t, err)

	// Empty input should produce a zero vector
	for _, val := range v {
		assert.Equal(t, float32(0), val)
	}
}

func TestLocalEmbedDim(t *testing.T) {
	e := NewLocalEmbedder(256)
	ctx := context.Background()

	v, err := e.Embed(ctx, "test")
	require.NoError(t, err)
	assert.Len(t, v, 256)
}

func TestLocalEmbedDefaultDim(t *testing.T) {
	e := NewLocalEmbedder(0)
	assert.Equal(t, 256, e.Dim())

	e2 := NewLocalEmbedder(-1)
	assert.Equal(t, 256, e2.Dim())
}

func TestLocalEmbedDimMethod(t *testing.T) {
	e := NewLocalEmbedder(128)
	assert.Equal(t, 128, e.Dim())
}

func TestLocalEmbedL2Normalized(t *testing.T) {
	e := NewLocalEmbedder(256)
	ctx := context.Background()

	v, err := e.Embed(ctx, "this is a test sentence with multiple tokens")
	require.NoError(t, err)

	// Check L2 norm is approximately 1
	var sumSquares float64
	for _, val := range v {
		sumSquares += float64(val) * float64(val)
	}
	norm := math.Sqrt(sumSquares)
	assert.InDelta(t, 1.0, norm, 0.0001, "L2 norm should be ~1.0")
}

func TestLocalEmbedDifferentInputs(t *testing.T) {
	e := NewLocalEmbedder(256)
	ctx := context.Background()

	v1, err := e.Embed(ctx, "hello world")
	require.NoError(t, err)

	v2, err := e.Embed(ctx, "goodbye world")
	require.NoError(t, err)

	// Different inputs should produce different vectors
	sim := cosineSimilarity(v1, v2)
	assert.Less(t, sim, 0.99, "different inputs should have different vectors")
}

func TestLocalEmbedChineseSimilarity(t *testing.T) {
	e := NewLocalEmbedder(256)
	ctx := context.Background()

	// Similar Chinese texts should have higher similarity than unrelated ones
	similarA, _ := e.Embed(ctx, "如何配置数据库")
	similarB, _ := e.Embed(ctx, "数据库配置方法")
	unrelated, _ := e.Embed(ctx, "天气晴朗")

	simSimilar := cosineSimilarity(similarA, similarB)
	simUnrelated := cosineSimilarity(similarA, unrelated)

	assert.Greater(t, simSimilar, simUnrelated,
		"similar Chinese texts should have higher cosine similarity")
}

func TestLocalEmbedMixedContent(t *testing.T) {
	e := NewLocalEmbedder(256)
	ctx := context.Background()

	v, err := e.Embed(ctx, "Go语言 programming 测试")
	require.NoError(t, err)
	assert.Len(t, v, 256)

	// Should not panic or produce NaN
	for _, val := range v {
		assert.False(t, math.IsNaN(float64(val)))
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"hello world", []string{"hello", "world"}},
		{"hello,world", []string{"hello", "world"}},
		{"hello.world!", []string{"hello", "world"}},
		{"如何配置数据库", []string{"如", "何", "配", "置", "数", "据", "库"}},
		{"hello世界", []string{"hello", "世", "界"}},
		{"", nil},
		{"   ", nil},
		{"a,b,c", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := tokenize(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
