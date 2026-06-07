package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractGoLineComments(t *testing.T) {
	content := `// Package main is the entry point.
// It initializes the application.
package main

func main() {}
`
	comments := ExtractComments("main.go", content)
	assert.Contains(t, comments, "Package main")
	assert.Contains(t, comments, "entry point")
	assert.Contains(t, comments, "initializes")
}

func TestExtractGoBlockComments(t *testing.T) {
	content := `/*
Package main provides the application entry point.

It handles initialization and cleanup.
*/
package main
`
	comments := ExtractComments("main.go", content)
	assert.Contains(t, comments, "Package main")
	assert.Contains(t, comments, "initialization")
}

func TestExtractPythonComments(t *testing.T) {
	content := `# This is a Python module
# It provides utility functions

def hello():
    pass
`
	comments := ExtractComments("utils.py", content)
	assert.Contains(t, comments, "Python module")
	assert.Contains(t, comments, "utility functions")
}

func TestExtractNoComments(t *testing.T) {
	content := `package main
func main() {}
`
	comments := ExtractComments("main.go", content)
	assert.Empty(t, comments)
}

func TestExtractTopCommentBlock(t *testing.T) {
	content := `// Package doc
// Contains the main logic
//
// This is a second paragraph

// Another function
func Foo() {}
`
	comments := ExtractComments("doc.go", content)
	assert.Contains(t, comments, "Package doc")
	assert.Contains(t, comments, "main logic")
	assert.Contains(t, comments, "second paragraph")
	assert.NotContains(t, comments, "Another function")
}

func TestExtractJavaScriptComments(t *testing.T) {
	content := `// JavaScript module
// Provides utility functions
const util = require("util");
`
	comments := ExtractComments("utils.js", content)
	assert.Contains(t, comments, "JavaScript module")
}

func TestExtractRustComments(t *testing.T) {
	content := `//! Rust crate documentation
//! This is the main module

pub fn main() {}
`
	comments := ExtractComments("main.rs", content)
	assert.Contains(t, comments, "Rust crate")
}

func TestExtractUnknownExtension(t *testing.T) {
	content := "# Comment in unknown file"
	comments := ExtractComments("makefile", content)
	assert.Empty(t, comments)
}
