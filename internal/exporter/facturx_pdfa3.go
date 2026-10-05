package exporter

import (
"bytes"
"encoding/xml"
"fmt"
"io"
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
BuyerName     string
IssueDate     time.Time
Currency      string
TotalHT       float64
TotalTTC      float64
Profile       FacturXProfile
}

// validateCIIXMLPayload vérifie que le payload XML n'est pas vide,
// est bien formé et a pour racine CrossIndustryInvoice.
func validateCIIXMLPayload(ciiXML []byte) error {
if len(bytes.TrimSpace(ciiXML)) == 0 {
return fmt.Errorf("facturx: empty CII XML")
}

decoder := xml.NewDecoder(bytes.NewReader(ciiXML))
var rootSeen bool

for {
tok, err := decoder.Token()
if err == io.EOF {
break
}
if err != nil {
return fmt.Errorf("facturx: invalid CII XML: %w", err)
}

if start, ok := tok.(xml.StartElement); ok && !rootSeen {
rootSeen = true
if start.Name.Local != "CrossIndustryInvoice" {
return fmt.Errorf("facturx: root element must be CrossIndustryInvoice, got %s", start.Name.Local)
}
}
}

if !rootSeen {
return fmt.Errorf("facturx: missing CII root element")
}

return nil
}

// GenerateFacturXPDFA3 génère un conteneur PDF avec le XML CII embarqué.
//
// ATTENTION : cette fonction produit une structure de conteneur mais ne doit pas
// être considérée comme une garantie de conformité PDF/A-3b complète (absence de profil ICC intégré, etc.).
// La conformité PDF/A stricte doit être validée par un outil externe (ex: VeraPDF).
func GenerateFacturXPDFA3(meta InvoiceMetadata, ciiXML []byte) ([]byte, error) {
if err := validateCIIXMLPayload(ciiXML); err != nil {
return nil, err
}

if meta.Profile == "" {
meta.Profile = ProfileEN16931
}

var buf bytes.Buffer

// En-tête PDF 1.7
buf.WriteString("%PDF-1.7\n%\xe2\xe3\xcf\xd3\n")

// 1: Catalog
off1 := buf.Len()
buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R /Names << /EmbeddedFiles << /Names [(factur-x.xml) 6 0 R] >> >> /AF [6 0 R] /OutputIntents [4 0 R] /Metadata 5 0 R >>\nendobj\n")

// 2: Pages
off2 := buf.Len()
buf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

// 3: Page simple
off3 := buf.Len()
buf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << >> /Contents [] >>\nendobj\n")

// 4: OutputIntent sRGB
off4 := buf.Len()
buf.WriteString("4 0 obj\n<< /Type /OutputIntent /S /GTS_PDFA1 /OutputConditionIdentifier (sRGB) /Info (sRGB IEC61966-2.1) >>\nendobj\n")

// Métadonnées XMP PDF/A-3b & Factur-X
xmp := fmt.Sprintf(`<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
    <rdf:Description rdf:about="" xmlns:pdfaid="http://www.aiim.org/pdfa/ns/id/">
      <pdfaid:part>3</pdfaid:part>
      <pdfaid:conformance>B</pdfaid:conformance>
    </rdf:Description>
    <rdf:Description rdf:about="" xmlns:fx="urn:factur-x:pdfa:CrossIndustryDocument:invoice:1p0#">
      <fx:DocumentType>INVOICE</fx:DocumentType>
      <fx:DocumentFileName>factur-x.xml</fx:DocumentFileName>
      <fx:Version>1.0</fx:Version>
      <fx:ConformanceLevel>%s</fx:ConformanceLevel>
    </rdf:Description>
  </rdf:RDF>
</x:xmpmeta>
<?xpacket end="w"?>`, meta.Profile)

// 5: Métadonnées Stream
off5 := buf.Len()
buf.WriteString(fmt.Sprintf("5 0 obj\n<< /Type /Metadata /Subtype /XML /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(xmp), xmp))

// 6: Filespec
off6 := buf.Len()
buf.WriteString("6 0 obj\n<< /Type /Filespec /F (factur-x.xml) /UF (factur-x.xml) /EF << /F 7 0 R >> /AFRelationship /Alternative /Desc (Factur-X Invoice XML) >>\nendobj\n")

// 7: EmbeddedFile Stream (CII XML exact)
off7 := buf.Len()
buf.WriteString(fmt.Sprintf("7 0 obj\n<< /Type /EmbeddedFile /Subtype /text#2Fxml /Length %d >>\nstream\n", len(ciiXML)))
buf.Write(ciiXML)
buf.WriteString("\nendstream\nendobj\n")

// XREF
startXref := buf.Len()
buf.WriteString("xref\n0 8\n")
buf.WriteString("0000000000 65535 f \n")
for _, off := range []int{off1, off2, off3, off4, off5, off6, off7} {
buf.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
}

// Trailer
buf.WriteString(fmt.Sprintf("trailer\n<< /Size 8 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", startXref))

return buf.Bytes(), nil
}

// VerifyFacturXContainer vérifie la présence des marqueurs structurels Factur-X dans le flux PDF.
func VerifyFacturXContainer(pdfData []byte) error {
if len(pdfData) == 0 {
return fmt.Errorf("facturx: empty PDF")
}

raw := string(pdfData)

if !strings.HasPrefix(raw, "%PDF-1.") {
return fmt.Errorf("facturx: PDF header missing")
}

required := []string{
"/AFRelationship /Alternative",
"(factur-x.xml)",
"<pdfaid:part>3</pdfaid:part>",
"<pdfaid:conformance>B</pdfaid:conformance>",
"urn:factur-x:pdfa:CrossIndustryDocument:invoice:1p0#",
}

for _, marker := range required {
if !strings.Contains(raw, marker) {
return fmt.Errorf("facturx: required marker missing: %s", marker)
}
}

return nil
}

// VerifyPDFA3Conformance est maintenu pour compatibilité descendante.
func VerifyPDFA3Conformance(pdfData []byte) error {
return VerifyFacturXContainer(pdfData)
}
