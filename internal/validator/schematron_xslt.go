package validator

import (
"bytes"
"fmt"
"os/exec"
)

// XSLTExecutor abstrait l'exécution d'une transformation XSLT
type XSLTExecutor interface {
Transform(xmlInput []byte, xsltSheet []byte) ([]byte, error)
}

// CLIExecutor utilise un binaire externe standard (xsltproc, Saxon-HE, etc.) s'il est présent
type CLIExecutor struct {
BinaryPath string
}

func NewDefaultXSLTExecutor() XSLTExecutor {
return &MockXSLTExecutor{}
}

func NewCLIExecutor(binPath string) *CLIExecutor {
if binPath == "" {
binPath = "xsltproc"
}
return &CLIExecutor{BinaryPath: binPath}
}

func (c *CLIExecutor) Transform(xmlInput []byte, xsltSheet []byte) ([]byte, error) {
cmd := exec.Command(c.BinaryPath, "-", "-")
cmd.Stdin = bytes.NewReader(append(xsltSheet, xmlInput...))

var stdout bytes.Buffer
var stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr

if err := cmd.Run(); err != nil {
return nil, fmt.Errorf("cli xslt: %w (stderr: %s)", err, stderr.String())
}

return stdout.Bytes(), nil
}

// MockXSLTExecutor permet de tester le pipeline sans dépendance OS lourde
type MockXSLTExecutor struct {
MockOutput []byte
MockErr    error
}

func (m *MockXSLTExecutor) Transform(xmlInput []byte, xsltSheet []byte) ([]byte, error) {
if m.MockErr != nil {
return nil, m.MockErr
}
if len(m.MockOutput) > 0 {
return m.MockOutput, nil
}
// SVRL vide par défaut = document valide
return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<schematron-output xmlns="http://purl.oclc.org/dsdl/svrl" title="Validation EN 16931">
</schematron-output>`), nil
}
