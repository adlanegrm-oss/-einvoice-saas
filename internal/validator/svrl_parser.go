package validator

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

// Structures XML conformes au namespace ISO Schematron SVRL (http://purl.oclc.org/dsdl/svrl)
type rawSVRLOutput struct {
	XMLName           xml.Name              `xml:"schematron-output"`
	FailedAsserts     []rawSVRLFailedAssert `xml:"failed-assert"`
	SuccessfulReports []rawSVRLReport       `xml:"successful-report"`
}

type rawSVRLFailedAssert struct {
	ID       string `xml:"id,attr"`
	Flag     string `xml:"flag,attr"`
	Location string `xml:"location,attr"`
	Text     string `xml:"text"`
}

type rawSVRLReport struct {
	ID       string `xml:"id,attr"`
	Flag     string `xml:"flag,attr"`
	Location string `xml:"location,attr"`
	Text     string `xml:"text"`
}

func mapSVRLSeverity(flag string) Severity {
	switch strings.ToLower(strings.TrimSpace(flag)) {
	case "fatal":
		return SeverityFatal
	case "warning":
		return SeverityWarning
	case "info":
		return SeverityInfo
	default:
		return SeverityError
	}
}

// ParseSVRL analyse un rapport SVRL XML et génère un SchematronReport unifié
func ParseSVRL(svrlData []byte) (*SchematronReport, error) {
	trimmed := bytes.TrimSpace(svrlData)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("svrl: flux SVRL vide")
	}

	var output rawSVRLOutput
	if err := xml.Unmarshal(trimmed, &output); err != nil {
		return nil, fmt.Errorf("svrl: XML mal formé: %w", err)
	}

	report := &SchematronReport{
		Valid:  true,
		Issues: make([]SchematronIssue, 0),
	}

	for _, fa := range output.FailedAsserts {
		sev := mapSVRLSeverity(fa.Flag)
		report.Issues = append(report.Issues, SchematronIssue{
			RuleID:   strings.TrimSpace(fa.ID),
			Severity: sev,
			Message:  strings.TrimSpace(fa.Text),
			XPath:    strings.TrimSpace(fa.Location),
		})
		if sev == SeverityError || sev == SeverityFatal {
			report.Valid = false
		}
	}

	return report, nil
}
