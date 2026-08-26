# tldg — Full Product and Technical Specification

**Name:** `tldg`  
**Expansion:** *too long didn’t git*  
**Tagline:** Local-first, evidence-grounded repository intelligence from the terminal.  
**Status:** Design specification  
**Primary implementation language:** Go  
**Primary interface:** Command-line interface  
**Default AI posture:** Local models; external research is explicit, policy-controlled, and source-cited.

---

## 1. Executive summary

`tldg` is a terminal application for understanding, interrogating, and researching software repositories. It accepts a local working directory, a Git remote, a repository URL, or a provider-qualified repository reference; extracts a structured view of the repository; and uses a locally hosted language model to answer questions from verifiable evidence.

It is not only “chat with a codebase.” Its core output is a defensible repository model: what the project does, how it is structured, how it executes, which dependencies it relies on, how it evolved, how active it is, and what external sources indicate about usage, issues, and maintenance.

`tldg` is designed around five principles:

1. **Local-first:** Repository reading, indexing, embeddings, planning, and answer synthesis run locally by default.
2. **Evidence before prose:** The model receives retrieved evidence, not unrestricted access to invent claims from a repository dump.
3. **Provider-agnostic:** GitHub, GitLab, Bitbucket, Gitea, Forgejo, Codeberg, self-hosted hosts, and generic Git remotes are supported through adapters.
4. **Explicit external access:** Web and social/developer-source research are opt-in or policy-gated, minimally disclosing, cached, and cited.
5. **Terminal-native:** Fast, scriptable, composable, keyboard-first, and usable as a single binary.

---

## 2. Goals

### 2.1 Primary goals

`tldg` MUST:

- Produce a high-quality summary of a repository’s actual purpose and implementation.
- Answer natural-language questions about local or remote repositories with precise citations.
- Inspect source code, project documentation, manifests, test suites, CI definitions, Git history, and hosting-provider metadata.
- Resolve local repository remotes to known hosting providers when possible.
- Support GitHub, GitLab, Bitbucket, Forgejo/Gitea-compatible hosts, and generic Git remotes by default.
- Allow users to add hosts and tune behavior through YAML configuration.
- Use local LLM and embedding runtimes by default, including Ollama and OpenAI-compatible endpoints such as LM Studio.
- Support optional online research through pluggable sources, including general web search, developer-centric sources, package registries, and X.com where credentials and applicable access are available.
- Clearly separate direct evidence, inferred conclusions, external commentary, and uncertainty.
- Protect private source code, secrets, and local metadata from accidental external disclosure.
- Work well on macOS, Linux, and Windows where Git is available.

### 2.2 Secondary goals

`tldg` SHOULD:

- Run entirely offline for local repository analysis.
- Be useful as both an interactive assistant and a scriptable JSON-producing CLI.
- Allow remote analysis without forcing a full clone when provider APIs are sufficient.
- Cache indexes and provider/web retrieval results safely.
- Be extensible by Go plugins or external executable plugins.
- Support a rich terminal UI later without making it required for core use.

### 2.3 Non-goals for the first stable release

`tldg` is NOT initially intended to:

- Replace a full IDE, code browser, Git client, issue tracker, or hosted code-search product.
- Autonomously edit code, create pull requests, close issues, or make provider-side changes.
- Execute arbitrary repository code by default.
- Send source code to a hosted model provider without clear user opt-in.
- Claim formal correctness, complete security auditing, or legal compliance certification.
- Treat social-media commentary as authoritative project documentation.

---

## 3. User personas

### 3.1 Local developer

A developer opens an unfamiliar repository and wants a concise but accurate explanation of architecture, entry points, dependencies, and extension points.

Example:

```sh
tldg summary .
tldg ask . "Where does the HTTP request enter the application?"
```

### 3.2 Open-source evaluator

A user wants to evaluate whether a project is healthy, actively maintained, compatible with their stack, and affected by known ecosystem concerns.

Example:

```sh
tldg investigate github:owner/project \
  "Is this project actively maintained, and what are its operational risks?" \
  --web
```

### 3.3 Maintainer or contributor

A maintainer wants to understand code ownership, architectural history, open problem clusters, release evolution, and likely impact of a proposed change.

Example:

```sh
tldg history . "Why was the storage layer redesigned?"
tldg impact . "What would be affected by replacing the authentication provider?"
```

### 3.4 Self-hosting or enterprise user

A user works with private repositories and self-hosted GitLab, Forgejo, Gitea, or another Git service. They need local processing and custom-host configuration.

Example:

```sh
tldg host add git.example.internal --kind gitlab
tldg summary git.example.internal:platform/ledger
```

---

## 4. Product principles

### 4.1 Evidence-grounded output

Every material factual claim MUST carry one or more references unless the user explicitly selects an ungrounded creative mode. Citations MUST identify the evidence origin:

- Local file path and line range.
- Git object, commit hash, tag, or branch.
- Provider object such as issue, pull request, release, or discussion.
- External URL and retrieval timestamp.
- Derived analysis label, when the claim is computed rather than directly stated.

### 4.2 Confidence and uncertainty

Answers MUST distinguish among:

- **Verified:** Directly supported by source or primary project metadata.
- **Strong inference:** Consistent with multiple evidence sources but not explicitly stated.
- **Tentative inference:** Plausible but incomplete or conflicting evidence exists.
- **Unknown:** Evidence is insufficient.

The model MUST NOT conceal uncertainty merely to make an answer more fluent.

### 4.3 Privacy by default

`tldg` MUST not transmit local code, file contents, local paths, environment variables, Git configuration, credentials, or internal URLs to online services unless the user explicitly enables a policy that permits it.

### 4.4 Least-privilege operations

`tldg` SHOULD default to read-only access. Provider tokens SHOULD use the narrowest available scopes. Credential material MUST be stored outside normal YAML config, preferably in OS-native secure credential storage.

### 4.5 Repository text is untrusted data

README files, source comments, issues, pull requests, commit messages, CI logs, package metadata, web pages, and social-media posts MUST be treated as untrusted content. They MUST NOT override system instructions, policies, user intent, or tool permission boundaries.

---

## 5. Terminology

| Term | Meaning |
| --- | --- |
| Target | A local directory, remote URL, provider URL, or provider-qualified repository reference supplied to `tldg`. |
| Repository identity | Canonical host, owner/namespace, repository name, default branch, and remote URL information. |
| Evidence | A retrieved, normalized, source-attributed content item used for a response. |
| Index | Persisted lexical, structural, semantic, and metadata representation of repository content. |
| Provider | A version-control hosting platform such as GitHub or GitLab. |
| Host adapter | An implementation that knows how to resolve and query a particular provider or self-hosted host. |
| Research source | An external source adapter such as web search, package registry, Hacker News, documentation search, or X.com. |
| Grounding bundle | The ranked and compressed evidence provided to the model for a question. |
| Local model | An LLM served on the user machine or trusted local network. |
| External research | A retrieval operation contacting a public or configured remote service. |

