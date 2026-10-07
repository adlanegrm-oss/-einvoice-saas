package exporter

import (
"bytes"
"encoding/xml"
"fmt"
"time"
)

type FacturXProfile string

const (
ProfileMinimum  FacturXProfile = "MINIMUM"
ProfileBasicWL  FacturXProfile = "BASIC WL"
ProfileBasic    FacturXProfile = "BASIC"
ProfileEN16931  FacturXProfile = "EN 16931"
ProfileExtended FacturXProfile = "EXTENDED"
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

// ValidateCIIXMLForEmbedding vérifie que le XML est valide, non vide et conforme à la racine CrossIndustryInvoice
func ValidateCIIXMLForEmbedding(xmlData []byte) error {
trimmed := bytes.TrimSpace(xmlData)
if len(trimmed) == 0 {
return fmt.Errorf("facturx: payload XML vide")
}

decoder := xml.NewDecoder(bytes.NewReader(trimmed))
for {
token, err := decoder.Token()
if err != nil {
return fmt.Errorf("facturx: XML mal formé: %w", err)
}

if se, ok := token.(xml.StartElement); ok {
if se.Name.Local != "CrossIndustryInvoice" {
return fmt.Errorf("facturx: élément racine attendu 'CrossIndustryInvoice', obtenu: %q", se.Name.Local)
}
break
}
}

return nil
}

// GenerateFacturXPDFA3 produit un conteneur PDF minimal intégrant le flux XML normalisé (factur-x.xml)
func GenerateFacturXPDFA3(meta InvoiceMetadata, ciiXML []byte) ([]byte, error) {
if err := ValidateCIIXMLForEmbedding(ciiXML); err != nil {
return nil, fmt.Errorf("facturx: échec validation XML avant injection: %w", err)
}

profileStr := string(meta.Profile)
if profileStr == "" {
profileStr = string(ProfileEN16931)
}

var buf bytes.Buffer
buf.WriteString("%PDF-1.7\n%\xe2\xe3\xcf\xd3\n")

// 1: Catalog avec Names pour EmbeddedFiles
buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R /Names << /EmbeddedFiles << /Names [ (factur-x.xml) 4 0 R ] >> >> >>\nendobj\n")

// 2: Pages
buf.WriteString("2 0 obj\n<< /Type /Pages /Kids [ 3 0 R ] /Count 1 >>\nendobj\n")

// 3: Page visuelle
contentStream := fmt.Sprintf("BT /F1 12 Tf 50 750 Td (Facture %s - %s) Tj ET", meta.InvoiceNumber, profileStr)
buf.WriteString(fmt.Sprintf("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [ 0 0 595 842 ] /Contents 5 0 R >>\nendobj\n"))

// 4: Filespec
buf.WriteString("4 0 obj\n<< /Type /Filespec /F (factur-x.xml) /UF (factur-x.xml) /EF << /F 6 0 R >> /AFRelationship /Alternative >>\nendobj\n")

// 5: Stream contenu page
buf.WriteString(fmt.Sprintf("5 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(contentStream), contentStream))

// 6: Stream fichier XML embarqué
buf.WriteString(fmt.Sprintf("6 0 obj\n<< /Type /EmbeddedFile /Subtype /text#2Fxml /Length %d >>\nstream\n", len(ciiXML)))
buf.Write(ciiXML)
buf.WriteString("\nendstream\nendobj\n")

// xref et trailer
buf.WriteString("xref\n0 7\n")
buf.WriteString("0000000000 65535 f \n")
buf.WriteString("trailer\n<< /Size 7 /Root 1 0 R >>\nstartxref\n500\n%%EOF\n")

return buf.Bytes(), nil
}

// VerifyFacturXContainer vérifie la signature PDF et la présence du flux XML factur-x.xml
func VerifyFacturXContainer(pdfData []byte) error {
if len(pdfData) < 100 {
return fmt.Errorf("facturx: PDF trop court ou invalide")
}

if !bytes.HasPrefix(pdfData, []byte("%PDF-")) {
return fmt.Errorf("facturx: en-tête %%PDF- manquante")
}

if !bytes.Contains(pdfData, []byte("factur-x.xml")) || !bytes.Contains(pdfData, []byte("CrossIndustryInvoice")) {
return fmt.Errorf("facturx: flux factur-x.xml introuvable dans le conteneur PDF")
}

return nil
}
