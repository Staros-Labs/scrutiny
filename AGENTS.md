# Scrutiny Builder Instructions

<!-- staros-agents-baseline: Staros-Labs/infra AGENTS.md -->

This repository adopts the Staros organization agent baseline by reference.
Rules below cover Scrutiny-specific workflow and validation.

## Working Rules

- Read this file and `CONTRIBUTING.md` before making changes.
- Sync the main checkout with `origin`, then work in an isolated git worktree.
- Preserve user changes and existing linked worktrees. Never reset, restore, or
  clean work you did not create.
- Keep this maintained fork easy to compare with `AnalogJ/scrutiny`; avoid
  unrelated rewrites of upstream-owned code.
- Never commit API keys, tokens, credentials, runtime config, DB files, or local
  agent/session state.
- Do not use emojis or assistant-product attribution in commits, PRs, comments,
  or docs.

## Branch Model

- `master` is production and release source.
- `develop` is normal feature and fix integration source.
- Normal implementation bootstrap must pass `--base origin/develop`.
- Use `--base origin/master` only for explicit hotfix, release, or workflow
  enrollment work.
- Shared Linear branches use `<repo>/<issue-key>-<slug>`; PR target still follows
  this branch model.

## Repository Shape

- `webapp/backend/`: Go API and web server.
- `webapp/frontend/`: Angular frontend.
- `collector/`: SMART, ZFS, mdadm, Btrfs, and performance collectors.
- `.github/workflows/`: CI, security, image, testing, and release automation.
- `docs/`: maintained fork and operator documentation.

## Validation

- Backend: `go test ./...`
- Focused backend: `go test -v ./path/to/pkg/...`
- Frontend: `cd webapp/frontend && npm test -- --watch=false`
- Use focused checks for documentation or workflow-only changes.
- Never run deployment or production mutations without explicit approval.

<!-- STAROS_LINEAR_GUARD_CONTRACT_START -->
## Shared agent execution contract