---

## 6. Supported targets

`tldg` MUST accept the following target formats.

### 6.1 Local paths

```sh
tldg summary .
tldg summary ~/src/example
tldg ask ../another-project "What is the main executable?"
```

Behavior:

- Find the enclosing Git worktree when possible.
- Allow analysis of a non-Git directory in reduced mode.
- Resolve remotes from `.git/config` when available.
- Respect `.gitignore`, `.ignore`, and configurable exclusion rules while indexing.

### 6.2 Git remote URLs

```sh
tldg summary git@github.com:owner/project.git
tldg summary https://gitlab.com/group/project.git
tldg summary ssh://git@git.example.internal/team/project.git
```

Behavior:

- Normalize URL forms.
- Select a host adapter when configured or recognized.
- Clone only when explicitly requested or necessary for required analysis.
- Permit metadata-only remote analysis where provider APIs provide enough data.

### 6.3 Provider URLs

```sh
tldg summary https://github.com/neovim/neovim
tldg ask https://gitlab.com/gitlab-org/gitlab "How are background jobs organized?"
```

### 6.4 Qualified repository references

```sh
tldg summary github:neovim/neovim
tldg summary gitlab:group/project
tldg summary codeberg:forgejo/forgejo
tldg summary git.example.internal:team/service
```

Host aliases map to configured host definitions.

---

## 7. CLI specification

### 7.1 Global syntax

```text
tldg [global flags] <command> [target] [arguments]
```

### 7.2 Global flags

| Flag | Purpose |
| --- | --- |
| `--config <path>` | Use a non-default configuration file. |
| `--profile <name>` | Select a named configuration profile. |
| `--model <name>` | Override the active LLM configuration. |
| `--offline` | Forbid all network activity. |
| `--web` | Permit configured external research for this invocation. |
| `--sources <list>` | Limit enabled evidence sources, e.g. `local,github,web`. |
| `--refresh` | Bypass relevant caches and retrieve fresh data. |
| `--no-index` | Analyze without persisting an index. |
| `--json` | Emit stable machine-readable JSON. |
| `--markdown` | Emit Markdown rather than terminal-oriented text. |
| `--quiet` | Suppress progress output. |
| `--verbose` | Show retrieval, planner, cache, and source diagnostics. |
| `--yes` | Accept non-sensitive prompts automatically; MUST NOT bypass credential or privacy policy. |
| `--policy <name>` | Select a privacy/execution policy profile. |

### 7.3 Commands

#### `tldg summary`

Produce a repository overview.

```sh
tldg summary .
tldg summary github:owner/project --depth architecture
tldg summary . --format json
```

Options:

- `--depth <brief|standard|architecture|exhaustive>`
- `--include <areas>` such as `docs,code,history,activity,deps`
- `--exclude <areas>`
- `--branch <name>`
- `--ref <commit-or-tag>`

Expected standard sections:

1. Purpose.
2. Primary users and use cases.
3. Runtime and deployment model.
4. Architecture and major subsystems.
5. Entry points and critical execution flows.
6. Dependencies and integrations.
7. Development workflow and testing.
8. Repository health and activity.
9. Risks, unknowns, and caveats.
10. Evidence list.

#### `tldg ask`

Answer a repository-grounded question.

```sh
tldg ask . "How is configuration loaded?"
tldg ask github:owner/project "Is OAuth supported?" --web
tldg ask . "Which package owns retry behavior?" --json
```

Options:

- `--scope <local|provider|web|all>`
- `--explain-retrieval`
- `--max-evidence <n>`
- `--follow-links`
- `--allow-external-code <never|metadata|selected|all>`

#### `tldg investigate`

Perform a multi-step evidence-gathering investigation.

```sh
tldg investigate . "Why did build time increase after v2?"
tldg investigate github:owner/project \
  "What are the biggest adoption barriers reported in the past year?" \
  --web --sources github,web,hn,x
```

This command MAY generate a plan, retrieve sequentially, compare sources, and return a research report. It MUST show external queries in verbose mode and cite each conclusion.

#### `tldg history`

Analyze repository evolution using Git history and provider activity.

```sh
tldg history . "Why was the cache added?"
tldg history . --between v1.0.0..v2.0.0
tldg history . --file internal/auth/service.go
```

#### `tldg impact`

Estimate change impact from symbols, import/dependency relationships, tests, configuration, and history.

```sh
tldg impact . "Replace the database driver"
tldg impact . --symbol AuthenticateUser
tldg impact . --file cmd/server/main.go
```

#### `tldg sources`

Show available, used, cached, and disabled sources.

```sh
tldg sources .
tldg sources github:owner/project --json
tldg sources status
```

#### `tldg index`

Manage persistent repository indexes.

```sh
tldg index build .
tldg index update .
tldg index status .
tldg index remove .
tldg index gc
```

#### `tldg host`

Manage configured repository hosts.

```sh
tldg host list
tldg host add git.example.internal --kind gitlab
tldg host test git.example.internal
tldg host remove git.example.internal
```

#### `tldg auth`

Manage host and research-source credentials.

```sh
tldg auth login github.com
tldg auth status
tldg auth logout gitlab.com
```

#### `tldg config`

Inspect and manage configuration.

```sh
tldg config path
tldg config show
tldg config validate
tldg config edit
```

#### `tldg doctor`

Diagnose Git, model, index, credential, host, and network configuration.

```sh
tldg doctor
tldg doctor --model
tldg doctor --host github.com
```

### 7.4 Exit codes

| Exit code | Meaning |
| --- | --- |
| 0 | Successful completion. |
| 1 | General command failure. |
| 2 | Invalid CLI arguments or configuration. |
| 3 | Target cannot be resolved. |
| 4 | Authentication or authorization failure. |
| 5 | Network failure or rate limit prevented required retrieval. |
| 6 | Local model or embedding service unavailable. |
| 7 | Privacy/execution policy denied an operation. |
| 8 | Index corruption or storage failure. |
| 9 | Partial result; output was produced but required evidence retrieval failed. |

---

## 8. Output contract

### 8.1 Human-readable output

Default terminal output MUST include:

- A direct answer or summary before supporting detail.
- Citations placed near the supported claim.
- A compact evidence section at the end.
- Explicit indication when online sources were used.
- A source-quality and confidence note when applicable.

Example:

```text
Purpose

This repository implements a self-hosted event-processing service. Its HTTP
server is initialized in cmd/server, workflows are handled by internal/engine,
and persistent state is backed by PostgreSQL. [local:cmd/server/main.go:31-97]
[local:internal/engine/engine.go:18-166] [local:go.mod]

Confidence: verified for server architecture; tentative for intended scale.

Caveat

The README describes Kubernetes deployment, but the current CI and Helm chart
were not present at the selected revision. [local:README.md:72-126]
```

