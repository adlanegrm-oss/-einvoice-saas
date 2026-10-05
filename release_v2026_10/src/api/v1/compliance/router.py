"""
Point d'entrée unifié POST /v1/compliance/validate
"""
from fastapi import APIRouter, UploadFile, File, Form, HTTPException, Header
from typing import Optional, Dict, Any
from ....compliance_core.document_engine.models import RawInvoiceDocument
from ....compliance_core.validation_engine.engine import ComplianceValidationEngine
from ....compliance_core.evidence_engine.sealer import EvidenceSealer

router = APIRouter(prefix="/v1/compliance", tags=["Compliance"])
engine = ComplianceValidationEngine()

@router.post("/validate")
async def validate_invoice(
    file: Optional[UploadFile] = File(None),
    raw_xml: Optional[str] = Form(None),
    profile: Optional[str] = Form(None)
) -> Dict[str, Any]:
    if not file and not raw_xml:
        raise HTTPException(status_code=400, detail="Fournir un fichier XML ou une chaîne raw_xml.")

    content = await file.read() if file else raw_xml.encode("utf-8")
    doc = RawInvoiceDocument(raw_bytes=content, filename=file.filename if file else "payload.xml")

    report = engine.validate(doc, profile_override=profile)
    evidence = EvidenceSealer.generate_evidence_pack(doc, report)

    return {
        "validation_id": report.validation_id,
        "status": report.status.value,
        "compliance_score": report.compliance_score,
        "syntax": report.syntax,
        "profile": report.profile,
        "rule_pack": report.rule_pack_id,
        "execution_time_ms": report.execution_time_ms,
        "errors": [
            {
                "rule_id": v.rule_id,
                "level": v.level.value,
                "severity": v.severity.value,
                "target_field": v.target_field,
                "current_value": v.current_value,
                "expected_value": v.expected_value,
                "delta": v.delta,
                "human_explanation": v.human_explanation,
                "suggested_fix": v.suggested_fix.__dict__ if v.suggested_fix else None
            }
            for v in report.violations
        ],
        "warnings": [{"rule_id": w.rule_id, "human_explanation": w.human_explanation} for w in report.warnings],
        "evidence": {
            "evidence_id": evidence.evidence_id,
            "document_hash_sha256": evidence.document_hash_sha256,
            "canonical_hash_sha256": evidence.canonical_hash_sha256,
            "validated_at": evidence.timestamp_iso,
            "rfc3161_token": evidence.rfc3161_token_simulated
        }
    }
@router.get("/profiles")
def get_compliance_profiles():
    return {"profiles": engine.rpm.list_packs()}
