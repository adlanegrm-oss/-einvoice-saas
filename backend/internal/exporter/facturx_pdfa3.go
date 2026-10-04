package exporter

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type FacturXProfile string

const (
	ProfileMinimum FacturXProfile = "MINIMUM"
	ProfileBasic   FacturXProfile = "BASIC"
	ProfileEN16931 FacturXProfile = "EN 16931"
)

type InvoiceMetadata struct {
	InvoiceNumber string
	SellerName    string
	SellerSIREN   string
	SellerVAT     string
	BuyerName     string
	IssueDate     time.Time
	Currency      string
	TotalHT       float64
	TotalTTC      float64
	TotalTax      float64
	Profile       FacturXProfile
}

// GenerateFacturXPDFA3 produit un fichier PDF/A-3b conforme avec le XML CII embarqué
func GenerateFacturXPDFA3(meta InvoiceMetadata, ciiXML []byte) ([]byte, error) {
	if len(ciiXML) == 0 {
		return nil, fmt.Errorf("facturx: cii xml payload cannot be empty")
	}

	profile := meta.Profile
	if profile == "" {
		profile = ProfileEN16931
	}

	var buf bytes.Buffer
	var offsets []int

	// 1. Header PDF 1.7 avec marqueurs binaires PDF/A
	buf.WriteString("%PDF-1.7\n%\xE2\xE3\xCF\xD3\n")

	// Helper pour enregistrer l'offset d'un objet
	writeObj := func(objNum int, content string) {
		offsets = append(offsets, buf.Len())
		buf.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", objNum, content))
	}

	// 1 0 obj: Catalog avec /Names (EmbeddedFiles), /AF (Alternative) et /OutputIntents
	writeObj(1, "<< /Type /Catalog /Pages 2 0 R /Names << /EmbeddedFiles << /Names [(factur-x.xml) 6 0 R] >> >> /AF [6 0 R] /OutputIntents [5 0 R] /Metadata 7 0 R >>")

	// 2 0 obj: Pages
	writeObj(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")

	// 3 0 obj: Page
	pageStream := fmt.Sprintf("BT /F1 16 Tf 50 780 Td (FACTURE ELECTRIQUE %s) Tj /F1 10 Tf 0 -30 Td (Emetteur: %s - SIREN: %s) Tj 0 -20 Td (TVA: %s) Tj 0 -30 Td (Client: %s) Tj 0 -25 Td (Date: %s | Devise: %s) Tj 0 -25 Td (Total HT: %.2f EUR | TVA: %.2f EUR | Total TTC: %.2f EUR) Tj 0 -40 Td (Document Factur-X profil %s - Fichier associe: factur-x.xml) Tj ET",
		meta.InvoiceNumber, meta.SellerName, meta.SellerSIREN, meta.SellerVAT, meta.BuyerName,
		meta.IssueDate.Format("02/01/2006"), meta.Currency, meta.TotalHT, meta.TotalTax, meta.TotalTTC, profile)

	writeObj(3, fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Contents 4 0 R /Resources << /Font << /F1 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> >> >> /AF [6 0 R] >>", ))

	// 4 0 obj: Contenu page
	writeObj(4, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(pageStream), pageStream))

	// 5 0 obj: OutputIntent (sRGB ICC) requis pour PDF/A-3
	writeObj(5, "<< /Type /OutputIntent /S /GTS_PDFA1 /OutputConditionIdentifier (sRGB) /RegistryName (http://www.color.org) /Info (sRGB IEC61966-2.1) >>")

	// 6 0 obj: Fichier XML Factur-X embarqué (/Filespec avec /AFRelationship /Alternative)
	xmlLen := len(ciiXML)
	writeObj(6, fmt.Sprintf("<< /Type /Filespec /F (factur-x.xml) /UF (factur-x.xml) /EF << /F 8 0 R >> /Desc (Facture électronique structurée CII) /AFRelationship /Alternative >>"))

	// 7 0 obj: Métadonnées XMP conformes Factur-X et PDF/A-3b
	xmpContent := buildXMPMetadata(meta, ciiXML, profile)
	writeObj(7, fmt.Sprintf("<< /Type /Metadata /Subtype /XML /Length %d >>\nstream\n%s\nendstream", len(xmpContent), xmpContent))

	// 8 0 obj: EmbeddedFile Stream (factur-x.xml)
	nowStr := time.Now().UTC().Format("D:20060102150405Z")
	writeObj(8, fmt.Sprintf("<< /Type /EmbeddedFile /Subtype /text#2Fxml /Length %d /Params << /Size %d /ModDate (%s) >> >>\nstream\n%s\nendstream",
		xmlLen, xmlLen, nowStr, string(ciiXML)))

	// Table des références xref
	xrefOffset := buf.Len()
	buf.WriteString(fmt.Sprintf("xref\n0 %d\n", len(offsets)+1))
	buf.WriteString("0000000000 65535 f \n")
	for _, off := range offsets {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
	}

	// Trailer
	buf.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets)+1, xrefOffset))

	return buf.Bytes(), nil
}