### 8.2 JSON schema

All `--json` answers MUST be machine-readable and versioned.

```json
{
  "schema_version": "1.0",
  "command": "ask",
  "target": {
    "kind": "local",
    "path": "/redacted/project",
    "repository": {
      "host": "github.com",
      "owner": "owner",
      "name": "project",
      "ref": "main"
    }
  },
  "answer": {
    "text": "...",
    "confidence": "strong_inference",
    "external_research_used": false
  },
  "claims": [
    {
      "text": "The CLI entry point is cmd/tldg/main.go.",
      "confidence": "verified",
      "evidence_ids": ["ev_001"]
    }
  ],
  "evidence": [
    {
      "id": "ev_001",
      "kind": "local_file",
      "uri": "file:cmd/tldg/main.go",
      "ref": "HEAD",
      "line_start": 1,
      "line_end": 114,
      "content_hash": "sha256:...",
      "retrieved_at": "2026-08-26T13:56:00Z"
    }
  ],
  "warnings": [],
  "timing": {
    "index_ms": 0,
    "retrieval_ms": 324,
    "model_ms": 1701
  }
}
```

### 8.3 Citation formats

Human output MUST use recognizable URI-like citation identifiers:

- `[local:README.md:10-44]`
- `[git:abc1234]`
- `[git:tag:v2.1.0]`
- `[github:issue:123]`
- `[gitlab:mr:group/project!44]`
- `[web:https://example.org/article]`
- `[registry:npm:package@1.4.0]`

Citations MUST resolve through `tldg sources` or JSON evidence records.

---

## 9. Repository analysis system

### 9.1 Analysis phases

The analysis pipeline MUST have distinct phases:

1. Target resolution.
2. Repository identity collection.
3. File selection and safe traversal.
4. Documentation extraction.
5. Build, package, language, and runtime detection.
6. Structural code extraction.
7. Git history extraction.
8. Provider metadata retrieval, if permitted.
9. External research retrieval, if permitted.
10. Evidence normalization, deduplication, ranking, and compression.
11. Local model synthesis.
12. Citation validation and output rendering.

### 9.2 Repository identity

For a Git working tree, tldg SHOULD collect:

- Worktree root.
- Current branch and detached-HEAD state.
- HEAD commit hash and commit timestamp.
- Configured remotes and normalized URLs.
- Default upstream branch when discoverable.
- Dirty working-tree state, excluding file contents from logs unless requested.
- Tags nearest to HEAD.
- Repository format and submodule declarations.

### 9.3 File selection

`tldg` MUST avoid blindly ingesting every file. The file selector SHOULD prioritize:

- README and project documentation.
- Package manifests and lockfiles.
- Build scripts and task definitions.
- CI/CD definitions.
- Container and deployment definitions.
- Top-level application entry points.
- Public APIs, exported symbols, routes, commands, schemas, and configuration.
- Tests and examples.
- Architecture decision records and changelogs.

By default, it SHOULD exclude:

- `.git/`
- dependency directories such as `node_modules/`, `vendor/`, `.venv/`
- build products such as `dist/`, `build/`, `target/`, `coverage/`
- binary files
- oversized files
- generated files when detected
- user-configured ignore patterns

It MUST permit explicit inclusion and exclusion overrides.

### 9.4 Language and ecosystem detection

The system SHOULD detect common language ecosystems through manifests, conventional files, and source extensions.

Initial priority:

- Go: `go.mod`, `go.work`.
- Rust: `Cargo.toml`, workspace manifests.
- JavaScript/TypeScript: `package.json`, lockfiles, `tsconfig.json`.
- Python: `pyproject.toml`, `requirements*.txt`, `setup.py`.
- Zig: `build.zig`, `build.zig.zon`.
- Java/Kotlin: Gradle and Maven manifests.
- C/C++: CMake, Meson, Makefiles, Bazel.
- Ruby, PHP, C#, Elixir, Haskell, Lua, and shell projects as subsequent built-ins.

Detection MUST be evidence-driven and MAY identify a polyglot repository.

### 9.5 Documentation extraction

Extract, classify, and prioritize:

- `README*`
- `CONTRIBUTING*`
- `SECURITY*`
- `LICENSE*`
- `CHANGELOG*`
- `CODE_OF_CONDUCT*`
- `docs/**`
- `adr/**`, `docs/adr/**`, `decisions/**`
- examples and tutorials
- generated API docs only when configured

Documentation claims MUST be distinguishable from implementation-verified claims.

### 9.6 Structural extraction

`tldg` SHOULD use Tree-sitter where parsers exist and fall back to lightweight language-aware heuristics otherwise.

Extractable structural entities include:

- Files and directories.
- Modules/packages/namespaces.
- Imports and dependency edges.
- Functions, methods, classes, traits, interfaces, structs, enums, and types.
- Exported/public symbols.
- CLI commands, flags, routes, RPC endpoints, handlers, jobs, and consumers.
- Configuration keys and environment variables.
- Test suites and fixture relationships.
- Database migrations and schema definitions.

Each entity MUST retain source location and revision metadata.

### 9.7 Execution-flow discovery

`tldg` SHOULD identify likely execution paths through a combination of:

- Manifest-defined executables.
- Framework conventions.
- `main` functions and equivalent entry points.
- `package.json` scripts and bin declarations.
- Docker `ENTRYPOINT` and `CMD`.
- CI command invocations.
- Makefile, Taskfile, Justfile, Nix, Bazel, and build configuration.
- Route/handler registration and dependency injection patterns.

Execution flow output SHOULD distinguish static certainty from framework-convention inference.

### 9.8 Dependency analysis

Dependency analysis MUST include direct manifest dependencies. It SHOULD include lockfile-resolved versions where practical.

It SHOULD classify dependencies by role:

- Runtime framework.
- Database/client.
- Network/protocol.
- Authentication/security.
- Logging/observability.
- Testing.
- Build/tooling.
- UI.
- Generated-code toolchain.

A dependency claim MUST cite its manifest, lockfile, import use, or provider/registry source.

### 9.9 Git history analysis

`tldg` MUST support Git-native evidence without depending on a hosting provider.

It SHOULD collect:

- Recent commits and affected files.
- Tags and release boundaries.
- File and symbol history.
- Rename detection where Git provides it.
- High-churn modules.
- Co-change relationships.
- Contributor concentration, without inferring identity or competence.
- Commit-message topic clusters.

History interpretation MUST avoid asserting motivations not evidenced by commit messages, linked issues, pull requests, or documentation.

### 9.10 Repository health model

A health report MAY include:

