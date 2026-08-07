#!/usr/bin/env bash

set -euo pipefail

usage() {
  printf 'Usage: %s <new-module-path>\n' "$(basename "$0")"
  printf 'Example: %s github.com/acme/my-service\n' "$(basename "$0")"
}

if [[ $# -ne 1 ]]; then
  usage >&2
  exit 1
fi

new_module=$1
if [[ -z "$new_module" || "$new_module" == /* || "$new_module" == */ || "$new_module" == *//* || "$new_module" =~ [[:space:]\\] ]]; then
  printf 'Invalid module path: %s\n' "$new_module" >&2
  exit 1
fi

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
project_root=$(cd -- "$script_dir/.." && pwd)
go_mod="$project_root/go.mod"

if [[ ! -f "$go_mod" ]]; then
  printf 'go.mod not found: %s\n' "$go_mod" >&2
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  printf 'The go command is required but was not found.\n' >&2
  exit 1
fi

old_module=$(awk '$1 == "module" { print $2; exit }' "$go_mod")
if [[ -z "$old_module" ]]; then
  printf 'No module directive found in %s\n' "$go_mod" >&2
  exit 1
fi

if [[ "$old_module" == "$new_module" ]]; then
  printf 'Module path is already %s\n' "$new_module"
  exit 0
fi

# Let the Go tool validate the new path and update only the module directive.
go mod edit -module="$new_module" "$go_mod"

changed_files=0
while IFS= read -r -d '' file; do
  if ! grep -Fq -- "$old_module" "$file"; then
    continue
  fi

  temp_file="${file}.rename-module.$$"
  awk -v old="$old_module" -v new="$new_module" '
    function replace_all(text, position, result) {
      result = ""
      while ((position = index(text, old)) != 0) {
        result = result substr(text, 1, position - 1) new
        text = substr(text, position + length(old))
      }
      return result text
    }
    { print replace_all($0) }
  ' "$file" > "$temp_file"
  mv -- "$temp_file" "$file"
  changed_files=$((changed_files + 1))
done < <(
  find "$project_root" \
    -type d \( -name .git -o -name vendor \) -prune -o \
    -type f -name '*.go' -print0
)

printf 'Module path changed: %s -> %s\n' "$old_module" "$new_module"
printf 'Updated %d Go file(s).\n' "$changed_files"
