package validator

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SVRLOutput modélise le rapport standardisé ISO Schematron (SVRL)
type SVRLOutput struct {
	XMLName       xml.Name       `xml:"schematron-output"`
	FailedAsserts []SVRLFailedAssert `xml:"failed-assert"`
}

type SVRLFailedAssert struct {
	ID       string `xml:"id,attr"`
	Flag     string `xml:"flag,attr"`
	Location string `xml:"location,attr"`
	Text     string `xml:"text"`
}

type SchematronSidecarClient struct {
	endpointURL string
	httpClient  *http.Client
}

func NewSchematronSidecarClient(endpointURL string) *SchematronSidecarClient {
	if endpointURL == "" {
		endpointURL = "http://schematron:8080/validate"
	}
	return &SchematronSidecarClient{
		endpointURL: endpointURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// HTTP422ValidationError structure standardisée retournée pour les rejets PDP/PPF
type HTTP422ValidationError struct {
	Status  int                  `json:"status"`
	Code    string               `json:"code"`
	Message string               `json:"message"`
	Violations []RuleViolationDetail `json:"violations"`
}

type RuleViolationDetail struct {
	RuleID   string `json:"rule_id"`
	Severity string `json:"severity"`
	Location string `json:"location"`
	Message  string `json:"message"`
}

func (c *SchematronSidecarClient) Validate(ctx context.Context, profile string, xmlPayload []byte) (*HTTP422ValidationError, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpointURL, bytes.NewReader(xmlPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("X-Validation-Profile", profile)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("schematron sidecar unreachable: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading sidecar response: %w", err)
	}

	var svrl SVRLOutput
	if err := xml.Unmarshal(bodyBytes, &svrl); err != nil {
		return nil, fmt.Errorf("invalid SVRL returned by sidecar: %w", err)
	}

	if len(svrl.FailedAsserts) == 0 {
		return nil, nil // Conforme
	}

	report := &HTTP422ValidationError{
		Status:  http.StatusUnprocessableEntity,
		Code:    "NORMATIVE_VALIDATION_FAILED",
		Message: "Le document est non conforme aux exigences CEN EN 16931 ou CIUS-FR.",
	}

	for _, fa := range svrl.FailedAsserts {
		severity := "ERROR"
		if fa.Flag == "warning" {
			severity = "WARNING"
		}
		report.Violations = append(report.Violations, RuleViolationDetail{
			RuleID:   fa.ID,
			Severity: severity,
			Location: fa.Location,
			Message:  fa.Text,
		})
	}

	return report, nil
}
