package fatturapa

import (
"encoding/xml"
"fmt"
"strconv"
"strings"
"time"

"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
)

type FatturaElettronica struct {
XMLName xml.Name `xml:"FatturaElettronica"`
Header  Header   `xml:"FatturaElettronicaHeader"`
Body    Body     `xml:"FatturaElettronicaBody"`
}

type Header struct {
CedentePrestatore Supplier `xml:"CedentePrestatore"`
Cessionario       Customer `xml:"CessionarioCommittente"`
}

type Supplier struct {
DatiAnagrafici struct {
IdFiscaleIVA struct {
IdPaese  string `xml:"IdPaese"`
IdCodice string `xml:"IdCodice"`
} `xml:"IdFiscaleIVA"`
Anagrafica struct {
Denominazione string `xml:"Denominazione"`
} `xml:"Anagrafica"`
} `xml:"DatiAnagrafici"`
Sede Address `xml:"Sede"`
}

type Customer struct {
DatiAnagrafici struct {
IdFiscaleIVA struct {
IdPaese  string `xml:"IdPaese"`
IdCodice string `xml:"IdCodice"`
} `xml:"IdFiscaleIVA"`
Anagrafica struct {
Denominazione string `xml:"Denominazione"`
Nome          string `xml:"Nome"`
Cognome       string `xml:"Cognome"`
} `xml:"Anagrafica"`
} `xml:"DatiAnagrafici"`
Sede Address `xml:"Sede"`
}

type Address struct {
Indirizzo string `xml:"Indirizzo"`
CAP       string `xml:"CAP"`
Comune    string `xml:"Comune"`
Nazione   string `xml:"Nazione"`
}

type Body struct {
DatiGenerali struct {
DatiGeneraliDocumento struct {
Divisa        string `xml:"Divisa"`
Data          string `xml:"Data"`
Numero        string `xml:"Numero"`
ImportoTotale string `xml:"ImportoTotaleDocumento"`
} `xml:"DatiGeneraliDocumento"`
} `xml:"DatiGenerali"`
DatiBeniServizi struct {
DettaglioLinee []LineItem      `xml:"DettaglioLinee"`
DatiRiepilogo  []SummaryRecord `xml:"DatiRiepilogo"`
} `xml:"DatiBeniServizi"`
}

type LineItem struct {
NumeroLinea  int     `xml:"NumeroLinea"`
Descrizione  string  `xml:"Descrizione"`
Quantita     float64 `xml:"Quantita"`
PrezzoUnit   float64 `xml:"PrezzoUnitario"`
PrezzoTotale float64 `xml:"PrezzoTotale"`
AliquotaIVA  float64 `xml:"AliquotaIVA"`
}

type SummaryRecord struct {
AliquotaIVA float64 `xml:"AliquotaIVA"`
Imponibile  float64 `xml:"ImponibileImporto"`
Imposta     float64 `xml:"Imposta"`
}

func Parse(xmlBytes []byte) (*invoice.Invoice, error) {
var fpa FatturaElettronica
if err := xml.Unmarshal(xmlBytes, &fpa); err != nil {
return nil, fmt.Errorf("échec décodage FatturaPA: %w", err)
}

doc := fpa.Body.DatiGenerali.DatiGeneraliDocumento
currency := invoice.Currency(doc.Divisa)
if currency == "" {
currency = invoice.CurrencyEUR
}

parsedDate, _ := time.Parse("2006-01-02", doc.Data)
sellerName := fpa.Header.CedentePrestatore.DatiAnagrafici.Anagrafica.Denominazione
sellerID := fpa.Header.CedentePrestatore.DatiAnagrafici.IdFiscaleIVA.IdCodice

buyerName := fpa.Header.Cessionario.DatiAnagrafici.Anagrafica.Denominazione
if buyerName == "" {
buyerName = strings.TrimSpace(fpa.Header.Cessionario.DatiAnagrafici.Anagrafica.Nome + " " + fpa.Header.Cessionario.DatiAnagrafici.Anagrafica.Cognome)
}
buyerID := fpa.Header.Cessionario.DatiAnagrafici.IdFiscaleIVA.IdCodice

items := make([]invoice.InvoiceItem, 0, len(fpa.Body.DatiBeniServizi.DettaglioLinee))
for _, l := range fpa.Body.DatiBeniServizi.DettaglioLinee {
qty := int(l.Quantita)
if qty <= 0 {
qty = 1
}
items = append(items, invoice.InvoiceItem{
Description: l.Descrizione,
Quantity:    qty,
UnitPrice:   invoice.NewMoneyFromFloat(l.PrezzoUnit, 2, currency),
VATRate:     invoice.NewMoneyFromFloat(l.AliquotaIVA, 2, currency),
})
}

var sumHT, sumVAT float64
for _, r := range fpa.Body.DatiBeniServizi.DatiRiepilogo {
sumHT += r.Imponibile
sumVAT += r.Imposta
}

totTTCFloat, _ := strconv.ParseFloat(doc.ImportoTotale, 64)
if totTTCFloat == 0 {
totTTCFloat = sumHT + sumVAT
}

return &invoice.Invoice{
ID:        doc.Numero,
Number:    doc.Numero,
IssueDate: parsedDate,
Currency:  currency,
Seller: invoice.Party{
Name:  sellerName,
SIRET: sellerID,
Address: invoice.PostalAddress{
StreetName:  fpa.Header.CedentePrestatore.Sede.Indirizzo,
PostalZone:  fpa.Header.CedentePrestatore.Sede.CAP,
CityName:    fpa.Header.CedentePrestatore.Sede.Comune,
CountryCode: fpa.Header.CedentePrestatore.Sede.Nazione,
},
},
Customer: invoice.Party{
Name:  buyerName,
SIRET: buyerID,
Address: invoice.PostalAddress{
StreetName:  fpa.Header.Cessionario.Sede.Indirizzo,
PostalZone:  fpa.Header.Cessionario.Sede.CAP,
CityName:    fpa.Header.Cessionario.Sede.Comune,
CountryCode: fpa.Header.Cessionario.Sede.Nazione,
},
},
Items:    items,
TotalHT:  invoice.NewMoneyFromFloat(sumHT, 2, currency),
TotalVAT: invoice.NewMoneyFromFloat(sumVAT, 2, currency),
TotalTTC: invoice.NewMoneyFromFloat(totTTCFloat, 2, currency),
}, nil
}
