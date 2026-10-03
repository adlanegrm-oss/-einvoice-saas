package validator

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ExtractEmbeddedXML attempts to recover the Factur-X / ZUGFeRD XML
// that is embedded inside a hybrid PDF, including streams that live
// inside compressed object streams (/ObjStm + FlateDecode).
//
// It also normalises the extracted content to UTF-8 (handles ISO-8859-1
// / Windows-1252 that some generators still emit).
func ExtractEmbeddedXML(pdf []byte) (xml []byte, name string, err error) {
	if len(pdf) < 8 || !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		return nil, "", fmt.Errorf("not a PDF")
	}

	// 1. Collect all FlateDecode streams (raw + object-stream)
	streams := extractFlateStreams(pdf)

	// 2. Known attachment names (case-insensitive)
	candidates := []string{
		"factur-x.xml",
		"zugferd-invoice.xml",
		"xrechnung.xml",
		"zugferd.xml",
		"facturx.xml",
	}

	for _, stream := range streams {
		lower := bytes.ToLower(stream)
		for _, cand := range candidates {
			if bytes.Contains(lower, []byte(cand)) {
				// The XML itself is usually the whole stream or follows the name
				xmlBytes := extractXMLFromStream(stream)
				if xmlBytes != nil {
					xmlBytes = normalizeToUTF8(xmlBytes)
					return xmlBytes, cand, nil
				}
			}
		}
		// Fallback: any well-formed CrossIndustryInvoice or Invoice root
		if xmlBytes := extractXMLFromStream(stream); xmlBytes != nil {
			if bytes.Contains(xmlBytes, []byte("CrossIndustryInvoice")) ||
				bytes.Contains(xmlBytes, []byte("Invoice")) {
				xmlBytes = normalizeToUTF8(xmlBytes)
				return xmlBytes, "embedded.xml", nil
			}
		}
	}

	return nil, "", fmt.Errorf("no embedded Factur-X / ZUGFeRD XML found")
}

// extractFlateStreams walks the PDF and decompresses every stream that
// declares /FlateDecode (including those nested inside /ObjStm).
func extractFlateStreams(pdf []byte) [][]byte {
	var out [][]byte

	// Simple regex for stream dictionaries that mention FlateDecode
	// (works for the vast majority of Factur-X PDFs)
	re := regexp.MustCompile(`(?s)/FlateDecode.*?stream\r?\n(.*?)\r?\nendstream`)
	matches := re.FindAllSubmatch(pdf, -1)

	for _, m := range matches {
		raw := m[1]
		// Strip possible leading whitespace / CR
		raw = bytes.TrimLeft(raw, "\r\n")
		decompressed, err := inflate(raw)
		if err != nil {
			continue
		}
		out = append(out, decompressed)

		// If this is an object stream (/ObjStm), the decompressed data
		// contains further objects that may themselves hold the XML.
		if bytes.Contains(m[0], []byte("/ObjStm")) {
			nested := parseObjStm(decompressed)
			out = append(out, nested...)
		}
	}

	// Also try raw (non-regex) search for the classic attachment names
	// inside already-decompressed content that may have been missed.
	return out
}

func inflate(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		// Some generators omit the zlib header; try raw flate
		r2, err2 := zlib.NewReader(bytes.NewReader(append([]byte{0x78, 0x9c}, data...)))
		if err2 != nil {
			return nil, err
		}
		defer r2.Close()
		return io.ReadAll(r2)
	}
	defer r.Close()
	return io.ReadAll(r)
}

// parseObjStm extracts the individual objects that live inside an
// object stream. The format is: N pairs of (objnum offset) followed by
// the concatenated object bodies.
func parseObjStm(data []byte) [][]byte {
	// Very lightweight parser – good enough for Factur-X embeddings
	var objects [][]byte
	// Look for XML-looking chunks
	re := regexp.MustCompile(`(?s)<\?xml[^>]*>.*?</[^>]+>|<CrossIndustryInvoice[\s\S]*?</CrossIndustryInvoice>|<Invoice[\s\S]*?</Invoice>`)
	for _, m := range re.FindAll(data, -1) {
		objects = append(objects, m)
	}
	return objects
}