- Most recent commit/release/activity dates.
- Open vs closed issue/PR patterns when a host adapter provides data.
- Release cadence.
- Documentation and test signals.
- CI configuration presence.
- Security policy and advisory signals.
- Bus-factor indicators presented carefully as descriptive, not deterministic.

It MUST NOT collapse these into an authoritative single score by default. If a score is enabled, the input factors and uncertainty MUST be shown.

---

## 10. Local AI system

### 10.1 Supported model providers

Built-in providers MUST include:

- Ollama.
- OpenAI-compatible HTTP endpoints.

The OpenAI-compatible provider enables local runtimes such as LM Studio, llama.cpp servers, vLLM, LocalAI, and compatible gateways when configured by the user.

Potential future providers:

- Apple Foundation Models, where platform APIs permit.
- Direct llama.cpp embedding/inference integration.
- MLX-backed local servers.

### 10.2 Model roles

`tldg` SHOULD support separately configured models for:

- Answer synthesis.
- Research planning.
- Embeddings.
- Optional query rewriting.
- Optional reranking.

A minimal configuration MAY use one model for all generative roles and one embedding model.

### 10.3 Model interaction rules

The model MUST:

- Receive a system policy that repository and web content are untrusted data.
- Be restricted to declared read-only tools unless the user explicitly enables an execution policy.
- Be instructed to cite evidence for material claims.
- Be instructed to state when evidence is missing or conflicting.
- Never receive secrets detected by the redaction layer.
- Not decide independently to perform online research when policy requires confirmation.

### 10.4 Context construction

The grounding bundle SHOULD contain:

- User question.
- Target identity and selected revision.
- Repository profile summary.
- Ranked source snippets with stable evidence IDs.
- Tool summaries where full content is unnecessary.
- Explicit source classes and trust levels.
- Hard budget limits by tokens and source count.

The context builder MUST favor diverse, non-duplicative evidence over many redundant chunks.

### 10.5 Answer validation

Before output, tldg SHOULD validate:

- Each citation identifier exists.
- Each cited evidence item supports a nearby claim where detectable.
- No source text was interpreted as tool instructions.
- The final response does not include redacted content.
- Claims labeled verified have direct supporting evidence.

---

## 11. Retrieval and index design

### 11.1 Storage model

The default local store SHOULD use SQLite for durable metadata and full-text search, plus a local vector index implementation.

Recommended logical layout:

```text
~/.local/share/tldg/
├── tldg.db
├── indexes/
│   └── <repository-id>/
│       ├── lexical.db
│       ├── vectors.hnsw
│       └── manifest.json
├── cache/
│   ├── providers/
│   └── research/
└── logs/
```

Platform-specific config/data directories MUST follow OS conventions where possible.

### 11.2 Evidence record

Every extracted or retrieved item MUST normalize into an evidence record.

```go
type Evidence struct {
    ID            string
    RepositoryID  string
    Kind          EvidenceKind
    Source        string
    URI           string
    Ref           string
    Title         string
    Content       string
    ContentHash   string
    LineStart     *int
    LineEnd       *int
    PublishedAt   *time.Time
    RetrievedAt   time.Time
    Author        *string
    TrustClass    TrustClass
    Visibility    Visibility
    Metadata      map[string]string
}
```

### 11.3 Evidence kinds

Initial evidence kinds:

- `local_file`
- `git_commit`
- `git_tag`
- `git_blame`
- `provider_repo`
- `provider_issue`
- `provider_pull_request`
- `provider_discussion`
- `provider_release`
- `provider_security_advisory`
- `package_registry`
- `web_page`
- `web_search_result`
- `social_post`
- `generated_analysis`

### 11.4 Chunking

Chunking MUST preserve meaningful boundaries.

Preferred order:

1. Markdown/document sections.
2. Tree-sitter declarations and enclosing symbol blocks.
3. Configuration blocks.
4. Test cases.
5. Fixed-size fallback chunks with overlap.

Each chunk MUST retain parent file path, line range, content hash, language, symbol metadata where available, and revision.

### 11.5 Search modes

`tldg` SHOULD combine:

- Exact/lexical retrieval for identifiers, errors, filenames, flags, and API names.
- Structural retrieval for symbols and dependency edges.
- Semantic retrieval for conceptual questions.
- History retrieval for “why,” “when,” “who changed,” and regression questions.
- Provider retrieval for issue/PR/release questions.
- External retrieval only under the selected policy.

A hybrid ranker SHOULD combine lexical score, embedding similarity, structural proximity, document importance, freshness, and source trust.

### 11.6 Incremental indexing

`tldg index update` SHOULD:

- Identify revision changes and uncommitted changes.
- Reprocess only affected files/chunks when feasible.
- Invalidate dependent structural and embedding records.
- Preserve immutable evidence from previous revisions when explicitly indexed.

### 11.7 Index reproducibility

Index manifests MUST record:

- Repository identity.
- Indexed Git revision.
- Working-tree state indicator.
- tldg version.
- Parser versions.
- embedding model identifier.
- chunking configuration hash.
- timestamp.

---

## 12. Provider adapter architecture

### 12.1 Adapter interface

```go
type HostAdapter interface {
    Kind() string
    MatchRemote(remoteURL string) bool
    NormalizeRemote(remoteURL string) (RepoRef, error)
    Resolve(ctx context.Context, input string) (RepoRef, error)
    Capabilities(ctx context.Context, host HostConfig) (CapabilitySet, error)
    GetRepository(ctx context.Context, ref RepoRef) ([]Evidence, error)
    GetRepositoryTree(ctx context.Context, ref RepoRef, opts TreeOptions) ([]Evidence, error)
    GetFile(ctx context.Context, ref RepoRef, path string, revision string) (Evidence, error)
    GetCommits(ctx context.Context, ref RepoRef, opts CommitOptions) ([]Evidence, error)
    SearchIssues(ctx context.Context, ref RepoRef, q string, opts SearchOptions) ([]Evidence, error)
    SearchPullRequests(ctx context.Context, ref RepoRef, q string, opts SearchOptions) ([]Evidence, error)
    GetReleases(ctx context.Context, ref RepoRef, opts ReleaseOptions) ([]Evidence, error)
    RateLimit(ctx context.Context) (RateLimitInfo, error)
}
```

Methods MAY return capability-not-supported errors. Caller behavior MUST degrade gracefully.

### 12.2 Built-in host kinds

| Kind | Default aliases | API model |
| --- | --- | --- |
| `github` | `github`, `github.com` | GitHub REST and GraphQL APIs |
| `gitlab` | `gitlab`, `gitlab.com` | GitLab REST API |
| `bitbucket` | `bitbucket`, `bitbucket.org` | Bitbucket Cloud REST API |
| `forgejo` | `forgejo`, `codeberg` | Forgejo/Gitea-compatible REST API |
| `gitea` | `gitea` | Gitea-compatible REST API |
| `generic` | user-defined | Git remote only, optional web templates |

