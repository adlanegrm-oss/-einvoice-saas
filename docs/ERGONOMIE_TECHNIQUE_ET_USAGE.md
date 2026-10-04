# 🧩 Ergonomie Technique & Guide d'Usage

Guide d'intégration et principes de conception à destination des développeurs utilisant la plateforme.

---

## 1. Principes d'Ergonomie Technique

### 1.1. Précision Monétaire en Centimes (`int64`)
Tous les montants financiers dans `pkg/canonical/model.go` sont manipulés en minor units (centimes d'euros sous forme d'entiers `int64`).
* Élimine les erreurs d'arrondi binaires IEEE 754 propres aux flottants.
* La division par 100 n'est appliquée que lors de la sérialisation finale en XML.

### 1.2. Structure Standardisée des Diagnostics
Le moteur Schematron (`pkg/validation`) retourne un rapport structuré directement exploitable pour les API ou interfaces clientes :
```json
{
  "valid": false,
  "diagnostics": [
    {
      "rule_id": "BR-CO-15",
      "severity": "ERROR",
      "message": "GrandTotal (1800.00) must equal Net (1500.00) + Tax (300.00)",
      "path": "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeSettlement/ram:SpecifiedTradeSettlementHeaderMonetarySummation/ram:GrandTotalAmount"
    }
  ]
}
```

---

## 2. Exemple d'Usage

```go
package main

import (
	"log"
	"github.com/adlanegrm-oss/einvoice-saas/pkg/canonical"
	"github.com/adlanegrm-oss/einvoice-saas/pkg/syntax"
	"github.com/adlanegrm-oss/einvoice-saas/pkg/validation"
)

func ProcessInvoice(inv *canonical.CanonicalInvoice) {
	// 1. Génération du XML CII standard UN/CEFACT D16B
	xmlBytes, err := syntax.GenerateCIIXML(inv)
	if err != nil {
		log.Fatalf("Erreur sérialisation CII: %v", err)
	}

	// 2. Validation stricte selon les règles EN 16931 et CIUS-FR
	engine := validation.NewSchematronEngine(true)
	report, err := engine.ValidateXML(xmlBytes)
	if err != nil || !report.Valid {
		log.Printf("Facture invalide. Diagnostic: %+v", report.Diagnostics)
		return
	}

	log.Println("Facture 100% conforme et prête pour transmission PDP/Chorus Pro")
}
```
