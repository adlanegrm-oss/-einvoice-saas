"""
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