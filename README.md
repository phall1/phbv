# phbv

A terminal UI for [beads](https://github.com/gastownhall/beads) (`bd`) — view,
filter, and manage your dependency-aware issues at a glance.

phbv talks to the `bd` CLI exclusively via `bd … --json` subprocesses; it never
touches the underlying Dolt store. That keeps it robust across beads storage
changes — the JSON contract is the only thing it depends on, and golden tests
pin that boundary so a `bd` upgrade fails CI instead of silently breaking the UI.

## Requirements

- Go 1.26+
- `bd` on your `PATH` ([install beads](https://github.com/gastownhall/beads))

## Install

```sh
# Go toolchain (installs to `go env GOBIN`, else $GOPATH/bin)
go install github.com/phall1/phbv@latest

# …or grab the latest release binary for your OS/arch
curl -fsSL https://raw.githubusercontent.com/phall1/phbv/main/install.sh | sh
#   override the dir: … | PREFIX=/usr/local/bin sh

# …or Homebrew (coming once the tap exists)
# brew install phall1/tap/phbv
```

Releases are cut by pushing a `v*` tag — GitHub Actions runs
[GoReleaser](https://goreleaser.com) to build the cross-platform archives and
publish the release (`just release v0.1.0`).

## Run

```sh
phbv                     # resolve the .beads workspace from the cwd
phbv --dir path/.beads   # point at a specific workspace
BEADS_DIR=… phbv         # or via the env var bd itself uses
```

## Keys

| Key | Action |
|---|---|
| `tab` / `shift+tab` | cycle list → ready → kanban |
| `j` / `k` (or arrows) | move cursor |
| `enter` / `l` | open detail |
| `esc` / `h` | back out of a modal view |
| `/` | fuzzy filter the list |
| `t` | dependency tree of the selected issue |
| `p` | switch project (sibling `.beads` workspaces) |
| `x` / `X` | close / reopen |
| `+` / `-` | raise / lower priority |
| `a` | assign to you (`git config user.email`, else `$USER`) |
| `A` | toggle showing closed issues (`bd` hides them by default) |
| `r` | refresh |
| `q` | quit |

> The list view mirrors `bd list`, which hides closed issues — so a fully-done
> board looks empty. Press `A` to include closed work; the header shows `(all)`.

## Configure keybindings

Defaults are baked into the binary (canonical source:
`internal/config/keys.default.toml`). Override any subset at
`$XDG_CONFIG_HOME/phbv/keys.toml` (typically `~/.config/phbv/keys.toml`) — your
file is merged over the defaults action-by-action, so you only list what you
change:

```toml
[keys]
quit   = ["ctrl+c"]
filter = ["/", "f"]
```

## Develop

Tasks run through [`just`](https://github.com/casey/just):

```sh
just         # list tasks
just check   # gofmt + vet + test gate
just build   # ./bin/phbv
just run     # go run against the cwd's .beads
just snapshot  # local goreleaser build (no publish), validates the release config
```

Architecture: `internal/bd` is the anti-corruption layer (the only package that
knows `bd`'s JSON shape); `internal/config` loads the keymap; `internal/ui`
is the Bubble Tea app. Live updates come from an fsnotify watch on `.beads`.
