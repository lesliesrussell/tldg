# tldg — too long didn't git

Local-first, evidence-grounded repository intelligence from the terminal.

`tldg` explains what a repository actually does, answers questions from code and
history, and cites its evidence — running locally by default. See
[spec.md](spec.md) for the full design.

## Status

Early development. Implemented milestones:

- **M0 — Foundation:** CLI skeleton, YAML config, local target parsing, Git
  worktree/remote/ref inspection, Ollama connectivity, text/JSON renderers,
  `tldg doctor`.
- **M1 — Local summary MVP:** safe file traversal, documentation/manifest
  extraction, language + dependency detection, SQLite FTS index, structured
  repository profile, and a local-model `tldg summary` with file citations.

## Requirements

- Go 1.26+
- `git` on `PATH`
- [Ollama](https://ollama.com) (or an OpenAI-compatible endpoint) for synthesis;
  optional — `--offline` works without a model.

## Build

```sh
go build -o tldg ./cmd/tldg
```

## Usage

```sh
tldg doctor                 # diagnose git, model, index, config
tldg summary .              # cited overview via the local model
tldg summary . --offline    # structured profile, no model call
tldg summary . --json       # stable machine-readable output (schema 1.0)
tldg config path            # show active config/data/cache paths
```

The default model is `qwen2.5-coder:7b` via Ollama; override in
`config.yaml` or with `--model`. Configuration lives at the path reported by
`tldg config path`.

## Development

```sh
go test ./...
go vet ./...
```

Project layout follows [spec.md §19](spec.md). Issue tracking uses beads
(`bd`); each code block carries its originating issue id as a comment.
