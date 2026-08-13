from pathlib import Path

import yaml


WORKFLOW = Path(__file__).parents[2] / ".github/workflows/security-scan.yaml"


def test_security_scan_is_ghas_free_and_preserves_outage_gate() -> None:
    text = WORKFLOW.read_text(encoding="utf-8")
    lowered = text.lower()

    for forbidden in ("github/codeql-action", "security-events", "upload-sarif", "sarif"):
        assert forbidden not in lowered

    assert "format: 'json'" in text
    assert "trivy-summary.json" in text
    assert "trivy-sbom.cdx.json" in text
    assert "retention-days: 7" in text
    assert "continue-on-error: true" not in text.split("\n  osv:", 1)[0].lower()


def test_security_scan_covers_dependencies_iac_and_sbom() -> None:
    text = WORKFLOW.read_text(encoding="utf-8")
    workflow = yaml.safe_load(text)
    triggers = workflow.get("on", workflow.get(True, {}))

    assert "pull_request" in triggers
    assert "schedule" in triggers
    assert "paths" not in triggers
    assert "osv:" in text
    assert 'scan source -r "${{ github.workspace }}"' in text
    assert "scan-type: 'config'" in text
    assert "format: 'cyclonedx'" in text
    assert "OSV_SHA256" in text
    assert "trivy.redacted.json" in text
    assert "trivy-iac.redacted.json" in text
    assert "with_entries(" in text
    assert "path: ${{ runner.temp }}/trivy-artifact" in text
    assert "path: ${{ runner.temp }}/trivy-iac.redacted.json" in text
    assert "path: trivy-summary.json" not in text
    assert "path: trivy-iac-summary.json" not in text
    assert "runs-on: ubuntu-24.04" in text
    assert "actions/checkout@fbc6f3992d24b796d5a048ff273f7fcc4a7b6c09" in text
    assert "actions/upload-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02" in text


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
