"""
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