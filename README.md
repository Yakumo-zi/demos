# demos

Small, self-contained demos. One idea per directory — each directory builds and runs on its own.

## Index

| Demo | Language | What it shows | Run |
|---|---|---|---|
| _add a row per demo_ | | | |

## Layout

```
.
├── _template/README.md   # copy this for a new demo
├── go/<name>/            # go.mod lives here
├── rust/<name>/          # Cargo.toml lives here
├── c/<name>/             # Makefile lives here
├── python/<name>/        # pyproject.toml or a single file
└── web/<name>/           # index.html, servable from GitHub Pages
```

Rules:

1. **Self-contained.** Every demo directory owns its manifest (`go.mod`, `Cargo.toml`, `Makefile`, `package.json`). No shared code between demos, no root-level module.
2. **Independent.** Adding, rewriting, or deleting a demo never touches another one.
3. **Readable.** Each demo has a short `README.md`: what it is, how to run it, what I learned. Copy [`_template`](_template).
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