GitLab exposes project and repository REST resources, including repository tree/file access and resources such as issues and merge requests, making it suitable for a full adapter. [web:1][web:5][web:7][web:9]

Forgejo’s API is published through Swagger/OpenAPI and supports paginated access, enabling a reusable Forgejo/Gitea-style adapter for Forgejo-compatible hosts. [web:2][web:10]

### 12.3 GitHub behavior

The GitHub adapter SHOULD use REST for straightforward retrieval and GraphQL where it substantially reduces request count or improves relationship retrieval.

It MUST:

- Respect API rate-limit responses and headers.
- Cache immutable data by commit/release revision.
- Use conditional requests where available.
- Support unauthenticated public metadata in degraded mode.
- Support token-backed access for private repositories where user-configured.

GitHub’s REST API imposes materially lower unauthenticated limits than authenticated access, so `tldg doctor` SHOULD warn users when unauthenticated retrieval is likely to be constrained. [web:3]

### 12.4 Generic-host behavior

A `generic` host MUST support:

- Remote normalization.
- Git clone/fetch under explicit policy.
- Local Git analysis after clone.
- Optional user-configured URL templates for repository home, issues, pull requests, and raw files.

It MUST NOT assume a provider API or scrape arbitrary HTML by default.

### 12.5 Custom host YAML

```yaml
hosts:
  git.example.internal:
    kind: gitlab
    alias: corp
    api_url: https://git.example.internal/api/v4
    web_url: https://git.example.internal
    token_keychain_service: tldg/git.example.internal
    default_namespace: platform
    tls:
      verify: true

  forge.example.net:
    kind: forgejo
    alias: forge
    api_url: https://forge.example.net/api/v1
    web_url: https://forge.example.net
    token_keychain_service: tldg/forge.example.net

  source.example.org:
    kind: generic
    alias: source
    web_url: https://source.example.org
    remote_patterns:
      - "git@source.example.org:{owner}/{repo}.git"
      - "https://source.example.org/{owner}/{repo}.git"
```

---

## 13. External research system

### 13.1 Design goals

External research supplements repository-local evidence. It MUST NOT silently replace direct source analysis.

It SHOULD answer questions such as:

- Is a project still maintained?
- What release, migration, or incident context exists outside the repository?
- What installation failures or adoption barriers are repeatedly reported?
- Which adjacent tools, standards, or packages are relevant?
- What have maintainers publicly said about roadmap or breaking changes?

### 13.2 Source interface

```go
type ResearchSource interface {
    Name() string
    Capabilities() SourceCapabilities
    IsEnabled(profile PolicyProfile) bool
    Search(ctx context.Context, q ResearchQuery) ([]Evidence, error)
    Fetch(ctx context.Context, uri string) (Evidence, error)
}
```

### 13.3 Built-in research sources

Initial sources SHOULD include:

- General web search.
- Official project documentation domains.
- GitHub provider search and discussions where available.
- Package registries such as npm, PyPI, crates.io, Go proxy/pkg.go.dev metadata, Maven Central, and RubyGems as ecosystem support expands.
- Developer community sources with stable accessible interfaces.
- X.com as an optional, token-backed source.

### 13.4 Source trust model

Each result MUST receive a trust classification based on source type, provenance, repository linkage, and freshness.

| Trust class | Examples | Intended use |
| --- | --- | --- |
| Primary | Source code, official docs, official release, maintainer post | Strong support for factual claims |
| Operational | Issue, PR, CI log, registry metadata | Operational and maintenance evidence |
| Secondary | Reputable technical article, conference talk | Context and interpretation |
| Community | Forum discussion, Q&A, social post | User experience and leads |
| Unverified | Unattributed post, weakly linked content | Discovery only unless corroborated |

### 13.5 X.com support

X.com integration MUST be disabled by default. When enabled, it MUST:

- Use approved API access or a user-supplied compliant retrieval provider.
- Store credentials only in secure credential storage.
- Retrieve only content necessary for the stated question.
- Record author, post URL, timestamps, query, retrieval timestamp, and repository-link confidence.
- Treat posts as community or primary evidence only when authorship can be established.
- Never present engagement metrics as technical evidence.

### 13.6 Query minimization

By default, external queries MAY include:

- Public repository name.
- Public owner/organization name.
- Public package name.
- A short question-derived keyword phrase.

They MUST NOT include by default:

- Local source code snippets.
- Unpublished branch names.
- Private repository names.
- Internal hostnames or file paths.
- Environment values, secrets, stack traces containing sensitive data, or user identity details.

### 13.7 User consent behavior

Default policy:

- Local commands: no network required.
- Provider metadata for an explicit public remote target: permitted if configured.
- External web research: prompt once per command unless `--web`, an allowed profile, or a matching stored policy permits it.
- Sending any source-code content externally: deny unless a stricter opt-in flag and policy allow it.

Example prompt:

```text
External research is disabled by your current policy.
Allow tldg to search configured public sources using this query?
  "owner project database migration issues"
No local code or file paths will be sent. [y/N]
```

### 13.8 Caching and freshness

Research evidence MUST record retrieval time and cache expiry. Default TTLs:

| Evidence type | Default TTL |
| --- | --- |
| Repository metadata | 1 hour |
| Issue/PR listings | 15 minutes |
| Release metadata | 6 hours |
| Package registry metadata | 6 hours |
| Search result list | 1 hour |
| Fetched web page | 24 hours |
| X/social result | 1 hour |

Users MUST be able to bypass cache with `--refresh` and disable persistent external caching in policy.

---

## 14. Privacy and security specification

### 14.1 Data classification

`tldg` MUST classify content before model context or network use:

- Public remote metadata.
- Private remote metadata.
- Local repository source.
- Local documentation.
- Local untracked file.
- Generated artifact.
- Credential/secret candidate.
- External public content.

### 14.2 Secret detection and redaction

Before indexing, model prompting, logging, output, or external operations, tldg SHOULD scan for likely secrets using:

- Entropy heuristics.
- Known token prefixes.
- Assignment-key patterns.
- Private-key block detection.
- User-configurable regular expressions.
- Common `.env`, credential, and key-file paths.

Defaults MUST exclude known sensitive file patterns, including:

```text
.env
.env.*
*.pem
*.key
id_rsa
id_ed25519
credentials*
secrets*
**/.aws/**
**/.ssh/**
```

Detection is a safety measure, not a guarantee. The user MUST be able to inspect redaction decisions locally.

### 14.3 Execution safety

`tldg` MUST NOT run repository code, install dependencies, invoke build systems, start containers, or execute test commands by default.

