package ksef

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
)

func GenerateXML(inv invoice.Invoice) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	buf.WriteString(`<Faktura xmlns="http://crd.gov.pl/wzor/2023/06/29/12648/">` + "\n")
	buf.WriteString("  <Naglowek>\n")
	buf.WriteString(fmt.Sprintf("    <KodWaluty>%s</KodWaluty>\n", string(inv.Currency)))
	buf.WriteString(fmt.Sprintf("    <DataWytworzeniaFa>%s</DataWytworzeniaFa>\n", time.Now().UTC().Format(time.RFC3339)))
	buf.WriteString("  </Naglowek>\n")
	buf.WriteString("  <Podmiot1>\n")
	buf.WriteString(fmt.Sprintf("    <DaneIdentyfikacyjne><NIP>%s</NIP><PelnaNazwa>%s</PelnaNazwa></DaneIdentyfikacyjne>\n",
		stripCountry(inv.Seller.SIRET), inv.Seller.Name))
	buf.WriteString("  </Podmiot1>\n")
	buf.WriteString("  <Podmiot2>\n")
	buf.WriteString(fmt.Sprintf("    <DaneIdentyfikacyjne><NIP>%s</NIP><PelnaNazwa>%s</PelnaNazwa></DaneIdentyfikacyjne>\n",
		stripCountry(inv.Customer.SIRET), inv.Customer.Name))
	buf.WriteString("  </Podmiot2>\n")
	buf.WriteString("  <Fa>\n")
	buf.WriteString(fmt.Sprintf("    <P_1>%s</P_1>\n", inv.IssueDate.Format("2006-01-02")))
	buf.WriteString(fmt.Sprintf("    <P_2>%s</P_2>\n", inv.Number))
	buf.WriteString(fmt.Sprintf("    <P_13_1>%.2f</P_13_1>\n", inv.TotalHT.ToFloat()))
	buf.WriteString(fmt.Sprintf("    <P_14_1>%.2f</P_14_1>\n", inv.TotalVAT.ToFloat()))
	buf.WriteString(fmt.Sprintf("    <P_15>%.2f</P_15>\n", inv.TotalTTC.ToFloat()))

	for idx, it := range inv.Items {
		buf.WriteString("    <FaWiersz>\n")
		buf.WriteString(fmt.Sprintf("      <NrWierszaFa>%d</NrWierszaFa>\n", idx+1))
		buf.WriteString(fmt.Sprintf("      <P_7>%s</P_7>\n", it.Description))
		buf.WriteString(fmt.Sprintf("      <P_8B>%d</P_8B>\n", int(it.Quantity)))
		buf.WriteString(fmt.Sprintf("      <P_9A>%.2f</P_9A>\n", it.UnitPrice.ToFloat()))
		buf.WriteString(fmt.Sprintf("      <P_12>%.0f</P_12>\n", it.VATRate.ToFloat()))
		buf.WriteString("    </FaWiersz>\n")
	}

	buf.WriteString("  </Fa>\n</Faktura>")
	return buf.Bytes(), nil
}

func stripCountry(vat string) string {
	vat = strings.TrimSpace(vat)
	if len(vat) > 2 && (strings.HasPrefix(vat, "PL") || strings.HasPrefix(vat, "IT") || strings.HasPrefix(vat, "FR")) {
		return vat[2:]
	}
	return vat
}
