package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractGoComments(t *testing.T) {
	content := `// Package main is the entry point.
// It demonstrates how to write Go code.
package main

import "fmt"
`
	result := ExtractComments("main.go", content)
	assert.Equal(t, "Package main is the entry point.\nIt demonstrates how to write Go code.", result)
}

func TestExtractGoBlockComments(t *testing.T) {
	content := `/*
Package main provides the entry point.
This is a multi-line block comment.
*/
package main

import "fmt"
`
	result := ExtractComments("main.go", content)
	assert.Equal(t, "Package main provides the entry point.\nThis is a multi-line block comment.", result)
}

func TestExtractPythonComments(t *testing.T) {
	content := `# This module provides utility functions.
# It handles file I/O and data processing.
import os
import sys
`
	result := ExtractComments("utils.py", content)
	assert.Equal(t, "This module provides utility functions.\nIt handles file I/O and data processing.", result)
}

func TestExtractNoComments(t *testing.T) {
	content := `package main

func main() {}
`
	result := ExtractComments("main.go", content)
	assert.Empty(t, result, "no comments should return empty")
}

func TestExtractTopCommentBlock(t *testing.T) {
	content := `// Top-level documentation block.
// This describes the package.
package main

// This is an internal function, not extracted.
func internal() {}
`
	result := ExtractComments("main.go", content)
	assert.Equal(t, "Top-level documentation block.\nThis describes the package.", result)
}

func TestExtractGoLeadingCodeNoComment(t *testing.T) {
	content := `package main

// This comment is after code
func main() {}
`
	result := ExtractComments("main.go", content)
	assert.Empty(t, result, "comments after code should not be extracted")
}

func TestExtractJavaScriptComments(t *testing.T) {
	content := `// React component for user profile.
// Handles avatar and bio display.
import React from 'react';

function Profile() {
  return <div>Profile</div>;
}
`
	result := ExtractComments("Profile.jsx", content)
	assert.Equal(t, "React component for user profile.\nHandles avatar and bio display.", result)
}

func TestExtractHashEmptyFile(t *testing.T) {
	result := ExtractComments("test.py", "")
	assert.Empty(t, result)
}

func TestExtractCStyleEmptyFile(t *testing.T) {
	result := ExtractComments("main.go", "")
	assert.Empty(t, result)
}

func TestExtractUnknownExtension(t *testing.T) {
	result := ExtractComments("data.json", `{"key": "value"}`)
	assert.Empty(t, result, "unknown extension should return empty")
}

func TestExtractBlockMixedStyle(t *testing.T) {
	content := `/* Block comment header */
// Line comment follows
package main
`
	result := ExtractComments("main.go", content)
	assert.Contains(t, result, "Block comment header", "should include block comment")
	assert.Contains(t, result, "Line comment follows", "should include line comment")
}

func TestExtractPythonNoLeadingComment(t *testing.T) {
	content := `import os
# comment after code
def foo():
    pass
`
	result := ExtractComments("test.py", content)
	assert.Empty(t, result, "comments after code should be ignored")
}
