package invoice

import (
"github.com/adlanegrm-oss/einvoice-saas/pkg/canonical"
)

// ToCanonical convertit un modèle de facture interne en CanonicalInvoice normé pour la validation
func (i *Invoice) ToCanonical() *canonical.CanonicalInvoice {
if i == nil {
return nil
}

can := &canonical.CanonicalInvoice{
ID:            i.ID,
InvoiceNumber: i.Number,
Seller: canonical.Party{
LegalEntityID: i.Seller.SIRET,
TaxID:         i.Seller.VATID,
},
Buyer: canonical.Party{
LegalEntityID: i.Customer.SIRET,
TaxID:         i.Customer.VATID,
},
MonetaryTotals: canonical.MonetaryTotals{
NetHT:     int64(i.TotalHT.ToFloat() * 100),
TaxAmount: int64(i.TotalVAT.ToFloat() * 100),
GrossTTC:  int64(i.TotalTTC.ToFloat() * 100),
},
}

return can
}