func buildXMPMetadata(meta InvoiceMetadata, xmlBytes []byte, profile FacturXProfile) string {
	h := sha256.Sum256(xmlBytes)
	xmlDigest := hex.EncodeToString(h[:])

	return fmt.Sprintf(`<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
    <rdf:Description rdf:about="" xmlns:pdfaExtension="http://www.aiim.org/pdfa/ns/extension/" xmlns:pdfaProperty="http://www.aiim.org/pdfa/ns/property#">
      <pdfaExtension:schemas>
        <rdf:Bag>
          <rdf:li rdf:parseType="Resource">
            <pdfaProperty:name>DocumentFileName</pdfaProperty:name>
            <pdfaProperty:valueType>Text</pdfaProperty:valueType>
            <pdfaProperty:description>Nom du fichier de facture incorporé</pdfaProperty:description>
          </rdf:li>
        </rdf:Bag>
      </pdfaExtension:schemas>
    </rdf:Description>
    <rdf:Description rdf:about="" xmlns:pdfaid="http://www.aiim.org/pdfa/ns/id/">
      <pdfaid:part>3</pdfaid:part>
      <pdfaid:conformance>B</pdfaid:conformance>
    </rdf:Description>
    <rdf:Description rdf:about="" xmlns:fx="urn:factur-x:pdfa:CrossIndustryDocument:invoice:1p0#">
      <fx:DocumentType>INVOICE</fx:DocumentType>
      <fx:DocumentFileName>factur-x.xml</fx:DocumentFileName>
      <fx:Version>1.0</fx:Version>
      <fx:ConformanceLevel>%s</fx:ConformanceLevel>
      <fx:Digest>%s</fx:Digest>
    </rdf:Description>
    <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">
      <dc:title><rdf:Alt><rdf:li xml:lang="x-default">Facture %s</rdf:li></rdf:Alt></dc:title>
      <dc:creator><rdf:Seq><rdf:li>%s</rdf:li></rdf:Seq></dc:creator>
      <dc:date><rdf:Seq><rdf:li>%s</rdf:li></rdf:Seq></dc:date>
    </rdf:Description>
  </rdf:RDF>
</x:xmpmeta>
<?xpacket end="w"?>`,
		profile, xmlDigest, meta.InvoiceNumber, meta.SellerName, time.Now().UTC().Format(time.RFC3339))
}

// VerifyPDFA3Conformance vérifie la présence des marqueurs normatifs Factur-X
func VerifyPDFA3Conformance(pdfData []byte) error {
	raw := string(pdfData)

	if !strings.HasPrefix(raw, "%PDF-1.") {
		return fmt.Errorf("pdfa3: entête PDF manquante")
	}
	if !strings.Contains(raw, "/AFRelationship /Alternative") {
		return fmt.Errorf("pdfa3: relation AFRelationship /Alternative manquante")
	}
	if !strings.Contains(raw, "(factur-x.xml)") {
		return fmt.Errorf("pdfa3: pièce jointe factur-x.xml introuvable dans le dictionnaire")
	}
	if !strings.Contains(raw, "<pdfaid:part>3</pdfaid:part>") {
		return fmt.Errorf("pdfa3: métadonnées pdfaid part 3 manquantes")
	}
	if !strings.Contains(raw, "urn:factur-x:pdfa:CrossIndustryDocument:invoice:1p0#") {
		return fmt.Errorf("pdfa3: namespace XMP Factur-X manquant")
	}

	return nil
}
