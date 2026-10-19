<!-- webtyp:base — copied from devskills skill agents-md; do not edit here -->
## Ecosystem rules — read before any change

`docs/PLAN.md` is the spec: no design judgment; anything unclear is a question in the PR, not a guess.

- Public API: typed (no `any`/`interface{}`), explicit, minimal; never export a symbol only a test uses.
- One intent, one path: the old path is deleted in the same change. No silent fallback, `TODO`, commented code or stub.
- Fail at compile time, else a loud diagnostic — never silently. Zero values deny.
- A defect in a dependency is reported, never wrapped or patched here.
- No string literals in logic: named constants. The same logic never lives in two repos.
- `cmd/*/main.go`: parse flags, inject, print/exit only. stdout = data, stderr = logs, exit codes are the contract.
- WASM code: `webtyp/fmt`, `webtyp/time`, `webtyp/json` instead of stdlib; no `map`; `dom.Element` embedded by value.
- `css.go`, `svg.go`, `js.go`, `html.go` carry `//go:build !wasm`; no `ssr.go`, no `front.go`.
- Code and keys in English; UI text is `lang.Text`, translations live in `lang.json`.
- Tests cover behaviour and domain rules, never the look (CSS, colours, markup order, class strings). In `tests/` (external package), real stack with fakes only at the edge; run `gotest`.
- Docs (`README.md`, `docs/*`) updated with the API; permanent docs never link `docs/PLAN.md`.
- Before the PR: `git diff --stat main` lists every file of the plan's stages, and nothing the plan did not order is deleted.
<!-- /webtyp:base -->

## This repo

- **A concrete channel for `webtyp.com/notify`.** It implements `notify.Driver` (kind `"email"`) and the `notify.Channel` it builds; it adds no concept of its own. Which alert goes to whom is `alert_manager`'s job, not this repo's.
- It is a **concrete implementation of an outside service** (Cloudflare Email Service), so it may use `net/http`; it is server-side code and is never imported by a browser (WASM) build: `//go:build !wasm` on every non-test file.
- **Secrets are never in the database, the repo or a log.** The configuration names a secret (`notify.Config.Secrets`); the value is read through the injected `notify.SecretResolver`. Never print a token, never put it in an error message.
- **Nothing fails silently:** a rejected or failed send is returned as an error from `Send`; there is no best-effort path.
- Cloudflare Email Sending is **beta**: its REST API is wrapped behind this one package so that a change of service or of method (REST, SMTP) touches only this repo.
- Keep `docs/PRICING.md` accurate: it is the document the owner reads to know what this costs.
