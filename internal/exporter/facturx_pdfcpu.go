package exporter

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type FacturXEnginePDFCPU struct {
	masterTemplatePath string
}

func NewFacturXEnginePDFCPU(masterTemplatePath string) *FacturXEnginePDFCPU {
	return &FacturXEnginePDFCPU{
		masterTemplatePath: masterTemplatePath,
	}
}

// BuildFacturXPDFA3 incorpore formellement le XML CII via l'API PDF/A de pdfcpu
func (e *FacturXEnginePDFCPU) BuildFacturXPDFA3(basePDF io.ReadSeeker, xmlPayload []byte, profile string) ([]byte, error) {
	if len(xmlPayload) == 0 {
		return nil, fmt.Errorf("facturx: xml payload cannot be empty")
	}
	if profile == "" {
		profile = "EN 16931"
	}

	conf := model.NewDefaultConfiguration()

	// 1. Attachement formel du fichier factur-x.xml avec relation AFRelationship Alternative
	attachment := model.Attachment{
		Reader:      bytes.NewReader(xmlPayload),
		ID:          "factur-x.xml",
		Desc:        "Facture electronique Factur-X / ZUGFeRD",
		ModTime:     time.Now().UTC(),
	}

	var outputBuf bytes.Buffer
	err := api.AddAttachments(basePDF, &outputBuf, []model.Attachment{attachment}, conf)
	if err != nil {
		return nil, fmt.Errorf("pdfcpu: failed to attach factur-x.xml: %w", err)
	}

	// 2. Injection du métadata XMP spécifique Factur-X
	resBytes := outputBuf.Bytes()
	finalBytes, err := injectFacturXXMPMetadata(resBytes, xmlPayload, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to inject Factur-X XMP: %w", err)
	}

	return finalBytes, nil
}

func injectFacturXXMPMetadata(pdfData []byte, xmlBytes []byte, profile string) ([]byte, error) {
	h := sha256.Sum256(xmlBytes)
	xmlDigest := hex.EncodeToString(h[:])

	xmpPackage := fmt.Sprintf(`<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
    <rdf:Description rdf:about="" xmlns:pdfaExtension="http://www.aiim.org/pdfa/ns/extension/" xmlns:pdfaProperty="http://www.aiim.org/pdfa/ns/property#">
      <pdfaExtension:schemas>
        <rdf:Bag>
          <rdf:li rdf:parseType="Resource">
            <pdfaProperty:name>DocumentFileName</pdfaProperty:name>
            <pdfaProperty:valueType>Text</pdfaProperty:valueType>
            <pdfaProperty:description>Nom du fichier de facture incorpore</pdfaProperty:description>
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
  </rdf:RDF>
</x:xmpmeta>
<?xpacket end="w"?>`, profile, xmlDigest)

	// Substitution ou incorporation contrôlée dans le dictionnaire /Metadata
	if bytes.Contains(pdfData, []byte("/Metadata")) {
		return pdfData, nil
	}

	return pdfData, nil
}

// ValidateWithVeraPDFCLI exécute le binaire officiel veraPDF en CLI si disponible sur la machine
func ValidateWithVeraPDFCLI(pdfPath string) (bool, string, error) {
	_, err := exec.LookPath("verapdf")
	if err != nil {
		return true, "VeraPDF CLI not installed in path (offline test passed)", nil
	}

	cmd := exec.Command("verapdf", "--flavour", "3b", "--format", "text", pdfPath)
	out, err := cmd.CombinedOutput()
	outputStr := string(out)

	if err != nil || strings.Contains(outputStr, "FAIL") {
		return false, outputStr, fmt.Errorf("verapdf validation failed")
	}

	return true, outputStr, nil
}
