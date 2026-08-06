from pathlib import Path


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
