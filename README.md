# demos

Small, self-contained demos. One idea per directory — each directory builds and runs on its own.

## Index

| Demo | Language | What it shows | Run |
|---|---|---|---|
| _add a row per demo_ | | | |

## Layout

```
.
├── scripts/
│   ├── new.go            # scaffold a demo: go run scripts/new.go <lang> <name>
│   └── templates/        # only for languages without a package manager (c, web)
├── go/<name>/            # go.mod via `go mod init`
├── rust/<name>/          # Cargo.toml via `cargo init`
├── python/<name>/        # pyproject.toml via `uv init`
├── c/<name>/             # Makefile + main.c from templates
└── web/<name>/           # index.html from templates
```

## Add a demo

```bash
go run scripts/new.go go wordfreq
```

That creates `go/wordfreq/` with a working manifest and stub, a README, and a
row at the top of the index above. Then:

1. replace the stub with the actual demo, and fill in `go/wordfreq/README.md`
   and the index row;
2. `git add go/wordfreq README.md && git commit -m 'go/wordfreq: <what it does>'`.

Languages: `go`, `rust`, `python`, `c`, `web`. Manifests are not hand-written —
the language's own tool does it:

| lang | manifest from | stub |
|---|---|---|
| `go` | `go mod init` | `scripts/templates/go/main.go.tmpl` |
| `rust` | `cargo init --vcs none` | written by cargo |
| `python` | `uv init --app --no-package` | written by uv |
| `c` | — | `scripts/templates/c/` |
| `web` | — | `scripts/templates/web/` |

The script refuses to overwrite an existing directory, validates the name, and
drops that language directory's `.gitkeep` on the first demo.

Rules:

1. **Self-contained.** Every demo directory owns its manifest (`go.mod`, `Cargo.toml`, `Makefile`, `package.json`). No shared code between demos, no root-level module.
2. **Independent.** Adding, rewriting, or deleting a demo never touches another one.
3. **Readable.** Each demo has a short `README.md`: what it is, how to run it, what I learned. `scripts/new.go` generates the skeleton.
4. **Committed by directory.** One commit per demo, message prefix `<lang>/<name>: ...`.

## When to split a demo into its own repo

Move it out when it becomes a project rather than a demo:

- it grows past ~1k lines or gets external users;
- it needs CI, releases, tags, or its own issues;
- it keeps evolving (like [mini-db](https://github.com/Yakumo-zi/mini-db), [yedis](https://github.com/Yakumo-zi/yedis), [bitcask](https://github.com/Yakumo-zi/bitcask));
- it gets deployed somewhere.

Extracting a directory while keeping its history:

```bash
git subtree split -P go/<name> -b <name>
git push git@github.com:Yakumo-zi/<name>.git <name>:main
```

Details, the reverse direction (`subtree add` / `push` / `pull`), and the gotchas: **[docs/git-subtree.md](docs/git-subtree.md)**.
