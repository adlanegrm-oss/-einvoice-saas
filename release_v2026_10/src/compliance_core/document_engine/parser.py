"""
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