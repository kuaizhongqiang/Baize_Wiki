package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// RuleMatcher evaluates ignore rules against file paths.
type RuleMatcher struct {
	rules       []rule
	builtinOnly bool
}

type rule struct {
	pattern string
	negate  bool // leading ! means negate
	dirOnly bool // trailing / means directory only
}

// NewRuleMatcher creates a RuleMatcher from a list of patterns.
// Patterns use gitignore-compatible syntax.
func NewRuleMatcher(patterns []string) *RuleMatcher {
	rm := &RuleMatcher{}
	for _, p := range patterns {
		rm.addPattern(p)
	}
	return rm
}

// NewBuiltinMatcher creates a RuleMatcher with only the built-in ignore rules.
func NewBuiltinMatcher() *RuleMatcher {
	rm := &RuleMatcher{builtinOnly: true}
	for _, p := range builtinIgnores {
		rm.addPattern(p)
	}
	return rm
}

// NewRuleMatcherFromFile reads ignore rules from a file (gitignore-compatible).
func NewRuleMatcherFromFile(path string) (*RuleMatcher, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rm := &RuleMatcher{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		rm.addPattern(line)
	}
	return rm, scanner.Err()
}

// Match checks whether a path should be ignored.
// Returns true if the path matches any ignore rule (and is not negated).
func (rm *RuleMatcher) Match(path string, isDir bool) bool {
	matched := false

	for _, r := range rm.rules {
		// For dir-only rules, skip non-directories
		if r.dirOnly && !isDir {
			continue
		}

		if matchGlob(path, r.pattern) {
			matched = !r.negate
		}
	}

	return matched
}

// HasBuiltinOnly returns true if only built-in rules are loaded.
func (rm *RuleMatcher) HasBuiltinOnly() bool {
	return rm.builtinOnly
}

func (rm *RuleMatcher) addPattern(line string) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return
	}

	r := rule{}

	// Negation
	if strings.HasPrefix(line, "!") {
		r.negate = true
		line = line[1:]
	}

	// Directory-only marker
	if strings.HasSuffix(line, "/") {
		r.dirOnly = true
		line = strings.TrimSuffix(line, "/")
	}

	// Strip leading slash (means relative to root)
	line = strings.TrimPrefix(line, "/")

	r.pattern = line
	rm.rules = append(rm.rules, r)
}

// matchGlob checks if a path matches a gitignore-style pattern.
// This is a simplified implementation that handles basic glob patterns.
func matchGlob(path, pattern string) bool {
	// Normalize to forward slashes
	path = filepath.ToSlash(path)
	pattern = filepath.ToSlash(pattern)

	// Double-star at the end means match everything inside
	if strings.HasPrefix(pattern, "**/") {
		pattern = pattern[3:]
		return matchName(filepath.Base(path), pattern)
	}

	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return strings.HasPrefix(path, prefix)
	}

	// If pattern contains a slash, match against full path
	if strings.Contains(pattern, "/") {
		return matchName(path, pattern)
	}

	// Otherwise match against filename only
	return matchName(filepath.Base(path), pattern) || matchName(path, pattern)
}

// matchName matches a name against a glob pattern with * and ? support.
func matchName(name, pattern string) bool {
	// Exact match
	if name == pattern {
		return true
	}

	// Wildcard * matches everything
	if pattern == "*" {
		return true
	}

	// Simple glob: only handle * and ? for Phase 1
	pi := 0
	ni := 0
	star := -1
	match := 0

	for ni < len(name) {
		if pi < len(pattern) && (pattern[pi] == '?' || pattern[pi] == name[ni]) {
			pi++
			ni++
		} else if pi < len(pattern) && pattern[pi] == '*' {
			star = pi
			match = ni
			pi++
		} else if star != -1 {
			pi = star + 1
			match++
			ni = match
		} else {
			return false
		}
	}

	for pi < len(pattern) && pattern[pi] == '*' {
		pi++
	}

	return pi == len(pattern)
}

// builtinIgnores are the default ignore patterns.
var builtinIgnores = []string{
	".git/",
	".svn/",
	".hg/",
	"node_modules/",
	"vendor/",
	"__pycache__/",
	"dist/",
	"build/",
	"target/",
	"bin/",
	"*.exe",
	"*.dll",
	"*.so",
	"*.dylib",
	".DS_Store",
	"Thumbs.db",
	"*.swp",
	"*.swo",
	".idea/",
	".vscode/",
}
