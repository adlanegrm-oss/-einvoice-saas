"""
Evidence Engine: Scellement cryptographique et preuve vérifiable RFC 3161
"""
import hashlib
import json
import base64
from datetime import datetime, timezone
from dataclasses import dataclass, asdict
from typing import Dict, Any, Optional
from ..validation_engine.models import ValidationReport
from ..document_engine.models import RawInvoiceDocument

@dataclass
class EvidencePack:
    evidence_id: str
    document_hash_sha256: str
    canonical_hash_sha256: str
    rule_pack_id: str
    rule_pack_hash: str
    engine_version: str
    validation_status: str
    compliance_score: float
    timestamp_iso: str
    rfc3161_token_simulated: str
    signature_manifest_b64: str

class EvidenceSealer:
    ENGINE_VERSION = "2026.10-compliance-core"

    @classmethod
    def generate_evidence_pack(cls, doc: RawInvoiceDocument, report: ValidationReport) -> EvidencePack:
        doc_hash = doc.sha256_hash
        clean_xml = b"".join(line.strip() for line in doc.raw_bytes.splitlines())
        c14n_hash = hashlib.sha256(clean_xml).hexdigest()
        rp_hash = hashlib.sha256(report.rule_pack_id.encode("utf-8")).hexdigest()
        now_utc = datetime.now(timezone.utc).isoformat()

        token_payload = {
            "version": 1,
            "policy": "1.3.6.1.4.1.99999.compliance.timestamp",
            "message_imprint": {"hash_algorithm": "SHA-256", "hashed_message": doc_hash},
            "gen_time": now_utc,
            "status": "GRANTED"
        }
        rfc3161_sim = base64.b64encode(json.dumps(token_payload).encode("utf-8")).decode("ascii")

        manifest = {
            "doc_hash": doc_hash,
            "validation_id": report.validation_id,
            "status": report.status.value,
            "score": report.compliance_score,
            "rule_pack_id": report.rule_pack_id,
            "timestamp": now_utc
        }
        manifest_sig = hashlib.sha256(json.dumps(manifest, sort_keys=True).encode("utf-8")).hexdigest()
        manifest_b64 = base64.b64encode(json.dumps({"manifest": manifest, "seal": manifest_sig}).encode("utf-8")).decode("ascii")

        return EvidencePack(
            evidence_id=f"evi_{report.validation_id.replace('val_', '')}",
            document_hash_sha256=doc_hash,
            canonical_hash_sha256=c14n_hash,
            rule_pack_id=report.rule_pack_id,
            rule_pack_hash=rp_hash,
            engine_version=cls.ENGINE_VERSION,
            validation_status=report.status.value,
            compliance_score=report.compliance_score,
            timestamp_iso=now_utc,
            rfc3161_token_simulated=rfc3161_sim,
            signature_manifest_b64=manifest_b64
        )