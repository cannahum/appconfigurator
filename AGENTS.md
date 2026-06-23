# AGENTS.md — appconfigurator

Knowledge harness for humans and AI agents working in this repo. Captures the
mental model, conventions, and the non-obvious bits that bite you at a glance.

## What this is

A tiny, deliberately-simple Go config loader. It does three things:

1. **Layered config** — load `default.json`, deep-merge a single `<env>.json` on
   top, unmarshal into your typed struct `T`.
2. **Reference resolution** — fields tagged `resolve:"..."` hold *references*
   (an env var name, a file path), which are resolved to real values at load.
3. That's it. Keep it that way.

```go
cfg, err := appconfigurator.Load[MyConfig](dir, "production")
// cfg.Variables is your *resolved* MyConfig
```

## Why it exists

`kelseyhightower/envconfig` (used elsewhere in these services) solves a
*different* problem: **environment variables → struct**, with per-field literal
`default:"…"` values, `required`, and prefixes. It has no notion of config files,
a `default`→`<env>` hierarchy, deep merging, or environment selection.

appconfigurator is **layered files → struct**: load `default.json`, deep-merge one
`<env>.json`, unmarshal into `T`, then resolve `env`/`file` *references*. That
file-layering is the gap it fills.

Overlap to be aware of: `resolve:"env"` reads env vars too — but the variable
*name* lives in the JSON (an indirection envconfig doesn't have). You could
approximate the whole thing as "a JSON loader + envconfig for overrides," but
you'd lose the single deep-merge model and the `file` resolver.

The design value is **simplicity** — keep it thin; a new feature must justify
itself against "does it stay this simple?" Don't turn it into viper.

## Where configs live

`Load(dir, env)` takes the directory explicitly. For the common case use
`DefaultConfigurationDirectoryPath()`, which:

- starts at the current working directory,
- **walks *up* to the nearest `go.mod`** (treated as project root),
- returns **`<project-root>/configuration`** — note: **singular `configuration`**.

So configs live in `<project-root>/configuration/{default,local,staging,...}.json`.
(If a project uses a different folder, it must pass the path to `Load` itself.)

## The environment model

- Files: `default.json` + exactly one `<env>.json` (e.g. `local`, `development`,
  `staging`, `production`). Env name is the engineer's choice.
- **The hierarchy is flat: `default` + one env.** There is no `prod`-overrides-`dev`
  chain — every env file overrides `default` and nothing else.
- `default.json` = shared/base structure. `<env>.json` = only the values that
  differ for that environment.
- **Empty env string → `default.json` only**, with a `[WARN]` printed. Forgetting
  to set the environment silently gives you defaults.

## Non-obvious behaviors (read before you trip)

- **Deep-merge is map-only.** Nested objects merge recursively by key; **arrays
  and scalars in `<env>.json` replace the default wholesale.** You cannot *append*
  to a default array from an env file — it's overwrite-or-inherit. (Map-of-objects
  like `consumers` *do* merge per key.)
- **Config round-trips through JSON**: file → `map[string]any` → re-marshal →
  unmarshal into `T`. So a custom field type must survive JSON marshalling (see
  `File`), and `T` must match the merged JSON shape.
- **Resolution runs inside `Load`, after unmarshal.** Callers never see
  references — only resolved values.

## Reference resolution (`resolve` tags)

Values you don't want written literally in config files are passed as
**references** and resolved at load. The field **type encodes intent**:

| field type | tag | behavior |
|---|---|---|
| `string` | *(none)* | literal value, untouched (e.g. a path you keep but don't read) |
| `string` | `resolve:"env"` | value is an **env var name** → replaced with the variable's value (empty if unset) |
| `File`   | `resolve:"file"` | value is a **path** → file is read into memory (`File.Data`) |

Two resolvers exist today (`env`, `file`), both **stdlib-only** — keeping the
library dependency-light is deliberate. Planned, not yet implemented:
`creds_file` (file contents → string) and `secret` (AWS Secrets Manager) — note
`secret` is the resolver that will introduce a non-stdlib dependency (the AWS
SDK), so weigh that before adding it. Empty references are left untouched
(supports values omitted in some envs). An unknown resolver, a type mismatch, or
a missing file **errors at load**.

### Example

`configuration/production.json`:

```json
{
  "temporal": {
    "namespace": "my-ns.my-acct",
    "api_key": "TEMPORAL_API_KEY",
    "client_cert": "/run/secrets/temporal/cert.pem"
  }
}
```

Go:

```go
type Temporal struct {
    Namespace  string                `json:"namespace"`
    APIKey     string                `json:"api_key"     resolve:"env"`
    ClientCert appconfigurator.File  `json:"client_cert" resolve:"file"`
}
```

After `Load` (with `TEMPORAL_API_KEY=sk-abc123` in the environment):

```
Temporal.Namespace        == "my-ns.my-acct"            // literal
Temporal.APIKey           == "sk-abc123"                 // env VALUE, not the name
Temporal.ClientCert.Path  == "/run/secrets/temporal/cert.pem"
Temporal.ClientCert.Data  == <file bytes>                // read into memory
```

### Secrets rule

**Secrets never live in config JSON — only references do.** In staging/production,
the referenced env vars / files are provisioned from a secret store (e.g. AWS
Secrets Manager, injected as env vars or mounted files); locally they're typically
unset/omitted. A field that would otherwise hold a literal secret should be a
`resolve:"env"` (or `file`) reference, never a plaintext value.

## The `File` type

`File{ Path string; Data []byte }`. JSON-marshals to/from its path string; the
`file` resolver fills `Data`. Use the value type (`File`, not `*File`) for
`resolve:"file"` fields.

## Conventions

- Keep it simple — new features must justify themselves against that.
- Resolvers are a small fixed set coded directly into the package. The expected
  universe is only a handful (`env`, `file`, `creds_file`, `secret`), so there is
  **no plugin/injection machinery by design** — add a new `case` to the resolver,
  don't build an abstraction.
- Tests: table-/case-style, in-package (`package appconfigurator`), matching the
  existing files. No special framework requirements.
