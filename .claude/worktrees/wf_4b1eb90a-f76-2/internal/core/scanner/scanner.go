package scanner

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
)

const (
	maxBinaryCheck = 512
)

// ScanConfig provides parameters for scanning.
type ScanConfig struct {
	MaxSize int64
	Exclude []string
	ScanAll bool // Phase 2+: if true, scan all text files; Phase 1: only .md/.mdx
}

// Scan recursively walks a root directory, returning a list of discoverd text files.
// The filter chain is: ignore rules → binary detection → file size limit.
func Scan(ctx context.Context, root string, cfg ScanConfig) ([]model.FileInfo, error) {
	// Load .baizeignore if present
	matcher := NewBuiltinMatcher()
	ignorePath := filepath.Join(root, ".baizeignore")
	if f, err := os.Open(ignorePath); err == nil {
		_ = f.Close()
		if rm, err := NewRuleMatcherFromFile(ignorePath); err == nil {
			matcher = rm
		}
	}

	// Add configured exclude patterns
	for _, p := range cfg.Exclude {
		matcher.addPattern(p)
	}

	var files []model.FileInfo

	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		relPath, _ := filepath.Rel(root, path)

		// Check ignore rules
		if matcher.Match(relPath, info.IsDir()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories (they're not files to scan)
		if info.IsDir() {
			return nil
		}

		// Skip symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		// Skip files exceeding max size
		if cfg.MaxSize > 0 && info.Size() > cfg.MaxSize {
			return nil
		}

		// Phase 1: only .md/.mdx files (unless ScanAll is set)
		ext := filepath.Ext(path)
		if !cfg.ScanAll && ext != ".md" && ext != ".mdx" {
			return nil
		}

		// Binary detection
		isBin, err := isBinaryFile(path)
		if err != nil {
			return nil // skip files we can't read
		}
		if isBin {
			return nil
		}

		files = append(files, model.FileInfo{
			Path:      relPath,
			AbsPath:   path,
			Size:      info.Size(),
			Extension: ext,
		})

		return nil
	})

	return files, err
}

// isBinaryFile checks whether a file appears to be binary by reading its first
// 512 bytes and looking for null bytes or invalid UTF-8 sequences.
func isBinaryFile(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()

	buf := make([]byte, maxBinaryCheck)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return false, err
	}

	return isBinaryData(buf[:n]), nil
}

// isBinaryData checks if the given data appears to be binary.
func isBinaryData(data []byte) bool {
	checkLen := min(len(data), maxBinaryCheck)
	for i := 0; i < checkLen; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return !utf8.Valid(data[:checkLen])
}