If execution is later enabled, it MUST require an explicit action such as:

```sh
tldg run . -- command "go test ./..."
tldg investigate . "Diagnose test failure" --allow-exec=sandbox
```

Execution policies SHOULD include:

- `deny`: no execution.
- `prompt`: explicit confirmation per command.
- `sandbox`: restricted container/sandbox execution.
- `local`: user-permitted local process execution.

### 14.4 Prompt injection resistance

The agent layer MUST:

- Delimit every retrieved source as untrusted content.
- Reject instructions embedded in source content that attempt to alter system behavior or access controls.
- Keep tool permissions outside model-controlled text.
- Require application-layer approval for side effects or external source disclosure.
- Log detected injection indicators in diagnostic mode without repeating harmful payloads unnecessarily.

### 14.5 Credential storage

Credentials MUST NOT be stored in plaintext YAML by default.

Preferred backend order:

- macOS Keychain.
- Windows Credential Manager.
- Linux Secret Service/keyring.
- Encrypted fallback file only when user explicitly chooses it.

`tldg auth` MUST expose credential status without printing secret values.

### 14.6 Telemetry

Telemetry MUST be off by default. If ever implemented, opt-in telemetry MUST avoid repository names, paths, code, prompts, tokens, and URLs by default.

---

## 15. Configuration specification

### 15.1 Default paths

| OS | Configuration | Data | Cache |
| --- | --- | --- | --- |
| macOS | `~/Library/Application Support/tldg/config.yaml` or XDG-compatible path | `~/Library/Application Support/tldg/` | `~/Library/Caches/tldg/` |
| Linux | `~/.config/tldg/config.yaml` | `~/.local/share/tldg/` | `~/.cache/tldg/` |
| Windows | `%AppData%\tldg\config.yaml` | `%LocalAppData%\tldg\` | `%LocalAppData%\tldg\cache\` |

`tldg config path` MUST report active paths.

### 15.2 Complete example configuration

```yaml
version: 1

profile: default

profiles:
  default:
    model: local-coder
    policy: local-first

  research:
    model: local-coder
    policy: public-research

models:
  local-coder:
    provider: ollama
    endpoint: http://127.0.0.1:11434
    model: qwen3-coder:30b
    context_window: 32768
    temperature: 0.15
    max_output_tokens: 4096

  local-compat:
    provider: openai-compatible
    endpoint: http://127.0.0.1:1234/v1
    model: local-model
    api_key_keychain_service: tldg/lm-studio

embeddings:
  default:
    provider: ollama
    endpoint: http://127.0.0.1:11434
    model: nomic-embed-text
    dimensions: 768

hosts:
  github.com:
    kind: github
    alias: github
    api_url: https://api.github.com
    web_url: https://github.com
    token_keychain_service: tldg/github.com

  gitlab.com:
    kind: gitlab
    alias: gitlab
    api_url: https://gitlab.com/api/v4
    web_url: https://gitlab.com
    token_keychain_service: tldg/gitlab.com

  bitbucket.org:
    kind: bitbucket
    alias: bitbucket
    api_url: https://api.bitbucket.org/2.0
    web_url: https://bitbucket.org
    token_keychain_service: tldg/bitbucket.org

  codeberg.org:
    kind: forgejo
    alias: codeberg
    api_url: https://codeberg.org/api/v1
    web_url: https://codeberg.org
    token_keychain_service: tldg/codeberg.org

  git.example.internal:
    kind: gitlab
    alias: corp
    api_url: https://git.example.internal/api/v4
    web_url: https://git.example.internal
    token_keychain_service: tldg/git.example.internal
    tls:
      verify: true

sources:
  web:
    enabled: true
    max_results: 10
    allowed_domains: []
    blocked_domains: []

  package_registries:
    enabled: true
    sources:
      - npm
      - pypi
      - crates
      - pkg-go-dev

  developer_communities:
    enabled: true
    sources:
      - hackernews
      - reddit

  x:
    enabled: false
    mode: api
    token_keychain_service: tldg/x.com
    max_results: 20

index:
  root: ~/.local/share/tldg/indexes
  sqlite_path: ~/.local/share/tldg/tldg.db
  max_file_size_kb: 512
  max_total_index_size_mb: 2048
  include_untracked: false
  follow_symlinks: false
  ignore:
    - .git
    - node_modules
    - vendor
    - dist
    - build
    - target
    - coverage
    - .venv
  include: []

retrieval:
  max_chunks: 24
  max_chunk_tokens: 1200
  lexical_weight: 0.45
  semantic_weight: 0.35
  structural_weight: 0.20
  reranker:
    enabled: false

privacy:
  external_research_default: ask
  external_code_default: never
  persistent_external_cache: true
  local_path_display: relative
  redact_patterns:
    - '(?i)api[_-]?key\\s*[:=]\\s*\\S+'
    - '(?i)secret\\s*[:=]\\s*\\S+'
    - '(?i)password\\s*[:=]\\s*\\S+'

execution:
  default: deny
  sandbox_engine: docker
  network_default: disabled

output:
  citations: inline
  confidence: true
  color: auto
  source_paths: relative

logging:
  level: info
  persist_prompts: false
  persist_model_responses: false
