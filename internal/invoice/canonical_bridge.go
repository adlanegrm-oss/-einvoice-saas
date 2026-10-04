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
TaxID:         i.Seller.VATNumber,
},
Buyer: canonical.Party{
LegalEntityID: i.Customer.SIRET,
TaxID:         i.Customer.VATNumber,
},
MonetaryTotals: canonical.MonetaryTotals{
TaxExclusiveAmount: int64(i.TotalHT.ToFloat() * 100),
TaxInclusiveAmount: int64(i.TotalTTC.ToFloat() * 100),
},
}

return can
}
