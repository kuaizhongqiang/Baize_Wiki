package vector

import (
	"context"
	"hash/fnv"
	"math"
	"unicode"
)

// Embedder converts text into a fixed-dimension vector embedding.
type Embedder interface {
	// Embed converts text to a vector.
	Embed(ctx context.Context, text string) ([]float32, error)

	// Dim returns the dimension of the embedding vector.
	Dim() int
}

// LocalEmbedder uses Feature Hashing to produce fixed-dimension vectors
// with zero external dependencies. It is deterministic: the same input
// always produces the same vector.
//
// Algorithm:
//  1. Tokenize: ASCII text is split on whitespace and punctuation;
//     CJK characters are split into individual runes.
//  2. Each token is hashed with FNV-1a (hash/fnv).
//  3. The hash is mapped to a bucket (0..dim-1) and a sign (+1 or -1).
//  4. The buckets are accumulated into a vector of dimension dim.
//  5. The vector is L2-normalized.
type LocalEmbedder struct {
	dim int
}

// NewLocalEmbedder creates a LocalEmbedder with the given dimension.
// If dim is 0 or negative, the default of 256 is used.
func NewLocalEmbedder(dim int) *LocalEmbedder {
	if dim <= 0 {
		dim = 256
	}
	return &LocalEmbedder{dim: dim}
}

// Dim returns the dimension of the embedding vector.
func (e *LocalEmbedder) Dim() int {
	return e.dim
}

// Embed converts text into a fixed-dimension feature-hashed vector.
func (e *LocalEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	if text == "" {
		return make([]float32, e.dim), nil
	}

	tokens := tokenize(text)
	vec := make([]float64, e.dim)

	for _, token := range tokens {
		h := fnv.New64a()
		h.Write([]byte(token))
		sum := h.Sum64()

		// bucket index: hash mod dim
		bucket := int(sum % uint64(e.dim))

		// sign: +1 if the least significant bit is 0, -1 if 1
		sign := float64(1)
		if sum&1 == 1 {
			sign = -1
		}

		vec[bucket] += sign
	}

	// L2 normalize
	l2Norm := 0.0
	for _, v := range vec {
		l2Norm += v * v
	}
	l2Norm = math.Sqrt(l2Norm)

	out := make([]float32, e.dim)
	if l2Norm > 0 {
		for i, v := range vec {
			out[i] = float32(v / l2Norm)
		}
	}

	return out, nil
}

// tokenize splits text into tokens.
// ASCII tokens are split on whitespace and punctuation boundaries.
// CJK characters (Chinese, Japanese, Korean) are split into individual runes.
// Mixed content is handled by processing each rune according to its category.
func tokenize(text string) []string {
	var tokens []string
	var buf []rune
	flush := func() {
		if len(buf) > 0 {
			tokens = append(tokens, string(buf))
			buf = nil
		}
	}

	for _, r := range text {
		if isCJK(r) {
			// CJK: flush any pending ASCII buffer, then emit this rune as its own token
			flush()
			tokens = append(tokens, string(r))
		} else if unicode.IsSpace(r) || unicode.IsPunct(r) {
			// Whitespace/punctuation: flush buffer (delimiter)
			flush()
		} else {
			// Regular ASCII/other: accumulate
			buf = append(buf, r)
		}
	}
	flush()

	return tokens
}

// isCJK reports whether a rune is a CJK (Chinese/Japanese/Korean) character.
func isCJK(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Katakana, r) ||
		unicode.Is(unicode.Hangul, r)
}

// Ensure LocalEmbedder implements Embedder.
var _ Embedder = (*LocalEmbedder)(nil)