```

### 15.3 Config validation

`tldg config validate` MUST check:

- YAML syntax and schema version.
- Duplicate aliases and host conflicts.
- Required URLs by host kind.
- Model endpoint syntax.
- Index path write permissions.
- Invalid ignore patterns.
- Unsupported source types.
- Insecure credential declarations.
- Policy contradictions, such as enabled external sources under a fully offline profile.

---

## 16. Policy profiles

### 16.1 Built-in policies

| Policy | Local analysis | Provider metadata | Web research | Send source externally | Execution |
| --- | --- | --- | --- | --- | --- |
| `offline` | Allowed | Denied | Denied | Denied | Denied |
| `local-first` | Allowed | Prompt or configured public host | Prompt | Denied | Denied |
| `public-research` | Allowed | Allowed for public targets | Allowed | Metadata only | Denied |
| `private-safe` | Allowed | Authenticated configured hosts | Denied by default | Denied | Denied |
| `sandboxed-analysis` | Allowed | As configured | Prompt | Denied | Sandboxed with prompt |

### 16.2 Policy enforcement

Policies MUST be enforced by application code, not prompt text. A model MUST not be able to weaken or replace the policy.

---

## 17. Research and answer behavior

### 17.1 Question classification

`tldg ask` SHOULD classify questions into retrieval strategies:

| Question type | Typical retrieval |
| --- | --- |
| Purpose or overview | README, docs, manifests, entry points, high-level symbols |
| How does X work? | Symbols, call paths, config, tests, docs |
| Where is X implemented? | Lexical and structural search |
| Why was X changed? | Git history, linked issues/PRs, commit messages |
| Is X supported? | Docs, config, tests, releases, issue status |
| Is project healthy? | Commit/release/activity metadata, CI, security policy, open issue patterns |
| What do users say? | Issues, discussions, external research, social sources |
| What breaks if I change X? | Dependency graph, callers, tests, config, history |

### 17.2 Investigation planning

For `investigate`, the system MAY create an internal plan such as:

```text
1. Establish selected revision and repository purpose.
2. Locate relevant subsystem and public interfaces.
3. Review change history around named concepts.
4. Search provider issues and releases.
5. Search approved external sources for corroborating reports.
6. Compare findings and report evidence quality.
```

The plan MUST remain bounded by configurable request, token, time, and retrieval budgets.

### 17.3 Handling disagreement

When direct sources disagree, the answer MUST state the conflict.

Example:

```text
The README says SQLite is intended only for development, while integration tests
exercise SQLite as a supported backend. The repository does not provide a clear
current production-support statement. [local:README.md:93-101]
[local:tests/sqlite_integration_test.go:1-148]
```

### 17.4 Hallucination controls

The system MUST prefer “I could not verify this” over unsupported completion.

It SHOULD flag answer fragments that:

- Mention a file or symbol absent from the evidence index.
- Cite an evidence item unrelated to the claim.
- Assert runtime behavior from only a documentation claim.
- Infer maintainer intent without associated primary evidence.

---

## 18. Plugin system

### 18.1 Scope

Plugins allow extension of:

- Host adapters.
- Research sources.
- Language analyzers.
- Manifest analyzers.
- Output renderers.
- Optional local model providers.

### 18.2 Recommended mechanism

Use external executable plugins communicating over JSON-RPC or a versioned stdio protocol. Avoid Go’s in-process `plugin` package as the primary public extension boundary because of compatibility and deployment limitations.

Example plugin discovery locations:

```text
~/.local/share/tldg/plugins/
~/.config/tldg/plugins/
$PATH entries named tldg-*
```

Example plugin declaration:

```yaml
plugins:
  - name: gerrit
    command: tldg-host-gerrit
    enabled: true

  - name: sourcegraph
    command: tldg-source-sourcegraph
    enabled: false
```

### 18.3 Plugin permissions

Plugins MUST declare capabilities:

- Network access.
- Credential access by named service only.
- Local repository read access.
- Index read/write access.
- External process execution.

The host application MUST allow users to deny capabilities by policy.

---

## 19. Internal package design

Suggested Go module layout:

```text
cmd/
  tldg/
    main.go

internal/
  app/             command orchestration
  cli/             cobra/urfave style command definitions
  config/          loading, validation, profiles
  policy/          privacy and capability enforcement
  target/          parsing and target resolution
  git/             worktree, objects, remotes, history
  repo/            repository domain model
  scan/            filesystem traversal and file selection
  language/        Tree-sitter and language analysis
  manifests/       ecosystem manifest readers
  docs/            markdown/document extraction
  index/           lexical, vector, SQLite metadata
  evidence/        normalization, provenance, citations
  retrieval/       hybrid search, ranking, context building
  models/          LLM and embedding provider clients
  agents/          planner and answer synthesis orchestration
  hosts/           host adapter registry
  research/        external research source registry
  secrets/         redaction and secret detection
  auth/            keychain and credential management
  cache/           TTL and conditional-request cache
  execsafe/        optional sandboxed execution
  render/          text, markdown, JSON output
  telemetry/       disabled-by-default diagnostics

pkg/
  protocol/        versioned plugin and external API types
