"""
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