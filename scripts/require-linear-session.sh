#!/usr/bin/env bash
set -euo pipefail

SCRIPT_NAME="$(basename "$0")"

resolve_infra_repo_path() {
  local candidate=""
  local common_git_dir=""

  if [ -n "${LINEAR_INFRA_REPO_PATH:-}" ] && [ -d "$LINEAR_INFRA_REPO_PATH" ]; then
    printf '%s\n' "$LINEAR_INFRA_REPO_PATH"
    return 0
  fi

  candidate="$(git config --get codex.linearInfraRepoPath 2>/dev/null || true)"
  if [ -n "$candidate" ] && [ -d "$candidate" ]; then
    printf '%s\n' "$candidate"
    return 0
  fi

  common_git_dir="$(git rev-parse --path-format=absolute --git-common-dir 2>/dev/null || true)"
  if [ -n "$common_git_dir" ]; then
    candidate="$(cd "$(dirname "$common_git_dir")/.." 2>/dev/null && pwd)/infra"
    if [ -d "$candidate" ]; then
      printf '%s\n' "$candidate"
      return 0
    fi
  fi

  candidate="$HOME/Documents/Repos.nosync/infra"
  if [ -d "$candidate" ]; then
    printf '%s\n' "$candidate"
    return 0
  fi

  echo "Unable to resolve the shared infra checkout for $SCRIPT_NAME." >&2
  echo "Set LINEAR_INFRA_REPO_PATH or git config codex.linearInfraRepoPath <path>." >&2
  return 2
}

INFRA_REPO_PATH="$(resolve_infra_repo_path)"
HELPER="$INFRA_REPO_PATH/scripts/$SCRIPT_NAME"
if [ ! -x "$HELPER" ]; then
  echo "Missing shared helper: $HELPER" >&2
  echo "Set LINEAR_INFRA_REPO_PATH or codex.linearInfraRepoPath to a valid infra checkout." >&2
  exit 2
fi

exec "$HELPER" "$@"
