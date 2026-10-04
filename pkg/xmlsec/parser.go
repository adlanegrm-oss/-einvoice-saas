package xmlsec

import (
"encoding/xml"
"errors"
"io"
)

const (
MaxXMLSize  = 10 * 1024 * 1024
MaxDepth    = 32
MaxElements = 50000
)

var (
ErrXMLTooLarge     = errors.New("fichier XML depasse la taille limite")
ErrXMLTooDeep      = errors.New("profondeur XML excessive (soupcon bomb)")
ErrTooManyElements = errors.New("nombre excessif d'elements XML")
ErrDTDForbidden    = errors.New("les declarations DTD/XXE sont strictement interdites")
)

func ValidateSecureXML(r io.Reader) error {
limitedReader := io.LimitReader(r, MaxXMLSize+1)
decoder := xml.NewDecoder(limitedReader)
decoder.Entity = map[string]string{}

depth := 0
elements := 0

for {
token, err := decoder.Token()
if errors.Is(err, io.EOF) {
break
}
if err != nil {
return err
}

switch token.(type) {
case xml.Directive:
return ErrDTDForbidden
case xml.StartElement:
depth++
elements++
if depth > MaxDepth {
return ErrXMLTooDeep
}
if elements > MaxElements {
return ErrTooManyElements
}
case xml.EndElement:
depth--
}

if decoder.InputOffset() > MaxXMLSize {
return ErrXMLTooLarge
}
}

return nil
}