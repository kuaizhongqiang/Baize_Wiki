#!/bin/bash
# Generate 1000+ .md test files for benchmarking.
# Usage: bash scripts/generate-testdata.sh [output-dir] [count]
# Default: generates 1000 files in testdata/large-project/

set -e

OUT_DIR="${1:-testdata/large-project}"
COUNT="${2:-1000}"

mkdir -p "$OUT_DIR"

echo "Generating $COUNT test files in $OUT_DIR..."

for i in $(seq 1 $COUNT); do
  # Distribute across categories
  case $((i % 8)) in
    0) CAT="architecture" ;;
    1) CAT="api" ;;
    2) CAT="guide" ;;
    3) CAT="reference" ;;
    4) CAT="tutorial" ;;
    5) CAT="design" ;;
    6) CAT="config" ;;
    7) CAT="deployment" ;;
  esac

  SUB=$((i / 100 + 1))
  SUBDIR="$OUT_DIR/$CAT/chapter-$SUB"
  mkdir -p "$SUBDIR"

  cat > "$SUBDIR/page-$i.md" << EOF
---
title: "Page $i - $CAT Chapter $SUB"
category: "$CAT"
tags: [$CAT, generated, benchmark]
weight: $i
---

# Page $i: $CAT Chapter $SUB

This is auto-generated test file number $i for benchmarking Baize Wiki's scan, parse, and generate pipeline.

## Introduction

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.

## Content Section

The purpose of this file is to provide realistic test data for performance testing.

### Subsection A

- Item 1 in the list
- Item 2 in the list
- Item 3 with \`inline code\` example

### Subsection B

More detailed content to simulate a real documentation page.

\`\`\`
code block example
for testing parser performance
\`\`\`

## Related Pages

- [[Page $((i-1)) - $(printf "Page %03d" $((i-1)))]]
- [[Page $((i+1)) - $(printf "Page %03d" $((i+1)))]]
EOF

  if [ $((i % 100)) -eq 0 ]; then
    echo "  $i files generated..."
  fi
done

echo "Done: $COUNT files generated in $OUT_DIR"
echo "Structure:"
du -sh "$OUT_DIR"
find "$OUT_DIR" -type d | head -20
