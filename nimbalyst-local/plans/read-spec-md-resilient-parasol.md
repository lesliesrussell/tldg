# tldg — M0 + M1 Build Plan

## Context

`tldg` ("too long didn't git") is a local-first, evidence-grounded repository
intelligence CLI in Go. The full design lives in [spec.md](/Users/leslierussell/repo/tldg/spec.md).
The repo is currently greenfield — only `spec.md` exists, not even git-initialized.

This plan covers the first two milestones from spec §24:

- **M0 — Foundation:** CLI skeleton, YAML config, local target parsing, Git
  inspection, Ollama connectivity check, text/JSON renderers, `tldg doctor`.
- **M1 — Local summary MVP:** safe file traversal, doc/manifest extraction,
  language detection, SQLite FTS index, structured repository profile, and a
  local-model `tldg summary` with file citations.

**Outcome:** `tldg doctor` reports environment health, `tldg summary . --offline`
emits a structured profile without a model, and `tldg summary .` produces a
cited purpose/architecture/entry-points/deps/tests/caveats overview via the
local Ollama model — all offline-capable.

**Environment (verified):** Go 1.26.4 (darwin/arm64), git 2.50, ollama 0.30
with `qwen2.5-coder:7b` (synthesis) + `nomic-embed-text` (embeddings, M2),
sqlite 3.51.

**Locked decisions:** CLI = spf13/cobra · Git = CLI shell-out behind a
`GitBackend` interface · scope = M0 + M1 together.

---

## Key technical choices

| Concern | Choice | Rationale |
| --- | --- | --- |
| Module path | `github.com/leslierussell/tldg` | Adjust if a different remote is intended. |
| CLI | `spf13/cobra` | Spec-aligned; idiomatic for a growing command tree. |
| Git | Shell out to system `git` behind `GitBackend` interface | Spec §26 — fidelity first; `go-git` deferrable. |
| YAML | `gopkg.in/yaml.v3` | Spec §19.1. |
| SQLite + FTS5 | `modernc.org/sqlite` (pure Go, no cgo) | Single static binary, cross-platform; FTS5 supported. Interface-wrapped so a cgo/vector backend can swap in later. |
| Model client | Ollama (`/api/tags`, `/api/chat`) + OpenAI-compatible (`/v1/models`, `/v1/chat/completions`) behind a `Provider` interface | Spec §10.1. |
| Vector/embeddings | **Deferred to M2** | M1 summary uses file selection + FTS only, not semantic retrieval. |

---

## Package layout (subset of spec §19 needed for M0/M1)

```
cmd/tldg/main.go             — entrypoint, wires cli.Execute()

internal/
  cli/       cobra commands: root (global flags), doctor, summary, config, version
  app/       orchestration: DoctorRun, SummaryRun pipelines
  config/    YAML load, OS/XDG path resolution, defaults, validation
  target/    parse local path → Target; resolve enclosing worktree
  git/        GitBackend interface + gitcli impl (root, branch, HEAD, remotes, tags, dirty)
  repo/       RepositoryProfile domain model
  scan/       traversal, ignore rules, file selection (spec §9.3)
  language/   ecosystem/language detection from manifests + extensions (spec §9.4)
  docs/       doc/manifest/build-file extraction + classification (spec §9.5)
  index/      SQLite metadata + FTS5 lexical index behind a Store interface
  evidence/   Evidence record, citation-ID build + validation (spec §8.3, §11.2)
  models/     Provider interface, ollama client, openai-compat client
  render/     text, markdown, JSON renderers (spec §8)
  version/    build/version info
```

`pkg/`, `hosts/`, `research/`, `retrieval/`, `agents/`, `secrets/`, `auth/`,
`policy/`, `execsafe/`, `cache/`, `telemetry/` are **out of scope** until later
milestones (stub only where a global flag references them, e.g. `--offline`).

---

## M0 — Foundation

**Deliverables & mapping:**

