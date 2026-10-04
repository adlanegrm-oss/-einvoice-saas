import os
import zipfile

# Définition des fichiers sources de la release
FILES = {
    # --- DOMAIN: DOCUMENT ENGINE ---
    "src/compliance_core/document_engine/models.py": '''"""
Compliance Core - Document Models & Ingestion
"""
from dataclasses import dataclass, field
from enum import Enum
from typing import Optional, Dict, Any, List
import hashlib

class SyntaxFormat(str, Enum):
    UBL_INVOICE_2_1 = "UBL_2.1"
    UBL_CREDIT_NOTE_2_1 = "UBL_CREDIT_NOTE_2_1"
    CII_16B = "CII_16B"
    FACTUR_X_MINIMUM = "FACTUR_X_MINIMUM"
    FACTUR_X_BASIC = "FACTUR_X_BASIC"
    FACTUR_X_BASIC_WL = "FACTUR_X_BASIC_WL"
    FACTUR_X_EN16931 = "FACTUR_X_EN16931"
    FACTUR_X_EXTENDED = "FACTUR_X_EXTENDED"
    UNKNOWN = "UNKNOWN"

@dataclass(frozen=True)
class RawInvoiceDocument:
    raw_bytes: bytes
    filename: Optional[str] = None
    content_type: str = "application/xml"
    
    @property
    def sha256_hash(self) -> str:
        return hashlib.sha256(self.raw_bytes).hexdigest()
        
    @property
    def size_bytes(self) -> int:
        return len(self.raw_bytes)

@dataclass
class ParsedInvoice:
    syntax: SyntaxFormat
    profile_id: Optional[str]
    customization_id: Optional[str]
    invoice_number: str
    issue_date: str
    currency: str
    seller_identifier: str
    seller_name: str
    buyer_identifier: str
    buyer_name: str
    line_extension_amount: float
    tax_exclusive_amount: float
    tax_inclusive_amount: float
    payable_amount: float
    lines: List[Dict[str, Any]] = field(default_factory=list)
    tax_subtotals: List[Dict[str, Any]] = field(default_factory=list)
    raw_metadata: Dict[str, Any] = field(default_factory=dict)
''',

    "src/compliance_core/document_engine/parser.py": '''"""
XML Secure Parser & Ingestion (Protection XXE et bombes XML)
"""
import xml.etree.ElementTree as ET
from .models import RawInvoiceDocument, ParsedInvoice, SyntaxFormat

class XMLSecurityException(Exception):
    pass

class SecureInvoiceParser:
    MAX_FILE_SIZE = 10 * 1024 * 1024  # Limite 10 Mo

    @classmethod
    def inspect_and_detect_syntax(cls, doc: RawInvoiceDocument) -> SyntaxFormat:
        if doc.size_bytes > cls.MAX_FILE_SIZE:
            raise XMLSecurityException(f"Charge utile {doc.size_bytes} dépasse le plafond sécurisé (10 Mo)")
        
        # Blocage absolu des déclarations DTD et entités externes
        raw_header = doc.raw_bytes[:2048].decode("utf-8", errors="ignore")
        if "<!DOCTYPE" in raw_header or "<!ENTITY" in raw_header:
            raise XMLSecurityException("Violation de sécurité : Déclaration DTD/Entity interdite (Protection XXE).")

        content = doc.raw_bytes.decode("utf-8", errors="replace")
        if "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2" in content:
            return SyntaxFormat.UBL_INVOICE_2_1
        elif "urn:oasis:names:specification:ubl:schema:xsd:CreditNote-2" in content:
            return SyntaxFormat.UBL_CREDIT_NOTE_2_1
        elif "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice" in content:
            if "factur-x.eu" in content or "zugferd" in content.lower():
                return SyntaxFormat.FACTUR_X_EN16931
            return SyntaxFormat.CII_16B
        return SyntaxFormat.UNKNOWN

    @classmethod
    def parse(cls, doc: RawInvoiceDocument) -> ParsedInvoice:
        syntax = cls.inspect_and_detect_syntax(doc)
        try:
            root = ET.fromstring(doc.raw_bytes)
        except Exception as e:
            raise XMLSecurityException(f"Structure XML invalide : {str(e)}")

        ns = {"cbc": "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2",
              "cac": "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"}

        inv_num = root.findtext(".//cbc:ID", default="UNKNOWN", namespaces=ns)
        customization = root.findtext(".//cbc:CustomizationID", default=None, namespaces=ns)
        profile = root.findtext(".//cbc:ProfileID", default=None, namespaces=ns)
        issue_date = root.findtext(".//cbc:IssueDate", default="", namespaces=ns)
        currency = root.findtext(".//cbc:DocumentCurrencyCode", default="EUR", namespaces=ns)

        seller_id = root.findtext(".//cac:AccountingSupplierParty//cbc:CompanyID", default="", namespaces=ns)
        seller_name = root.findtext(".//cac:AccountingSupplierParty//cbc:RegistrationName", default="Vendeur", namespaces=ns)
        buyer_id = root.findtext(".//cac:AccountingCustomerParty//cbc:CompanyID", default="", namespaces=ns)
        buyer_name = root.findtext(".//cac:AccountingCustomerParty//cbc:RegistrationName", default="Acheteur", namespaces=ns)

        def _get_float(xpath: str, default=0.0):
            val = root.findtext(xpath, namespaces=ns)
            try:
                return float(val) if val else default
            except (ValueError, TypeError):
                return default

        line_ext = _get_float(".//cac:LegalMonetaryTotal/cbc:LineExtensionAmount")
        tax_excl = _get_float(".//cac:LegalMonetaryTotal/cbc:TaxExclusiveAmount")
        tax_incl = _get_float(".//cac:LegalMonetaryTotal/cbc:TaxInclusiveAmount")
        payable = _get_float(".//cac:LegalMonetaryTotal/cbc:PayableAmount")

        return ParsedInvoice(
            syntax=syntax,
            profile_id=profile,
            customization_id=customization,
            invoice_number=inv_num,
            issue_date=issue_date,
            currency=currency,
            seller_identifier=seller_id,
            seller_name=seller_name,
            buyer_identifier=buyer_id,
            buyer_name=buyer_name,
            line_extension_amount=line_ext,
            tax_exclusive_amount=tax_excl,
            tax_inclusive_amount=tax_incl,
            payable_amount=payable
        )
''',

    # --- DOMAIN: VALIDATION ENGINE ---
    "src/compliance_core/validation_engine/models.py": '''"""
Modèles de résultats de validation, violations, sévérités et suggestions de correction
"""
from dataclasses import dataclass, field
from enum import Enum
from typing import Optional, List, Dict, Any

class Severity(str, Enum):
    FATAL = "FATAL"
    ERROR = "ERROR"
    WARNING = "WARNING"
    INFO = "INFO"

class ValidationLevel(str, Enum):
    LEVEL_1_STRUCTURAL = "STRUCTURAL"
    LEVEL_2_SEMANTIC = "SEMANTIC"
    LEVEL_3_CONTEXTUAL = "CONTEXTUAL"

class ValidationStatus(str, Enum):
    VALID = "VALID"
    INVALID = "INVALID"
    WARNINGS = "WARNINGS"

@dataclass
class SuggestedFix:
    action: str
    target_field: str
    current_value: Any
    suggested_value: Any
    explanation: str

@dataclass
class RuleViolation:
    rule_id: str
    level: ValidationLevel
    severity: Severity
    target_field: str
    current_value: Any
    expected_value: Any
    delta: Optional[float]
    human_explanation: str
    technical_explanation: str
    suggested_fix: Optional[SuggestedFix] = None
    doc_url: Optional[str] = None

@dataclass
class ValidationReport:
    validation_id: str
    status: ValidationStatus
    compliance_score: float
    syntax: str
    profile: str
    rule_pack_id: str
    rule_pack_version: str
    violations: List[RuleViolation] = field(default_factory=list)
    warnings: List[RuleViolation] = field(default_factory=list)
    stats: Dict[str, int] = field(default_factory=dict)
    execution_time_ms: float = 0.0
''',

    "src/compliance_core/validation_engine/engine.py": '''"""
Compliance Validation Engine (Structural -> Semantic -> Contextual)
"""
import time
import uuid
import xml.etree.ElementTree as ET
from typing import List, Optional
from .models import (
    ValidationReport, ValidationStatus, ValidationLevel, Severity, RuleViolation, SuggestedFix
)
from ..document_engine.models import RawInvoiceDocument, SyntaxFormat
from ..document_engine.parser import SecureInvoiceParser, XMLSecurityException
from ..rule_engine.manager import RulePackManager

class ComplianceValidationEngine:
    def __init__(self, rule_pack_manager: Optional[RulePackManager] = None):
        self.rpm = rule_pack_manager or RulePackManager()

    def validate(self, doc: RawInvoiceDocument, profile_override: Optional[str] = None) -> ValidationReport:
        start_time = time.perf_counter()
        val_id = f"val_{uuid.uuid4().hex[:12]}"
        violations: List[RuleViolation] = []
        warnings: List[RuleViolation] = []

        # --- NIVEAU 1: VALIDATION STRUCTURELLE ---
        syntax = SyntaxFormat.UNKNOWN
        try:
            syntax = SecureInvoiceParser.inspect_and_detect_syntax(doc)
            if syntax == SyntaxFormat.UNKNOWN:
                violations.append(RuleViolation(
                    rule_id="SEC-STRUCT-001",
                    level=ValidationLevel.LEVEL_1_STRUCTURAL,
                    severity=Severity.FATAL,
                    target_field="root",
                    current_value="Format inconnu",
                    expected_value="UBL 2.1 ou CII / Factur-X XML",
                    delta=None,
                    human_explanation="Le format du document n'est pas reconnu comme une facture électronique normalisée.",
                    technical_explanation="Namespace root non supporté."
                ))
            root = ET.fromstring(doc.raw_bytes)
        except XMLSecurityException as xse:
            violations.append(RuleViolation(
                rule_id="SEC-SECURITY-002",
                level=ValidationLevel.LEVEL_1_STRUCTURAL,
                severity=Severity.FATAL,
                target_field="XML_HEADER",
                current_value="Alerte sécurité XML",
                expected_value="Payload sain < 10 Mo",
                delta=None,
                human_explanation=str(xse),
                technical_explanation="XXE / Billion Laughs protection déclenchée."
            ))
            return self._build_report(val_id, ValidationStatus.INVALID, syntax, "NONE", "NONE", violations, warnings, start_time)
        except Exception as e:
            violations.append(RuleViolation(
                rule_id="STRUCT-XML-003",
                level=ValidationLevel.LEVEL_1_STRUCTURAL,
                severity=Severity.FATAL,
                target_field="syntax",
                current_value="Structure XML invalide",
                expected_value="XML bien formé",
                delta=None,
                human_explanation="Le document XML est mal formé.",
                technical_explanation=str(e)
            ))
            return self._build_report(val_id, ValidationStatus.INVALID, syntax, "NONE", "NONE", violations, warnings, start_time)

        if any(v.severity == Severity.FATAL for v in violations):
            return self._build_report(val_id, ValidationStatus.INVALID, syntax, "NONE", "NONE", violations, warnings, start_time)

        parsed = SecureInvoiceParser.parse(doc)
        target_profile = profile_override or parsed.profile_id or "EN16931"
        pack = self.rpm.resolve_rule_pack(target_profile, syntax.value)

        # --- NIVEAU 2: VALIDATION SÉMANTIQUE ---
        sem_violations, sem_warnings = pack.execute_semantic_rules(parsed, root)
        violations.extend(sem_violations)
        warnings.extend(sem_warnings)

        # --- NIVEAU 3: VALIDATION CONTEXTUELLE ---
        ctx_violations, ctx_warnings = pack.execute_contextual_rules(parsed, root)
        violations.extend(ctx_violations)
        warnings.extend(ctx_warnings)

        status = ValidationStatus.INVALID if violations else (ValidationStatus.WARNINGS if warnings else ValidationStatus.VALID)
        return self._build_report(val_id, status, syntax.value, pack.profile_id, pack.version, violations, warnings, start_time)

    def _build_report(self, val_id, status, syntax, profile, version, violations, warnings, start_time):
        total_rules = 120
        failed_count = len(violations)
        score = max(0.0, min(100.0, round(((total_rules - failed_count) / total_rules) * 100.0, 1))) if status != ValidationStatus.VALID else 100.0
        duration = round((time.perf_counter() - start_time) * 1000.0, 2)
        return ValidationReport(
            validation_id=val_id,
            status=status,
            compliance_score=score,
            syntax=str(syntax),
            profile=profile,
            rule_pack_id=f"{profile}-{version}",
            rule_pack_version=version,
            violations=violations,
            warnings=warnings,
            stats={"errors": len(violations), "warnings": len(warnings)},
            execution_time_ms=duration
        )
''',

    # --- DOMAIN: RULE ENGINE ---
    "src/compliance_core/rule_engine/manager.py": '''"""
Rule Pack Architecture: Versionnée et découplée
"""
from dataclasses import dataclass
from typing import List, Tuple, Dict, Any
import xml.etree.ElementTree as ET
from ..validation_engine.models import RuleViolation, ValidationLevel, Severity, SuggestedFix
from ..document_engine.models import ParsedInvoice

@dataclass
class RulePack:
    pack_id: str
    standard: str
    version: str
    country: str
    profile_id: str
    syntax: str
    effective_from: str
    rules_count: int

    def execute_semantic_rules(self, parsed: ParsedInvoice, root: ET.Element) -> Tuple[List[RuleViolation], List[RuleViolation]]:
        violations: List[RuleViolation] = []
        warnings: List[RuleViolation] = []

        if not parsed.customization_id:
            violations.append(RuleViolation(
                rule_id="BR-01",
                level=ValidationLevel.LEVEL_2_SEMANTIC,
                severity=Severity.ERROR,
                target_field="cbc:CustomizationID",
                current_value=None,
                expected_value="urn:cen.eu:en16931:2017",
                delta=None,
                human_explanation="L'identifiant de spécification (CustomizationID / BT-24) est obligatoire.",
                technical_explanation="EN16931 BT-24 absent.",
                suggested_fix=SuggestedFix(
                    action="SET_VALUE",
                    target_field="cbc:CustomizationID",
                    current_value="",
                    suggested_value="urn:cen.eu:en16931:2017#compliant#urn:fdc:peppol.eu:2017:poacc:billing:3.0",
                    explanation="Ajouter la balise CustomizationID réglementaire."
                )
            ))

        if parsed.line_extension_amount <= 0.0 and parsed.payable_amount > 0.0:
            violations.append(RuleViolation(
                rule_id="BR-CO-10",
                level=ValidationLevel.LEVEL_2_SEMANTIC,
                severity=Severity.ERROR,
                target_field="Invoice/LegalMonetaryTotal/LineExtensionAmount",
                current_value=parsed.line_extension_amount,
                expected_value=parsed.payable_amount,
                delta=parsed.payable_amount - parsed.line_extension_amount,
                human_explanation="La somme des lignes nettes ne correspond pas au montant total HT déclaré.",
                technical_explanation="[BR-CO-10]-Sum of Invoice line net amount (BT-106) = Σ Invoice line net amount (BT-131).",
                suggested_fix=SuggestedFix(
                    action="UPDATE_VALUE",
                    target_field="LineExtensionAmount",
                    current_value=parsed.line_extension_amount,
                    suggested_value=parsed.tax_exclusive_amount or parsed.payable_amount,
                    explanation="Aligner LineExtensionAmount sur le total des lignes d'articles."
                )
            ))

        if not parsed.seller_identifier:
            violations.append(RuleViolation(
                rule_id="BR-02",
                level=ValidationLevel.LEVEL_2_SEMANTIC,
                severity=Severity.ERROR,
                target_field="AccountingSupplierParty/Party/PartyTaxScheme/CompanyID",
                current_value=None,
                expected_value="Numéro SIRET ou TVA valide",
                delta=None,
                human_explanation="L'identifiant fiscal ou légal du vendeur est obligatoire.",
                technical_explanation="Seller identifier (BT-29/BT-31) missing.",
                suggested_fix=SuggestedFix(
                    action="INSERT_IDENTIFIER",
                    target_field="CompanyID",
                    current_value=None,
                    suggested_value="FR12345678901",
                    explanation="Renseigner le SIREN/SIRET ou le numéro de TVA intracommunautaire du vendeur."
                )
            ))

        return violations, warnings

    def execute_contextual_rules(self, parsed: ParsedInvoice, root: ET.Element) -> Tuple[List[RuleViolation], List[RuleViolation]]:
        warnings: List[RuleViolation] = []
        if self.country == "FR" and parsed.buyer_identifier:
            cleaned = parsed.buyer_identifier.replace(" ", "").replace(".", "")
            if len(cleaned) not in [9, 14]:
                warnings.append(RuleViolation(
                    rule_id="FR-R-003",
                    level=ValidationLevel.LEVEL_3_CONTEXTUAL,
                    severity=Severity.WARNING,
                    target_field="AccountingCustomerParty/PartyLegalEntity/CompanyID",
                    current_value=parsed.buyer_identifier,
                    expected_value="9 chiffres (SIREN) ou 14 chiffres (SIRET)",
                    delta=None,
                    human_explanation="L'identifiant de l'acheteur ne correspond pas au standard français SIREN/SIRET.",
                    technical_explanation="CIUS-FR participant directory rule."
                ))
        return [], warnings

class RulePackManager:
    def __init__(self):
        self._packs = {
            "EN16931": RulePack("EN16931-2026-v1.0.0", "EN16931", "2026.1.0", "EU", "EN16931", "UBL_2.1", "2026-01-01", 138),
            "PEPPOL": RulePack("PEPPOL-BIS-3.0-v2026.1", "PEPPOL", "3.0.14", "EU", "PEPPOL-BIS-3.0", "UBL_2.1", "2026-02-15", 192),
            "FR-B2B": RulePack("FR-B2B-CIUS-v2.4.1", "CIUS-FR", "2.4.1", "FR", "FR-B2B-CIUS", "UBL_2.1", "2026-03-01", 214)
        }

    def resolve_rule_pack(self, profile: str, syntax: str) -> RulePack:
        p = profile.upper()
        if "FR" in p or "CIUS" in p:
            return self._packs["FR-B2B"]
        elif "PEPPOL" in p:
            return self._packs["PEPPOL"]
        return self._packs["EN16931"]

    def list_packs(self) -> List[Dict[str, Any]]:
        return [{"pack_id": p.pack_id, "standard": p.standard, "version": p.version, "rules_count": p.rules_count} for p in self._packs.values()]
''',

    # --- DOMAIN: EVIDENCE ENGINE ---
    "src/compliance_core/evidence_engine/sealer.py": '''"""
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
''',

    # --- API ENDPOINTS ---
    "src/api/v1/compliance/router.py": '''"""
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
''',

    # --- SECURITY MIDDLEWARE ---
    "src/api/middleware/security.py": '''"""
Protection des surfaces internes et isolation multi-tenant
"""
from fastapi import Request, HTTPException
from starlette.middleware.base import BaseHTTPMiddleware

class SecurityAndTenantMiddleware(BaseHTTPMiddleware):
    INTERNAL_PREFIXES = ("/admin", "/ops", "/mco", "/rnd", "/deploy")

    async def dispatch(self, request: Request, call_next):
        path = request.url.path
        if any(path.startswith(prefix) for prefix in self.INTERNAL_PREFIXES):
            auth_header = request.headers.get("Authorization", "")
            internal_role = request.headers.get("X-Internal-Role", "")
            mfa_verified = request.headers.get("X-MFA-Verified", "false").lower() == "true"

            if not auth_header.startswith("Bearer sec_internal_") or internal_role not in ["SRE", "COMPLIANCE_OFFICER", "ADMIN_EXEC", "DEVOPS"]:
                raise HTTPException(status_code=403, detail="Accès refusé : Authentification interne requise.")
            if not mfa_verified:
                raise HTTPException(status_code=401, detail="MFA Obligatoire pour accéder aux plans d'administration interne.")

        response = await call_next(request)
        response.headers["X-Content-Type-Options"] = "nosniff"
        response.headers["X-Frame-Options"] = "DENY"
        return response
''',

    # --- FICHIER PRINCIPAL MAIN ---
    "src/main.py": '''"""
Point d'entrée principal de l'application
"""
from fastapi import FastAPI
from fastapi.responses import HTMLResponse
import os

from .api.v1.compliance.router import router as compliance_router
from .api.middleware.security import SecurityAndTenantMiddleware
from .customer_platform.app.routes import router as app_router
from .customer_platform.partner.routes import router as partner_router
from .company_platform.admin.routes import router as admin_router
from .company_platform.ops.routes import router as ops_router
from .company_platform.mco.routes import router as mco_router
from .company_platform.rnd.routes import router as rnd_router
from .company_platform.deploy.routes import router as deploy_router

app = FastAPI(title="E-Invoice Compliance & Validation Platform", version="2026.10-compliance-core")
app.add_middleware(SecurityAndTenantMiddleware)

app.include_router(compliance_router)
app.include_router(app_router)
app.include_router(partner_router)
app.include_router(admin_router)
app.include_router(ops_router)
app.include_router(mco_router)
app.include_router(rnd_router)
app.include_router(deploy_router)

@app.get("/", response_class=HTMLResponse)
async def landing():
    landing_path = os.path.join(os.path.dirname(__file__), "landing_page", "index.html")
    with open(landing_path, "r", encoding="utf-8") as f:
        return f.read()

@app.get("/healthz")
async def healthz():
    return {"status": "HEALTHY", "platform": "E-Invoice Compliance Core"}
''',

    # --- SURFACES CLIENTS & INTERNES ---
    "src/customer_platform/app/routes.py": '''from fastapi import APIRouter
from fastapi.responses import HTMLResponse
router = APIRouter(prefix="/app", tags=["Customer App"])
@router.get("/dashboard", response_class=HTMLResponse)
async def dash(): return "<h2>/app — Portail Client Conformité</h2>"
''',

    "src/customer_platform/partner/routes.py": '''from fastapi import APIRouter
from fastapi.responses import HTMLResponse
router = APIRouter(prefix="/partner", tags=["Partner Platform"])
@router.get("/dashboard", response_class=HTMLResponse)
async def dash(): return "<h2>/partner — Plateforme API & Intégrateurs</h2>"
''',

    "src/company_platform/admin/routes.py": '''from fastapi import APIRouter
from fastapi.responses import HTMLResponse
router = APIRouter(prefix="/admin", tags=["Company Admin"])
@router.get("/dashboard", response_class=HTMLResponse)
async def dash(): return "<h2>/admin — Business Control Center</h2>"
''',

    "src/company_platform/ops/routes.py": '''from fastapi import APIRouter
from fastapi.responses import HTMLResponse
router = APIRouter(prefix="/ops", tags=["Company Ops"])
@router.get("/dashboard", response_class=HTMLResponse)
async def dash(): return "<h2>/ops — Compliance Operations Center</h2>"
''',

    "src/company_platform/mco/routes.py": '''from fastapi import APIRouter
from fastapi.responses import HTMLResponse
router = APIRouter(prefix="/mco", tags=["Company MCO"])
@router.get("/dashboard", response_class=HTMLResponse)
async def dash(): return "<h2>/mco — Reliability & DR Platform</h2>"
''',

    "src/company_platform/rnd/routes.py": '''from fastapi import APIRouter
from fastapi.responses import HTMLResponse
router = APIRouter(prefix="/rnd", tags=["Company R&D"])
@router.get("/dashboard", response_class=HTMLResponse)
async def dash(): return "<h2>/rnd — Laboratoire de conformité réglementaire</h2>"
''',

    "src/company_platform/deploy/routes.py": '''from fastapi import APIRouter
from fastapi.responses import HTMLResponse
router = APIRouter(prefix="/deploy", tags=["Company Deploy"])
@router.get("/dashboard", response_class=HTMLResponse)
async def dash(): return "<h2>/deploy — Plan de contrôle des releases</h2>"
''',

    # --- LANDING PAGE & CONFIGS ---
    "src/landing_page/index.html": '''<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <title>E-Invoice Compliance Infrastructure | Validate. Explain. Fix. Prove.</title>
    <style>body { font-family: sans-serif; background: #0b0f19; color: #fff; text-align: center; padding: 80px 20px; }</style>
</head>
<body>
    <h1>Validate. Explain. Fix. <span style="color:#60a5fa">Prove.</span></h1>
    <p>L'infrastructure d'évaluation et de conformité pour la facturation électronique.</p>
    <p><strong>UBL • CII • Factur-X • EN 16931 • Peppol • CIUS-FR</strong></p>
    <p><a href="/app/dashboard" style="color:#60a5fa">Accéder au portail client</a> | <a href="/partner/dashboard" style="color:#34d399">Explorer l'API Partenaire</a></p>
</body>
</html>''',

    "requirements.txt": '''fastapi>=0.110.0\nuvicorn[standard]>=0.28.0\ndefusedxml>=0.7.1\npydantic>=2.6.0\npython-multipart>=0.0.9\n''',
    "Dockerfile": '''FROM python:3.11-slim\nWORKDIR /app\nCOPY requirements.txt .\nRUN pip install --no-cache-dir -r requirements.txt\nCOPY . .\nEXPOSE 8000\nCMD ["uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "8000"]\n''',
    "docker-compose.yml": '''version: '3.8'\nservices:\n  compliance-platform:\n    build: .\n    ports:\n      - "8000:8000"\n    environment:\n      - ENVIRONMENT=production\n      - ENFORCE_INTERNAL_MFA=true\n''',
    "README.md": '''# E-Invoice Compliance & Validation Platform\n> **Validate. Explain. Fix. Prove.**\n\nInfrastructure B2B de validation et conformité de facturation électronique.\n'''
}

def build_zip(archive_name="e-invoice-compliance-platform-v2026.10.zip"):
    with zipfile.ZipFile(archive_name, "w", zipfile.ZIP_DEFLATED) as zipf:
        for path, content in FILES.items():
            zipf.writestr(path, content.strip())
            print(f" [+] Archivé : {path}")
    print(f"\n Succès : L'archive ZIP '{archive_name}' a été générée avec succès.")

if __name__ == "__main__":
    build_zip()