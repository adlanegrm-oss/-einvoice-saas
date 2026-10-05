package exporter

import (
"fmt"
"os"
"os/exec"
)

type PDFAValidationResult struct {
Valid  bool
Output string
}

// ValidatePDFA3WithVeraPDF exécute VeraPDF pour certifier la conformité PDF/A-3.
func ValidatePDFA3WithVeraPDF(pdfPath string) (*PDFAValidationResult, error) {
if _, err := os.Stat(pdfPath); err != nil {
return nil, fmt.Errorf("pdfa: input file not found: %w", err)
}

cmd := exec.Command("verapdf", "--format", "text", pdfPath)
output, err := cmd.CombinedOutput()

result := &PDFAValidationResult{
Valid:  err == nil,
Output: string(output),
}

if err != nil {
return result, fmt.Errorf("pdfa validation failed: %w", err)
}

return result, nil
}