```

### 19.1 Dependency preferences

Use well-supported, lightweight dependencies where possible:

- Go standard library for networking, JSON, filesystem, process handling.
- `go-git` or Git CLI integration depending on fidelity/performance requirements; Git CLI may be preferable for complete behavior initially.
- Tree-sitter bindings for language parsing.
- SQLite driver with FTS5 support.
- OS keychain library.
- YAML v3 parser.
- A small HNSW/vector library or SQLite-vector extension behind an interface.

All external dependencies MUST be pinned and license-reviewed.

---

## 20. Error handling and degraded operation

`tldg` MUST return useful partial answers when non-critical sources fail.

Examples:

- If Ollama is unavailable, allow `tldg index`, `tldg sources`, and `tldg doctor` to work; report how to restore model connectivity.
- If a provider token is missing, analyze cloned/local content and state provider data was unavailable.
- If web research is blocked, answer from local/provider evidence and indicate the omitted source class.
- If a file is too large or binary, record it as skipped with a reason.
- If an index is stale, identify the indexed revision and suggest `tldg index update`.

Example warning:

```text
Partial result: GitHub issue retrieval was rate-limited. The maintenance
assessment uses local Git history and release metadata cached 42 minutes ago.
Run with --refresh after authentication for current issue data.
```

---

## 21. Observability and diagnostics

### 21.1 `tldg doctor`

Checks SHOULD include:

- Git installation and version.
- Local model reachability and model availability.
- Embedding model reachability.
- Config schema validity.
- Writable data/cache/index locations.
- Keychain availability.
- Host adapter connectivity and authentication status.
- Network-policy state.
- Database integrity.
- Plugin compatibility.

### 21.2 Verbose diagnostics

`--verbose` SHOULD show:

- Target normalization.
- Selected repository ref.
- Index state and freshness.
- Retrieval source counts.
- Cache hits/misses.
- External query strings after redaction.
- Rate-limit status.
- Timing by pipeline phase.

It MUST NOT print credentials, source contents marked sensitive, or raw model prompts containing secrets.

---

## 22. Testing strategy

### 22.1 Unit tests

Cover:

- URL and remote parsing.
- Host adapter normalization.
- Config parsing/validation.
- File selection and ignore rules.
- Secret redaction.
- Citation construction.
- Chunking and source location preservation.
- Policy enforcement.
- Cache expiry.
- JSON schema generation.

### 22.2 Integration tests

Use fixture repositories across languages and layouts:

- Small Go CLI.
- Rust workspace.
- TypeScript monorepo.
- Python package with docs/tests.
- Polyglot service with Docker and CI.
- Git submodule fixture.
- Repository with generated/vendor content.
- Repository containing intentionally malicious prompt-injection text.
- Repository containing fake secret patterns and actual test-only fixtures.

### 22.3 Contract tests

Each host adapter MUST have contract tests against mocked API responses and optional live smoke tests controlled by environment variables.

### 22.4 Evaluation corpus

Create a curated evaluation suite with known questions and source-grounded expected answers:

- Purpose questions.
- Entry-point questions.
- Dependency questions.
- History/motivation questions.
- Support-status questions.
- Ambiguous/conflicting-documentation questions.
- Prompt-injection resistance tests.

Metrics:

- Citation validity.
- Citation entailment quality.
- Answer completeness.
- Unsupported-claim rate.
- Retrieval recall.
- Time to first useful answer.
- Local resource usage.

---

## 23. Performance targets

Initial targets for a modern developer laptop:

| Operation | Target |
| --- | --- |
| Resolve local target | under 250 ms |
| Brief summary on already-indexed small/medium repo | under 10 seconds excluding model variability |
| Incremental index update after small change | under 5 seconds |
| Cold index of 10k source files | under 3 minutes, hardware/model dependent |
| Lexical query latency | under 300 ms |
| Retrieval-only question preparation | under 2 seconds |
| Remote metadata cache hit | under 500 ms |

All targets are best-effort and MUST be surfaced as diagnostics, not guarantees.

---

## 24. Release milestones

### Milestone 0 — Foundation

Deliver:

- Go CLI skeleton.
- YAML config loading and validation.
- Target parsing for local repositories.
- Git worktree/remote/ref inspection.
- Ollama model connectivity check.
- Text/JSON renderers.
- `tldg doctor`.

Acceptance:

```sh
tldg doctor
tldg summary . --offline
```

### Milestone 1 — Local summary MVP

Deliver:

- Safe file traversal.
- README/docs/manifests/build-file extraction.
- Basic language detection.
- FTS index.
- Structured repository profile.
- Local-model summary with file citations.

Acceptance:

```sh
tldg summary .
```

Produces purpose, architecture, entry points, dependencies, tests, and caveats with citations.

### Milestone 2 — Grounded local Q&A

Deliver:

- Code chunking and symbol extraction.
- Hybrid lexical/semantic retrieval.
- `tldg ask`.
- Citation validation.
- Persistent incremental indexes.

Acceptance:

```sh
tldg ask . "How does configuration flow into the server?"
```

### Milestone 3 — Git intelligence

Deliver:

- Commit/tag/file history retrieval.
- `tldg history`.
- Change-impact prototype.
- Revision-selectable analysis.

### Milestone 4 — Hosted repository support

Deliver:

- GitHub adapter.
- Keychain-backed authentication.
- GitLab and Forgejo/Gitea adapters.
- Custom host YAML.
- Caching/rate-limit support.

Acceptance:

```sh
tldg summary github:owner/project
tldg summary corp:platform/service
```

### Milestone 5 — External research

Deliver:

- Policy/consent layer.
- Web search adapter.
- Package registry adapter.
- Evidence trust classification.
- `tldg investigate`.
- External citation and cache UX.

### Milestone 6 — Advanced research and plugins

Deliver:

- Plugin protocol.
- Developer-community adapters.
- Optional X.com adapter.
- Sandboxed execution experiment.
- TUI exploration.

---

## 25. Example workflows

### 25.1 Understand an unfamiliar local project

```sh
cd ~/src/unfamiliar-service
tldg summary . --depth architecture
```

Expected behavior:

- Reads Git identity and selected revision.
- Extracts README, docs, manifests, CI, container files, entry points, and major symbols.
- Generates an evidence-backed overview using the local model.
- Does not perform network requests.

### 25.2 Trace a feature

```sh
tldg ask . "Trace a user login request from HTTP endpoint to session creation."
```

Expected behavior:

- Finds routes/handlers.
- Traces relevant calls, configuration, persistence, and tests.
- Answers in sequence with cited files/symbols.
- Labels inaccessible dynamic behavior as inferred rather than verified.

### 25.3 Research an open-source dependency

```sh
tldg investigate github:owner/project \
  "Is the project appropriate for a self-hosted production deployment?" \
  --web --sources github,web,package_registries
```

Expected behavior:

- Retrieves repository metadata, releases, documentation, relevant issues, and registry metadata.
- Searches approved web sources with minimized public queries.
- Separates project claims, operational evidence, and community reports.
- Does not send local code.

### 25.4 Analyze a self-hosted GitLab project

```sh
tldg host add git.example.internal --kind gitlab
tldg auth login git.example.internal
tldg summary corp:platform/ledger
```

Expected behavior:

- Uses the custom configured GitLab API endpoint.
- Retrieves only scopes available to the user token.
- Keeps private contents local unless explicitly permitted otherwise.

### 25.5 Find historical cause

```sh
tldg history . "Why did the project switch from Redis to PostgreSQL for queues?"
```

Expected behavior:

- Searches commit messages, diffs, tags, linked issues and PRs if available.
- Reports direct evidence first.
- Clearly says when the motivation is inferred from code changes rather than documented.

---

## 26. Open technical decisions

These require prototype validation rather than premature lock-in:

1. **Git access:** Git CLI versus `go-git` for object/history operations. Start with the Git CLI for fidelity and broad behavior, abstract it behind a `GitBackend` interface.
2. **Vector storage:** Standalone HNSW file versus SQLite extension. Start with an interface and a portable HNSW implementation if packaging SQLite extensions becomes burdensome.
3. **Tree-sitter packaging:** Static compilation, dynamic grammar loading, or per-language optional modules. Favor static support for the initial language set.
4. **TUI:** Start CLI-only. A Bubble Tea TUI can be added after core evidence and interaction contracts stabilize.
5. **Plugin sandboxing:** Begin with external executable plugins and explicit permissions; stronger OS sandboxing can follow.
6. **Web-source vendors:** Keep search behind an adapter so the user can select a provider or self-hosted search endpoint.
7. **Remote clone cache:** Determine whether to maintain bare mirror clones for generic-host deep analysis. Make clone behavior explicit and size-bounded.

---

## 27. Definition of done for v1.0

`tldg` v1.0 is complete when it can:

- Analyze a local Git repository entirely offline.
- Produce an accurate, cited summary covering purpose, structure, entry points, dependencies, tests, and caveats.
- Answer grounded questions over code and documentation with stable file/line citations.
- Analyze Git history for repository-local “why did this change?” questions.
- Resolve and inspect GitHub, GitLab, Bitbucket, and Forgejo/Gitea repositories using configured host adapters.
- Add self-hosted or custom hosts via YAML configuration.
- Use local Ollama and OpenAI-compatible model endpoints.
- Store tokens securely through supported OS credential mechanisms.
- Perform opt-in external research without uploading code by default.
- Clearly label source types, confidence, uncertainty, and partial failures.
- Resist common prompt-injection attempts in repository content.
- Produce stable JSON output suitable for shell scripts and downstream tooling.

---

## 28. Project identity

`tldg` should present itself as:

> **tldg — too long didn’t git**
>
> A local-first terminal tool that explains what repositories actually do, answers questions from code and history, and—when you allow it—investigates the wider developer ecosystem with evidence you can inspect.

The name is playful, memorable, and accurately frames the core promise: a repository may be too large, too old, too underdocumented, or too unfamiliar to absorb manually—but it is not too large to investigate methodically.