1. **Module + skeleton** — `git init`, `go mod init`, `cmd/tldg/main.go`,
   cobra root in `internal/cli/root.go` with global flags from spec §7.2 that
   M0/M1 honor: `--config`, `--profile`, `--model`, `--offline`, `--json`,
   `--markdown`, `--quiet`, `--verbose`, `--no-index`. Flags for unimplemented
   subsystems (`--web`, `--sources`, `--policy`, `--allow-external-code`) are
   parsed but return a clear "not available until milestone N" error if used.
2. **Config** (`internal/config`) — load `config.yaml`; resolve paths via
   `os.UserConfigDir`/`os.UserCacheDir` + a data dir per spec §15.1; apply
   defaults matching the spec §15.2 example (models, embeddings, index, output,
   privacy, logging). `tldg config path|show|validate`.
   Validation (spec §15.3 subset): YAML syntax + `version`, model endpoint
   syntax, index path writability, invalid ignore patterns, unsupported source
   types, offline-vs-enabled-source contradiction.
3. **Target parsing** (`internal/target`) — parse a local path into a `Target`;
   find enclosing Git worktree; allow non-Git dir in reduced mode. (Remote /
   provider / qualified-ref forms are recognized and produce a clear
   "supported in milestone 4" error — parsing scaffold only.)
4. **Git inspection** (`internal/git`) — `GitBackend` interface + `gitcli` impl:
   worktree root, current branch / detached-HEAD, HEAD hash + timestamp,
   remotes (normalized URLs), nearest tags, dirty state (no file contents in
   logs). Spec §9.2.
5. **Model connectivity** (`internal/models`) — `Provider` interface;
   `ollama` client with `Ping` (GET `/api/tags`) and model-presence check;
   `openai-compatible` client stub with `Ping` (GET `/v1/models`).
6. **Renderers** (`internal/render`) — text and stable versioned JSON
   (`schema_version: "1.0"`, spec §8.2). Markdown renderer minimal.
7. **`tldg doctor`** (`internal/app` + `internal/cli/doctor`) — checks (spec
   §21.1 subset): git present+version, config valid + active paths,
   data/cache/index dirs writable, ollama reachable + configured model present,
   embedding model reachable, SQLite open/integrity. Keychain/host/plugin
   checks deferred (report as "n/a — milestone 4").

**M0 acceptance:**
```sh
tldg doctor            # green/red checks, non-zero exit if a hard dep missing
tldg summary . --offline   # structured profile, no model call (see M1)
```

---

## M1 — Local summary MVP

Builds the `summary` pipeline. `--offline` runs steps 1–6 and renders the
profile with a "model synthesis skipped (offline)" note (satisfies M0
acceptance too); online adds steps 7–10.

**Summary pipeline** (`internal/app/summary.go`):

1. Resolve target + Git identity (M0 pieces).
2. **Scan** (`internal/scan`) — traverse honoring `.gitignore`/`.ignore` +
   config `ignore` list; exclude `.git`, dep/build dirs, binaries, oversized
   files (`index.max_file_size_kb`); prioritize README/docs/manifests/CI/
   entrypoints/tests per spec §9.3. Respect `--no-index`.
3. **Docs/manifests** (`internal/docs`) — extract + classify README\*,
   CONTRIBUTING/SECURITY/LICENSE/CHANGELOG, `docs/**`, ADRs; keep
   documentation-claim vs code-verified distinction (spec §9.5, §4.1).
4. **Language detection** (`internal/language`) — from manifests
   (`go.mod`, `package.json`, `Cargo.toml`, `pyproject.toml`, …) + extensions;
   may report polyglot (spec §9.4). Extract direct manifest deps for the deps
   section (spec §9.8, manifest-level only in M1).
5. **FTS index** (`internal/index`) — SQLite metadata + FTS5 over selected
   file chunks (doc-section / fixed-size fallback; full Tree-sitter chunking is
   M2). Manifest records repo identity, revision, tldg version, config hash
   (spec §11.7). Skip when `--no-index`.