func extractXMLFromStream(stream []byte) []byte {
	// Prefer a complete document starting with <?xml or <CrossIndustryInvoice / <Invoice
	startTags := [][]byte{
		[]byte("<?xml"),
		[]byte("<CrossIndustryInvoice"),
		[]byte("<Invoice"),
		[]byte("<rsm:CrossIndustryInvoice"),
	}
	for _, tag := range startTags {
		idx := bytes.Index(stream, tag)
		if idx < 0 {
			continue
		}
		candidate := stream[idx:]
		// Truncate at the matching closing tag if possible
		if end := findClosingTag(candidate); end > 0 {
			return candidate[:end]
		}
		// Otherwise return everything up to the first null or end
		if nul := bytes.IndexByte(candidate, 0); nul > 0 {
			return candidate[:nul]
		}
		return candidate
	}
	return nil
}

func findClosingTag(xml []byte) int {
	// Crude but effective for the two roots we care about
	roots := []string{
		"CrossIndustryInvoice",
		"Invoice",
		"rsm:CrossIndustryInvoice",
	}
	for _, root := range roots {
		open := []byte("<" + root)
		close := []byte("</" + root + ">")
		if !bytes.Contains(xml, open) {
			continue
		}
		idx := bytes.LastIndex(xml, close)
		if idx >= 0 {
			return idx + len(close)
		}
	}
	return -1
}

// normalizeToUTF8 converts ISO-8859-1 / Windows-1252 content to UTF-8
// when the declaration or the actual bytes indicate a legacy encoding.
// Streaming-friendly: works on the whole buffer (Factur-X XMLs are small).
func normalizeToUTF8(data []byte) []byte {
	// Already valid UTF-8 → keep as-is
	if utf8.Valid(data) {
		// Still rewrite a wrong encoding declaration if present
		return rewriteEncodingDecl(data, "UTF-8")
	}

	// Detect declared encoding
	decl := detectEncodingDeclaration(data)
	switch strings.ToLower(decl) {
	case "iso-8859-1", "latin1", "latin-1", "windows-1252", "cp1252":
		return rewriteEncodingDecl(latin1ToUTF8(data), "UTF-8")
	default:
		// Heuristic: treat as ISO-8859-1 (common for older ZUGFeRD)
		return rewriteEncodingDecl(latin1ToUTF8(data), "UTF-8")
	}
}

func detectEncodingDeclaration(data []byte) string {
	// Look at the first 200 bytes for encoding="..."
	head := data
	if len(head) > 200 {
		head = head[:200]
	}
	re := regexp.MustCompile(`(?i)encoding\s*=\s*["']([^"']+)["']`)
	if m := re.FindSubmatch(head); len(m) == 2 {
		return string(m[1])
	}
	return ""
}

func latin1ToUTF8(src []byte) []byte {
	// ISO-8859-1 maps 1:1 onto Unicode code points 0x00-0xFF
	var buf bytes.Buffer
	buf.Grow(len(src) + len(src)/2)
	for _, b := range src {
		if b < 0x80 {
			buf.WriteByte(b)
		} else {
			// Encode as UTF-8 two-byte sequence
			buf.WriteByte(0xC0 | (b >> 6))
			buf.WriteByte(0x80 | (b & 0x3F))
		}
	}
	return buf.Bytes()
}

func rewriteEncodingDecl(data []byte, enc string) []byte {
	re := regexp.MustCompile(`(?i)encoding\s*=\s*["'][^"']*["']`)
	repl := []byte(`encoding="` + enc + `"`)
	if re.Match(data) {
		return re.ReplaceAll(data, repl)
	}
	// Insert after <?xml if present
	if bytes.HasPrefix(bytes.TrimSpace(data), []byte("<?xml")) {
		idx := bytes.Index(data, []byte("?>"))
		if idx > 0 {
			insert := []byte(` encoding="` + enc + `"`)
			out := make([]byte, 0, len(data)+len(insert))
			out = append(out, data[:idx]...)
			out = append(out, insert...)
			out = append(out, data[idx:]...)
			return out
		}
	}
	return data
}

// ---------------------------------------------------------------------------
// Lightweight PDF object helper (used by advanced extractors)
// ---------------------------------------------------------------------------

// readPDFInt reads a big-endian integer of the given size from b.
func readPDFInt(b []byte, size int) (int64, error) {
	if len(b) < size {
		return 0, io.ErrUnexpectedEOF
	}
	switch size {
	case 1:
		return int64(b[0]), nil
	case 2:
		return int64(binary.BigEndian.Uint16(b)), nil
	case 4:
		return int64(binary.BigEndian.Uint32(b)), nil
	default:
		var v int64
		for i := 0; i < size; i++ {
			v = (v << 8) | int64(b[i])
		}
		return v, nil
	}
}

// parsePDFOffset is a tiny helper used by more complete xref parsers.
func parsePDFOffset(s string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}
