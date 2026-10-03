package fatturapa

import (
"bytes"
"fmt"
"strings"

"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
)

func GenerateXML(inv invoice.Invoice, codiceDestinatario string) ([]byte, error) {
if codiceDestinatario == "" {
codiceDestinatario = "0000000"
}

var buf bytes.Buffer
buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
buf.WriteString(`<p:FatturaElettronica versione="FPR12" xmlns:p="http://ivaservizi.agenziaentrate.gov.it/docs/xsd/fatture/v1.2">` + "\n")
buf.WriteString("  <FatturaElettronicaHeader>\n    <DatiTrasmissione>\n")
buf.WriteString("      <IdTrasmittente><IdPaese>FR</IdPaese><IdCodice>00000000000</IdCodice></IdTrasmittente>\n")
buf.WriteString(fmt.Sprintf("      <ProgressivoInvio>%s</ProgressivoInvio>\n", inv.Number))
buf.WriteString("      <FormatoTrasmissione>FPR12</FormatoTrasmissione>\n")
buf.WriteString(fmt.Sprintf("      <CodiceDestinatario>%s</CodiceDestinatario>\n", codiceDestinatario))
buf.WriteString("    </DatiTrasmissione>\n")

// CedentePrestatore (Vendeur)
buf.WriteString("    <CedentePrestatore>\n      <DatiAnagrafici>\n")
buf.WriteString(fmt.Sprintf("        <IdFiscaleIVA><IdPaese>FR</IdPaese><IdCodice>%s</IdCodice></IdFiscaleIVA>\n", clean(inv.Seller.SIRET)))
buf.WriteString(fmt.Sprintf("        <Anagrafica><Denominazione>%s</Denominazione></Anagrafica>\n", escape(inv.Seller.Name)))
buf.WriteString("        <RegimeFiscale>RF01</RegimeFiscale>\n      </DatiAnagrafici>\n")
buf.WriteString("      <Sede>\n")
buf.WriteString(fmt.Sprintf("        <Indirizzo>%s</Indirizzo><CAP>%s</CAP><Comune>%s</Comune><Nazione>%s</Nazione>\n",
escape(inv.Seller.Address.StreetName), inv.Seller.Address.PostalZone, escape(inv.Seller.Address.CityName), getCountry(inv.Seller.Address.CountryCode, "FR")))
buf.WriteString("      </Sede>\n    </CedentePrestatore>\n")

// CessionarioCommittente (Acheteur)
buf.WriteString("    <CessionarioCommittente>\n      <DatiAnagrafici>\n")
buf.WriteString(fmt.Sprintf("        <IdFiscaleIVA><IdPaese>IT</IdPaese><IdCodice>%s</IdCodice></IdFiscaleIVA>\n", clean(inv.Customer.SIRET)))
buf.WriteString(fmt.Sprintf("        <Anagrafica><Denominazione>%s</Denominazione></Anagrafica>\n", escape(inv.Customer.Name)))
buf.WriteString("      </DatiAnagrafici>\n      <Sede>\n")
buf.WriteString(fmt.Sprintf("        <Indirizzo>%s</Indirizzo><CAP>%s</CAP><Comune>%s</Comune><Nazione>%s</Nazione>\n",
escape(inv.Customer.Address.StreetName), inv.Customer.Address.PostalZone, escape(inv.Customer.Address.CityName), getCountry(inv.Customer.Address.CountryCode, "IT")))
buf.WriteString("      </Sede>\n    </CessionarioCommittente>\n")
buf.WriteString("  </FatturaElettronicaHeader>\n")

// Corps
buf.WriteString("  <FatturaElettronicaBody>\n    <DatiGenerali>\n      <DatiGeneraliDocumento>\n")
buf.WriteString("        <TipoDocumento>TD01</TipoDocumento>\n")
buf.WriteString(fmt.Sprintf("        <Divisa>%s</Divisa>\n", string(inv.Currency)))
buf.WriteString(fmt.Sprintf("        <Data>%s</Data>\n", inv.IssueDate.Format("2006-01-02")))
buf.WriteString(fmt.Sprintf("        <Numero>%s</Numero>\n", escape(inv.Number)))
buf.WriteString(fmt.Sprintf("        <ImportoTotaleDocumento>%.2f</ImportoTotaleDocumento>\n", inv.TotalTTC.ToFloat()))
buf.WriteString("      </DatiGeneraliDocumento>\n    </DatiGenerali>\n")

// Lignes
buf.WriteString("    <DatiBeniServizi>\n")
for idx, item := range inv.Items {
buf.WriteString("      <DettaglioLinee>\n")
buf.WriteString(fmt.Sprintf("        <NumeroLinea>%d</NumeroLinea>\n", idx+1))
buf.WriteString(fmt.Sprintf("        <Descrizione>%s</Descrizione>\n", escape(item.Description)))
buf.WriteString(fmt.Sprintf("        <Quantita>%d</Quantita>\n", int(item.Quantity)))
buf.WriteString(fmt.Sprintf("        <PrezzoUnitario>%.2f</PrezzoUnitario>\n", item.UnitPrice.ToFloat()))
lineTot := float64(int(item.Quantity)) * item.UnitPrice.ToFloat()
buf.WriteString(fmt.Sprintf("        <PrezzoTotale>%.2f</PrezzoTotale>\n", lineTot))
buf.WriteString(fmt.Sprintf("        <AliquotaIVA>%.2f</AliquotaIVA>\n", item.VATRate.ToFloat()))
buf.WriteString("      </DettaglioLinee>\n")
}

buf.WriteString("      <DatiRiepilogo>\n")
buf.WriteString("        <AliquotaIVA>20.00</AliquotaIVA>\n")
buf.WriteString(fmt.Sprintf("        <ImponibileImporto>%.2f</ImponibileImporto>\n", inv.TotalHT.ToFloat()))
buf.WriteString(fmt.Sprintf("        <Imposta>%.2f</Imposta>\n", inv.TotalVAT.ToFloat()))
buf.WriteString("        <EsigibilitaIVA>I</EsigibilitaIVA>\n")
buf.WriteString("      </DatiRiepilogo>\n")
buf.WriteString("    </DatiBeniServizi>\n")
buf.WriteString("  </FatturaElettronicaBody>\n</p:FatturaElettronica>")

return buf.Bytes(), nil
}

func escape(s string) string {
s = strings.ReplaceAll(s, "&", "&amp;")
s = strings.ReplaceAll(s, "<", "&lt;")
s = strings.ReplaceAll(s, ">", "&gt;")
return s
}

func getCountry(c, def string) string {
if len(c) == 2 { return strings.ToUpper(c) }
return def
}

func clean(s string) string {
s = strings.TrimSpace(s)
if s != "" { return s }
return "99999999999"
}
