from pathlib import Path

import yaml


WORKFLOW = Path(__file__).parents[2] / ".github/workflows/security-scan.yaml"


def test_security_scan_is_ghas_free_and_preserves_outage_gate() -> None:
    text = WORKFLOW.read_text(encoding="utf-8")
    lowered = text.lower()

    for forbidden in ("github/codeql-action", "security-events", "upload-sarif", "sarif"):
        assert forbidden not in lowered

    assert "format: 'json'" in text
    assert "path: trivy-summary.json" in text
    assert "retention-days: 7" in text
    assert "continue-on-error: true" not in lowered


def test_gitleaks_workflow_covers_current_tree_and_history_without_ghas() -> None:
    workflow_path = Path(__file__).parents[2] / ".github/workflows/security-gitleaks.yml"
    text = workflow_path.read_text(encoding="utf-8")
    lowered = text.lower()
    workflow = yaml.safe_load(text)
    triggers = workflow.get("on", workflow.get(True, {}))

    for forbidden in ("github/codeql-action", "security-events", "upload-sarif", "sarif"):
        assert forbidden not in lowered

    assert "pull_request" in triggers
    assert "schedule" in triggers
    assert "workflow_dispatch" in triggers
    assert "paths" not in triggers
    assert "fetch-depth:" in text
    assert "&& 0 || 1" in text
    assert 'gitleaks" git .' in text
    assert 'gitleaks" dir .' in text
    assert "retention-days: 7" in text
    assert "GITLEAKS_SHA256" in text
    assert "continue-on-error: true" not in lowered