6. **Repository profile** (`internal/repo`) — assemble structured
   `RepositoryProfile`: identity, languages, entry-point candidates, deps,
   test/CI signals, doc index, health basics (last commit/tag dates).
7. **Grounding bundle** (`internal/evidence`) — ranked, deduped, size-budgeted
   evidence snippets with stable IDs (spec §10.4); budget by
   `retrieval.max_chunks` / `max_chunk_tokens`.
8. **Synthesis** (`internal/models`) — system prompt: repo text is untrusted
   data, cite evidence, state uncertainty (spec §10.3, §4.5); produce the 10
   standard summary sections (spec §7.3 `summary`).
9. **Citation validation** (`internal/evidence`) — every `[local:path:a-b]` in
   output must resolve to a bundle evidence URI; flag unknown ones (spec §10.5,
   §17.4).
10. **Render** — text (answer-first + trailing evidence list + confidence
    note), `--json` (claims/evidence/timing), `--markdown`.

**`summary` options** (spec §7.3): `--depth brief|standard|architecture|exhaustive`
(M1: brief/standard fully; deeper depths widen file budget), `--include`/`--exclude`
areas, `--branch`, `--ref`.

**M1 acceptance:**
```sh
tldg summary .   # cited purpose, architecture, entry points, deps, tests, caveats
```

---

## Reused / referenced building blocks

- Evidence record + kinds — spec §11.2 / §11.3 (implement `local_file`,
  `git_commit`, `git_tag`, `generated_analysis` for M0/M1).
- Citation formats — spec §8.3 (`[local:…]`, `[git:…]`).
- JSON output contract — spec §8.2 (build the versioned envelope once in
  `render`, reused by all commands).
- Default config values — spec §15.2 (encode as Go defaults so a missing config
  file still works).
- Exit codes — spec §7.4 (map pipeline failures to codes 0–9).

---

## Testing (spec §22.1 subset)

Unit tests alongside each package:
- `target`: local path + worktree resolution; remote-form recognition errors.
- `config`: parse, defaults merge, validation cases (§15.3).
- `scan`: ignore rules, file selection, size/binary exclusion.
- `language`: single + polyglot detection fixtures.
- `evidence`: citation construction + validation (valid/invalid/unknown ID).
- `git`: parse `git` output via a fake backend (no live repo needed).
- `render`: JSON schema shape + stable field ordering.

Add a tiny **fixture repo** under `testdata/` (small Go CLI: `go.mod`, `main.go`,
`README.md`) for a `summary --offline` golden-ish integration test that asserts
profile structure without depending on model output.

---

## Execution workflow (per user CLAUDE.md — beads)

This project mandates the **beads (bd)** workflow with zero code before a bead.
During implementation:
1. Create beads for each work unit (suggested: `m0-skeleton`, `m0-config`,
   `m0-target`, `m0-git`, `m0-models`, `m0-render`, `m0-doctor`, `m1-scan`,
   `m1-docs`, `m1-language`, `m1-index`, `m1-profile`, `m1-summary-pipeline`,
   `m1-citations`, `m1-tests`).
2. Claim → branch named exactly the bead ID → implement → tests pass → commit →
   merge to master → delete branch → close bead.
3. Tag each added code block with a `// <bead-id>` comment.

---

## Verification (end-to-end)

```sh
go build ./... && go vet ./... && go test ./...

./tldg doctor                 # all local checks pass; exit 0
./tldg doctor --json          # stable JSON envelope
./tldg config validate        # OK on default config
./tldg summary . --offline    # structured profile, "synthesis skipped" note
./tldg summary .              # cited overview via qwen2.5-coder:7b
./tldg summary . --json | jq .schema_version   # "1.0"
./tldg summary ./testdata/gofixture   # integration fixture
```

Manual checks: citations in output resolve to real files/lines; `--offline`
makes zero network calls; missing Ollama degrades gracefully (doctor red,
`summary --offline` still works) per spec §20.
