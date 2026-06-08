#!/bin/bash
# Extract all .cs files from F:/Project/ into testdata/cs-samples/
# Preserves relative path structure with project name prefix.
#
# Usage: bash scripts/extract-cs-samples.sh
# Output: testdata/cs-samples/<project>/<relative-path>/file.cs

set -e

SRC_DIR="F:/Project"
OUT_DIR="testdata/cs-samples"

# Clear previous output
rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR"

echo "Scanning for .cs files in $SRC_DIR ..."

total=0
for project_dir in "$SRC_DIR"/*/; do
  project_name=$(basename "$project_dir")

  # Skip known non-C# directories
  case "$project_name" in
    Baize_Wiki|AgentRouter|DeepSeek-*|StockAnalysis|TradingWeb|test-agentrouter)
      echo "  [skip] $project_name (not a C# project)"
      continue
      ;;
  esac

  count=0
  while IFS= read -r -d '' cs_file; do
    # Get relative path within the project
    rel_path="${cs_file#$project_dir}"
    target="$OUT_DIR/$project_name/$rel_path"
    mkdir -p "$(dirname "$target")"
    cp "$cs_file" "$target"
    count=$((count + 1))
    total=$((total + 1))
  done < <(find "$project_dir" -name "*.cs" -type f -print0 2>/dev/null || true)

  if [ "$count" -gt 0 ]; then
    echo "  $project_name: $count files"
  fi
done

echo ""
echo "Total: $total .cs files extracted to $OUT_DIR"
du -sh "$OUT_DIR" 2>/dev/null || echo "(output directory created)"
