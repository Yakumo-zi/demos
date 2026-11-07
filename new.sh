#!/usr/bin/env bash
# Scaffold a new demo directory: manifest, stub, README, and an index row in README.md.
#
#   ./new.sh go wordfreq
#
# Creates go/wordfreq/ with go.mod + main.go + README.md, adds a row to the
# index table, and drops the .gitkeep of that language directory.
set -euo pipefail
cd "$(dirname "$0")"

usage() {
	cat >&2 <<'EOF'
usage: ./new.sh <lang> <name>

  lang  go | rust | c | python | web
  name  lowercase letters, digits and dashes

example: ./new.sh go wordfreq
EOF
	exit 2
}

[[ $# -eq 2 ]] || usage
lang=$1
name=$2

case $lang in
	go | rust | c | python | web) ;;
	*)
		echo "new.sh: unknown language '$lang'" >&2
		usage
		;;
esac

case $name in
	'' | *[!a-z0-9-]*)
		echo "new.sh: name must be lowercase letters, digits and dashes: '$name'" >&2
		exit 2
		;;
esac

dir="$lang/$name"
if [[ -e $dir ]]; then
	echo "new.sh: $dir already exists" >&2
	exit 1
fi
mkdir -p "$dir"

case $lang in
	go)
		printf 'module demos/%s/%s\n\ngo 1.22\n' "$lang" "$name" >"$dir/go.mod"
		cat >"$dir/main.go" <<EOF
package main

import "fmt"

func main() {
	fmt.Println("hello from $name")
}
EOF
		langname="Go"
		run="go run ."
		;;
	rust)
		mkdir -p "$dir/src"
		printf '[package]\nname = "%s"\nversion = "0.1.0"\nedition = "2021"\n\n[dependencies]\n' "$name" >"$dir/Cargo.toml"
		printf 'fn main() {\n    println!("hello from %s");\n}\n' "$name" >"$dir/src/main.rs"
		langname="Rust"
		run="cargo run"
		;;
	c)
		printf '#include <stdio.h>\n\nint main(void) {\n\tputs("hello from %s");\n\treturn 0;\n}\n' "$name" >"$dir/main.c"
		# printf, not a heredoc, so the recipe really starts with a tab.
		printf 'CC ?= cc\nCFLAGS ?= -std=c11 -Wall -Wextra -O2\n\n%s: main.c\n\t$(CC) $(CFLAGS) -o %s main.c\n\nclean:\n\trm -f %s\n' \
			"$name" "$name" "$name" >"$dir/Makefile"
		langname="C"
		run="make && ./$name"
		;;
	python)
		printf 'def main() -> None:\n    print("hello from %s")\n\n\nif __name__ == "__main__":\n    main()\n' "$name" >"$dir/main.py"
		langname="Python"
		run="python3 main.py"
		;;
	web)
		cat >"$dir/index.html" <<EOF
<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>$name</title>
</head>
<body>
<h1>$name</h1>
</body>
</html>
EOF
		langname="HTML/CSS/JS"
		run="open index.html"
		;;
esac

cat >"$dir/README.md" <<EOF
# $name

One sentence: what it is.

## Run

\`\`\`bash
cd $dir
$run
\`\`\`

## What it shows

- the trick / algorithm / API being demonstrated
- anything non-obvious

## Notes

- learned: one thing worth remembering
EOF

# First demo in a language directory makes its .gitkeep unnecessary.
rm -f "$lang/.gitkeep"

row="| [$dir]($dir) | $langname | _what it shows_ | \`$run\` |"
if grep -q '^| _add a row per demo_' README.md; then
	# Table is still empty: replace the placeholder row.
	awk -v row="$row" '{ if ($0 ~ /^\| _add a row per demo_/) print row; else print }' README.md >README.md.tmp
else
	# Otherwise insert at the top of the table, right under the header separator.
	awk -v row="$row" '
		{ print }
		/^## Index/ { inindex = 1 }
		inindex && /^\|---/ { print row; inindex = 0 }
	' README.md >README.md.tmp
fi
mv README.md.tmp README.md

echo "created $dir"
echo
echo "next:"
echo "  1. edit $dir/README.md and the '$dir' row in README.md"
echo "  2. git add $dir README.md && git commit -m '$dir: <what it does>'"
