"""
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