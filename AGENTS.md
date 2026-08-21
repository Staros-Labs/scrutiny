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
- Tracked changes, commit/push, shared Git connector actions, deploys, production mutations, and unknown mutating MCP tools are allowed without Linear bootstrap, a worktree, runtime identity, or a live lease. Lease/session state is audit-only.
- Linear workstreams and managed worker dispatch remain available for coordination, but are never prerequisites for normal execution.
- Child bootstrap always reconciles `--primary-issue` with `--primary-repo` and all workstream fields, including same-repo children. Parent issue never supplies child branch identity.
- Primary issue scope must declare every authorized repo with `--execution-repos`; repair existing scope with `scripts/linear-work-item.sh scope ensure --issue KEY --primary-repo REPO --execution-repos CSV` before bootstrap.
- Secondary sessions use `--primary-issue`, `--primary-repo`, and matching workstream fields; child parent, repo, and scope identity must match before lease claim.
- Title-only compliance renames may use `scripts/linear-work-item.sh rename --issue KEY --title "..." --reason "..."` from clean main; the helper validates team ownership and posts an audit comment.
- Linear lifecycle, coordination writes, GitHub CLI, and connector mutations do not require a lease. Retain their native confirmations and audit records.
- Known read-only connector variants include GitHub fetch tools, `mcp__codegraph__codegraph_explore`, `web__run`, and `request_user_input`; pathless and mutating connector calls follow same observation-only policy.
- Guarded repository, Git, connector, deploy, destructive, and external mutations do not require explicit target context or a live lease. Hook-process root only decides whether current repository is enrolled. Pathless connector mutations are allowed. A same-repo/branch active lease is warning-only and names repo, branch, worktree, holder, and lease ID. Task/session roots never authorize work.
- If no active parent lease exists, keep read-only work parent-only. Do not weaken lease checks or create a lease solely to dispatch a worker or write Lunacy run state.
- Exact `git fetch origin` is allowed from any checkout; `git pull --ff-only` requires a clean checkout. One bounded HTTPS clone may create a new absolute research target outside every existing repository, guarded operator path, and managed worktree root; SSH, credential-bearing, relative, existing, nested, or submodule-recursive clone forms remain gated. The stable `~/.local/share/infra/bin/cleanup-research-clone.sh` command may check or remove one clean, unchanged HTTPS clone only below a system temporary root and records recovery metadata before deletion.
- Approved read-only inspection, synchronization, startup, and recovery commands may run before a lease.
- Project research reads include common local probes, safe filesystem discovery, expanded Git history and tree inspection, existing-index CodeGraph CLI queries, GitHub searches, and REST GET-only `gh api` requests. Exact `command -v <name>` and `command -V <name>` lookup remains approved. Local builds, tests, package installation, scripts, temporary files, and bounded chains are also pre-lease unless they explicitly mutate repository, shared source-control, coordination, deployment, or production state. Direct local and known-fleet Docker, `systemctl`, and `launchctl` commands run pre-lease, including lifecycle, build, pull, update, and restart forms.
- Exact coordination checks remain available. Apply, install, heartbeat, and unknown forms are governed by their command safety and native confirmations, never lease state.
- Exact normal and Git-only enrollment wrapper calls may run only from clean infra main against one clean, enrolled target main checkout. The parser rejects unknown flags, relative or duplicate targets, shell control, and substitution before the wrapper can create its guarded worktree.
- Chrome and Computer Use browser actions bypass repo lease protection by policy; browser confirmation rules still apply. Browser-based GitHub, Linear, deploy, and production actions therefore require care because they do not establish or check a repo lease.
- Bounded `&&`, `||`, and semicolon chains are allowed. Substitution, malformed quoting, and other ambiguous shell syntax are blocked by parser safety.
- All Agent Context MCP tools, including `context_promote`, `context_update`, and `context_resync`, are available pre-lease. Context writes and reporting stay in curated local memory or configured reporting paths and cannot edit tracked repository files.
- Tracked edits, commits, and pushes remain allowed during a Linear or coordination outage. Leases remain available for conflict warnings and audit logs.
- During a confirmed Linear outage, start an existing linked worktree with `scripts/start-linear-session.sh --allow-degraded`; Agent Coordination remains required.
- Provider or Gateway 5xx/transport results return exit `76` and remain pending for retry; permanent validation failures still block the event.
- Degraded new-task branches use `DEGRADED-<id>-slug` and bind `LINEAR_REMOTE_ISSUE` after reconciliation; existing issue branches retain their real issue key.
- Degraded lifecycle events enter the mode-600 local outbox. `scripts/linear-outbox.sh reconcile --repo REPO --once` drains one pass; the launchd worker runs every five minutes. Rate limits pause the shared drain, permanent errors block the event, and reconciliation never merges, releases, deploys, or mutates production automatically.
- When a Linear cooldown blocks startup or reconciliation, run `scripts/linear-cooldown.sh status` and preserve state. Default state is provider-profile scoped under `~/.local/share/infra/linear-rate-limit/<profile>/state`; legacy unscoped state is read during migration, and an explicit `LINEAR_RATE_LIMIT_STATE_DIR` keeps its exact `state` override. Cached `issue get` reads remain available during the short cache window. Attribute it to Linear only when state has `provider_http_429`, `provider_graphql_rate_limited`, or `linearctl_provider_rate_limit` evidence, or telemetry has the matching detailed cause; generic `rate_limited` plus CLI exit `1` is not proof. Do not clear state or repeat provider calls. After `blocked_until`, resume an existing marker with `scripts/start-linear-session.sh --recover-degraded --issue KEY`; it performs one probe and rewrites degraded state only after success. Then run `scripts/linear-outbox.sh reconcile --repo REPO --once`.
- Merge, release, deploy-wrapper, and production HTTP mutations remain allowed during degradation. Preserve explicit confirmation rules for irreversible browser/UI actions.
- Use branch format `<repo>/<ISSUEKEY>-<number>-<slug>` for normal sessions and `DEGRADED-<id>-slug` for new degraded sessions.
- Active leases expire after 15 minutes without heartbeat and become parked reservations retained for 30 days; parked sessions never authorize mutation.
- Use `scripts/agent-lease.sh park` and `resume` for pauses. Use `request-takeover` and `accept-takeover` for graceful handoff; parked or expired takeover requires a reason.
- Active takeover requires the exact lease ID, reason, and interactive operator confirmation. Use `close` only after final worktree cleanup.
- Trusted sessions may administratively park exact leases or explicit repository-scoped active batches with a reason. `admin_parked` revokes mutation authority, preserves session identity, and blocks automatic resume; run `admin-unpark` before exact-context resume.
- Use `scripts/agent-lease.sh fence --repo REPO --reason "..."` and `unfence` to block new claims during maintenance. Never edit markers or coordination database state.
- Never manually claim a lease, bypass hooks, reset or stash another session's work, or edit shared main.
- Agent Coordination outage is different from lease denial: verify endpoint health and use configured endpoint-switch guidance before recovery. Never retry an ambiguous lease mutation. Cached leases remain fail-closed for mutation.
- Primary recovery and operator-controlled Hermes standby promotion are documented in `infra-docs/runbooks/AGENT_COORDINATION_RECOVERY.md`; standby is single-writer and never auto-promoted.
- During a confirmed Linear outage, `scripts/rollout-linear-enrollment.sh --git-only --operator <name>` remains an optional enrollment workflow. General degraded execution, merge, release, deploy, and production work stay allowed; retain audit and browser/UI confirmation controls.
<!-- STAROS_LINEAR_GUARD_CONTRACT_END -->