- Pure read-only inspection, audit, project research, and approved read-verb MCP calls may run before Linear bootstrap, an issue, or a worktree. Research includes source acquisition and review, documentation, architecture, dependency manifests, history, issues, releases, licenses, existing CodeGraph index queries, and usefulness recommendations.
- Repo execution, dependency installation, project-controlled scripts, tracked changes, tests, non-host builds, lifecycle updates, and unknown or mutating MCP tools require the repository's Linear bootstrap flow and a live worktree lease.
- Every actual execution workstream requires a durable Linear child under its primary issue, including work in primary repo; use `--workstream 'REPO|KEY|TITLE|DESCRIPTION'` or `workstream ensure`, and never reuse an unrelated repo issue or lease.
- Primary issue scope must declare every authorized repo with `--execution-repos`; repair existing scope with `scripts/linear-work-item.sh scope ensure --issue KEY --primary-repo REPO --execution-repos CSV` before bootstrap.
- Secondary sessions use `--primary-issue`, `--primary-repo`, and matching workstream fields; child parent, repo, and scope identity must match before lease claim.
- Title-only compliance renames may use `scripts/linear-work-item.sh rename --issue KEY --title "..." --reason "..."` from clean main; the helper validates team ownership and posts an audit comment.
- Direct Linear CLI writes and arbitrary MCP writes remain lease-gated.
- Known read-only connector variants may run pre-lease, including GitHub fetch tools, `mcp__codegraph__codegraph_explore`, `web__run`, and `request_user_input`; unknown or mutation variants remain gated.
- Pathless connector mutations require exact session or thread identity mapped to one live worktree and matching target repository; missing, stale, ambiguous, and mismatched context blocks.
- Exact `git fetch origin` is allowed from any checkout; `git pull --ff-only` requires a clean checkout. One bounded HTTPS clone may create a new absolute research target outside every existing repository, guarded operator path, and managed worktree root; SSH, credential-bearing, relative, existing, nested, or submodule-recursive clone forms remain gated.
- Approved read-only inspection, synchronization, startup, and recovery commands may run before a lease.
- Project research reads include common local probes, safe filesystem discovery, expanded Git history and tree inspection, existing-index CodeGraph CLI queries, GitHub searches, and REST GET-only `gh api` requests. Exact `command -v <name>` and `command -V <name>` lookup remains approved. Newline-separated read commands use the same bounded all-segments-safe rule as semicolon chains. Direct local and known-fleet Docker, `systemctl`, and `launchctl` commands run pre-lease, including lifecycle, build, pull, update, and restart forms. Other unclassified shell commands remain lease-gated and are reported as not approved read-only.
- Exact pre-lease coordination checks are also approved: `scripts/install-mcp-env.sh --check`, `scripts/sync-agent-guard-runtime.sh --check`, `scripts/sync-operator-config.sh --check`, `scripts/sync-codex-mcp.sh --check` or `scripts/sync-codex-mcp.sh --check --quiet --best-effort`, `scripts/sync-claude-agent-mcp.sh --check`, and `scripts/agent-lease.sh status`. Apply, install, heartbeat, and unknown argument forms remain lease-gated.
- Exact normal and Git-only enrollment wrapper calls may run only from clean infra main against one clean, enrolled target main checkout. The parser rejects unknown flags, relative or duplicate targets, shell control, and substitution before the wrapper can create its guarded worktree.
- Chrome and Computer Use browser actions bypass repo lease protection by policy; browser confirmation rules still apply. Browser-based GitHub, Linear, deploy, and production actions therefore require care because they do not establish or check a repo lease.
- Legacy non-browser GitHub close, comment, and update actions require an exact single-use receipt with verified Linear mapping, GitHub linkage, target, operator, and TTL; receipt replay and mismatch block.
- Bounded `&&`, `||`, and semicolon chains are allowed only when every segment is independently approved read-only; one unsafe or unknown segment rejects whole chain.
- All Agent Context MCP tools, including `context_promote`, `context_update`, and `context_resync`, are available pre-lease. Context writes and reporting stay in curated local memory or configured reporting paths and cannot edit tracked repository files.
- Tracked edits, commits, and pushes require a live worktree lease. During a confirmed Linear outage, existing tasks, new tasks, emergency documentation, branch commits, and branch pushes remain allowed in isolated leased worktrees.
- Degraded new-task branches use `DEGRADED-<id>-slug` and bind `LINEAR_REMOTE_ISSUE` after reconciliation; existing issue branches retain their real issue key.
- Degraded lifecycle events enter the mode-600 local outbox. `scripts/linear-outbox.sh reconcile` runs one drain pass; the launchd worker runs every five minutes. Rate limits pause the shared drain, permanent errors block the event, and reconciliation never merges, releases, deploys, or mutates production automatically.
- When a Linear cooldown blocks startup or reconciliation, run `scripts/linear-cooldown.sh status` and preserve state. Do not clear it or repeat provider calls. After `blocked_until`, resume an existing marker with `scripts/start-linear-session.sh --recover-degraded --issue KEY`; it performs one probe and rewrites degraded state only after success. Then run `scripts/linear-outbox.sh reconcile --repo REPO --once`.
- Merge, release, deploy-wrapper, and production HTTP mutations remain receipt-gated during degradation. Direct local and known-fleet host lifecycle commands remain outside lease and receipt gating. Receipts are single-use, exact-target, operator-bound, mode-600 files with a default 15-minute expiry and a 60-minute maximum; environment-variable bypasses are forbidden.
- Use branch format `<repo>/<ISSUEKEY>-<number>-<slug>` for normal sessions and `DEGRADED-<id>-slug` for new degraded sessions.
- Active leases expire after 15 minutes without heartbeat and become parked reservations retained for 30 days; parked sessions never authorize mutation.
- Use `scripts/agent-lease.sh park` and `resume` for pauses. Use `request-takeover` and `accept-takeover` for graceful handoff; parked or expired takeover requires a reason.
- Active takeover requires the exact lease ID, reason, and interactive operator confirmation. Use `close` only after final worktree cleanup.
- Trusted sessions may administratively park exact leases or explicit repository-scoped active batches with a reason. `admin_parked` revokes mutation authority, preserves session identity, and blocks automatic resume; run `admin-unpark` before exact-context resume.
- Use `scripts/agent-lease.sh fence --repo REPO --reason "..."` and `unfence` to block new claims during maintenance. Never edit markers or coordination database state.
- Never manually claim a lease, bypass hooks, reset or stash another session's work, or edit shared main.
- Agent Coordination outage is different from lease denial: verify endpoint health and use configured endpoint-switch guidance before recovery. Never retry an ambiguous lease mutation. Cached leases remain fail-closed for mutation.
- Primary recovery and operator-controlled Hermes standby promotion are documented in `infra-docs/runbooks/AGENT_COORDINATION_RECOVERY.md`; standby is single-writer and never auto-promoted.
- During a confirmed Linear outage, `rollout-linear-enrollment.sh --git-only --operator <name>` remains limited to enrollment-only rollout paths. General degraded execution uses the local outbox and still requires Agent Coordination; it never authorizes merge, release, deploy, or production work without a single-use receipt.
<!-- STAROS_LINEAR_GUARD_CONTRACT_END -->
