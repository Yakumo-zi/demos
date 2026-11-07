// Command new scaffolds a demo directory in this repository.
//
//	scripts              ← this program, run with: go run scripts/new.go <lang> <name>
//	go|rust|python/...   ← manifest created by the language's own package manager
//	c|web/...            ← rendered from scripts/templates
//
// Every demo gets a README.md from templates/demo.README.md.tmpl and a row at
// the top of the index table in the repository README.
package main

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

//go:embed templates
var templates embed.FS

// repoRoot is the directory that holds README.md, i.e. the parent of scripts/.
// go:embed cannot tell us that at runtime, so it is resolved from the executable
// path when built and from the working directory when run with `go run`.
var repoRoot = findRepoRoot()

// lang describes how one language is scaffolded.
type lang struct {
	dir      string // directory under the repo root
	display  string // name used in the index table
	run      string // what to type inside the demo directory
	manifest string // provided by the package manager, or rendered from templates
	init     func(root, dir, name string) error
}

func main() {
	if len(os.Args) != 3 {
		usage()
	}
	langs := map[string]lang{
		"go":     {dir: "go", display: "Go", run: "go run .", manifest: "go mod init", init: initGo},
		"rust":   {dir: "rust", display: "Rust", run: "cargo run", manifest: "cargo init", init: initRust},
		"python": {dir: "python", display: "Python", run: "uv run main.py", manifest: "uv init", init: initPython},
		"c":      {dir: "c", display: "C", run: "make && ./" + os.Args[2], manifest: "templates", init: initC},
		"web":    {dir: "web", display: "HTML/CSS/JS", run: "open index.html", manifest: "templates", init: initWeb},
	}

	l, ok := langs[os.Args[1]]
	if !ok {
		fmt.Fprintf(os.Stderr, "new: unknown language %q\n", os.Args[1])
		usage()
	}
	name := os.Args[2]
	if !regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`).MatchString(name) {
		die("name must be lowercase letters, digits and dashes: %q", name)
	}

	rel := filepath.Join(l.dir, name)
	abs := filepath.Join(repoRoot, rel)
	if _, err := os.Stat(abs); err == nil {
		die("%s already exists", rel)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		die("%v", err)
	}

	if err := l.init(repoRoot, rel, name); err != nil {
		die("%s: %v", l.manifest, err)
	}
	if err := render(abs, "templates/demo.README.md.tmpl", "README.md", l.templateData(rel, name)); err != nil {
		die("%v", err)
	}
	// The language directory no longer needs its placeholder.
	os.Remove(filepath.Join(repoRoot, l.dir, ".gitkeep"))

	if err := insertIndexRow(filepath.Join(repoRoot, "README.md"), l.indexRow(rel, name)); err != nil {
		die("%v", err)
	}

	fmt.Printf("created %s (%s, via %s)\n\n", rel, l.display, l.manifest)
	fmt.Printf("next:\n")
	fmt.Printf("  1. edit %s/README.md and its row in README.md\n", rel)
	fmt.Printf("  2. git add %s README.md && git commit -m '%s: <what it does>'\n\n", rel, rel)
	fmt.Printf("run: cd %s && %s\n", rel, l.run)
}

func usage() {
	fmt.Fprint(os.Stderr, `usage: go run scripts/new.go <lang> <name>

  lang  go | rust | python | c | web
  name  lowercase letters, digits and dashes

example: go run scripts/new.go go wordfreq
`)
	os.Exit(2)
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "new: "+format+"\n", args...)
	os.Exit(1)
}

func (l lang) templateData(rel, name string) map[string]string {
	return map[string]string{
		"Name":     name,
		"Dir":      rel,
		"Display":  l.display,
		"Run":      l.run,
		"Manifest": l.manifest,
	}
}

func (l lang) indexRow(rel, name string) string {
	return fmt.Sprintf("| [%s](%s) | %s | _what it shows_ | `%s` |", rel, rel, l.display, l.run)
}

// --- package managers -------------------------------------------------------

// initGo creates the module; there is no Go command that also writes main.go.
func initGo(root, rel, name string) error {
	dir := filepath.Join(root, rel)
	if err := runCmd(dir, "go", "mod", "init", "demos/"+filepath.ToSlash(rel)); err != nil {
		return err
	}
	return render(dir, "templates/go/main.go.tmpl", "main.go", map[string]string{"Name": name})
}

func initRust(root, rel, name string) error {
	// cargo init works in an existing directory (unlike cargo new); --vcs none
	// keeps it out of git, so the demo stays a plain subdirectory.
	return runCmd(filepath.Join(root, rel), "cargo", "init", "--vcs", "none", "--name", name)
}

func initPython(root, rel, name string) error {
	// --no-package keeps a flat main.py instead of a src/ package layout;
	// --no-workspace stops uv from joining an enclosing workspace.
	return runCmd(root, "uv", "init", "--app", "--no-package", "--vcs", "none",
		"--no-workspace", "--name", name, rel)
}

func initC(root, rel, name string) error {
	dir := filepath.Join(root, rel)
	data := map[string]string{"Name": name}
	if err := render(dir, "templates/c/main.c.tmpl", "main.c", data); err != nil {
		return err
	}
	return render(dir, "templates/c/Makefile.tmpl", "Makefile", data)
}

func initWeb(root, rel, name string) error {
	return render(filepath.Join(root, rel), "templates/web/index.html.tmpl", "index.html",
		map[string]string{"Name": name})
}

// --- helpers ----------------------------------------------------------------

func render(dir, tmplPath, outName string, data any) error {
	raw, err := templates.ReadFile(tmplPath)
	if err != nil {
		return err
	}
	tmpl, err := template.New(outName).Option("missingkey=error").Parse(string(raw))
	if err != nil {
		return err
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, outName), []byte(buf.String()), 0o644)
}

func runCmd(root string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, out)
	}
	return nil
}

// insertIndexRow puts row at the top of the index table in README.md, replacing
// the placeholder row when the table is still empty.
func insertIndexRow(readmePath, row string) error {
	raw, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(raw), "\n")
	hasPlaceholder := false
	for _, line := range lines {
		if strings.HasPrefix(line, "| _add a row per demo_") {
			hasPlaceholder = true
			break
		}
	}

	// Replace the placeholder when it is still there, otherwise insert at the
	// top of the table. Deciding this up front avoids doing both.
	out := make([]string, 0, len(lines)+1)
	inIndex, inserted := false, false
	for _, line := range lines {
		switch {
		case hasPlaceholder && strings.HasPrefix(line, "| _add a row per demo_"):
			out = append(out, row)
			continue
		case !hasPlaceholder && inIndex && strings.HasPrefix(line, "|---") && !inserted:
			out = append(out, line, row)
			inserted = true
			continue
		case strings.HasPrefix(line, "## Index"):
			inIndex = true
		}
		out = append(out, line)
	}
	if !hasPlaceholder && !inserted {
		return errors.New("README.md: index table not found")
	}
	return os.WriteFile(readmePath, []byte(strings.Join(out, "\n")), 0o644)
}

// findRepoRoot walks up from the executable or the working directory until it
// finds the README.md that belongs to this repository.
func findRepoRoot() string {
	if exe, err := os.Executable(); err == nil {
		if root := walkUp(filepath.Dir(exe)); root != "" {
			return root
		}
	}
	if wd, err := os.Getwd(); err == nil {
		if root := walkUp(wd); root != "" {
			return root
		}
	}
	fmt.Fprintln(os.Stderr, "new: cannot find the repository root (looked for scripts/templates)")
	os.Exit(1)
	return ""
}

func walkUp(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, "scripts", "templates")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
