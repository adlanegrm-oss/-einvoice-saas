# Project Status

## Current Release: Foundation Reliability Fix
**State:** `IN DEVELOPMENT / PRE-ALPHA`
*(Note: 'PRODUCTION VERIFIED' status removed until real-world sandbox and schematron validation are fully implemented).*

### Status of Core Engine
- **Monetary Engine:** Migrated to strict `int64` minor units with Banker's Rounding (Half-Even). Floating-point arithmetic completely removed.
- **VAT Model (BG-23):** Split tax categories (`S`, `Z`, `E`, `AE`, `K`, `G`, `O`) supported with dedicated `TaxPercent` type and exemption reasons.
- **Validation Rules:** Aligned with EN 16931 naming (`BR-02`, `BR-05`, `BR-CO-13`, `BR-CO-15`, `BR-CO-17`). Tolerances (0.05€ / 0.02€) eliminated.
- **Lifecycle Machine:** Single state machine following French life cycle (`BROUILLON`, `DEPOSEE`, `REJETEE`, `REFUSEE`, `ENCAISSEE`). Purge predicate aligned to real terminal states.

### Deferred to Future Releases
- Official Schematron EN 16931 & CIUS-FR integration (Saxon-HE / KoSIT sidecar).
- veraPDF PDF/A-3 Factur-X automated compliance assertion.
- Production PDP / PPF connectors and E-reporting engine.
