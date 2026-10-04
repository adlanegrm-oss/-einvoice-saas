package exporter

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"time"
)

// Finalisation Factur-X par mise a jour incrementale (ajout en fin de fichier,
// sans toucher aux octets existants) : independant de la version de pdfcpu.

var (
	reTailStartXref = regexp.MustCompile(`startxref\s+(\d+)\s*%%EOF\s*$`)
	reTrSize        = regexp.MustCompile(`/Size\s+(\d+)`)
	reTrRoot        = regexp.MustCompile(`/Root\s+(\d+)\s+(\d+)\s+R`)
	reTrInfo        = regexp.MustCompile(`/Info\s+\d+\s+\d+\s+R`)
	reTrID          = regexp.MustCompile(`/ID\s*\[[^\]]*\]`)
	reObjHeader     = regexp.MustCompile(`(?:^|[\r\n])\s*(\d+)\s+(\d+)\s+obj\b`)
	reFilespecType  = regexp.MustCompile(`/Type\s*/Filespec`)
	reEFStream      = regexp.MustCompile(`/EF\s*<<[^>]*?/F\s+(\d+)\s+\d+\s+R`)
	reMetadataRef   = regexp.MustCompile(`/Metadata\s+\d+\s+\d+\s+R`)
	reAFArray       = regexp.MustCompile(`/AF\s*\[`)
	reAFRel         = regexp.MustCompile(`/AFRelationship\s*/\w+`)
	reSubtype       = regexp.MustCompile(`/Subtype\s*/[^\s/<>\[\]()]+`)
	reStreamKw      = regexp.MustCompile(`>>\s*stream\r?\n`)
)

type pdfObj struct {
	num, gen, start, end int
}

func objectEnd(pdf []byte, from int) int {
	rest := pdf[from:]
	eo := bytes.Index(rest, []byte("endobj"))
	if eo < 0 {
		return -1
	}
	if loc := reStreamKw.FindIndex(rest[:eo]); loc != nil {
		es := bytes.Index(rest[loc[1]:], []byte("endstream"))
		if es < 0 {
			return -1
		}
		base := loc[1] + es
		eo2 := bytes.Index(rest[base:], []byte("endobj"))
		if eo2 < 0 {
			return -1
		}
		return from + base + eo2 + len("endobj")
	}
	return from + eo + len("endobj")
}

// lastObject retourne la derniere definition de l'objet num.
func lastObject(pdf []byte, num int) (pdfObj, bool) {
	var found pdfObj
	ok := false
	for _, m := range reObjHeader.FindAllSubmatchIndex(pdf, -1) {
		n, _ := strconv.Atoi(string(pdf[m[2]:m[3]]))
		if n != num {
			continue
		}
		g, _ := strconv.Atoi(string(pdf[m[4]:m[5]]))
		end := objectEnd(pdf, m[1])
		if end < 0 {
			continue
		}
		found = pdfObj{num: n, gen: g, start: m[2], end: end}
		ok = true
	}
	return found, ok
}

func insertInDict(obj []byte, kv string) ([]byte, error) {
	i := bytes.Index(obj, []byte("<<"))
	if i < 0 {
		return nil, fmt.Errorf("facturx: objet sans dictionnaire")
	}
	out := make([]byte, 0, len(obj)+len(kv))
	out = append(out, obj[:i+2]...)
	out = append(out, kv...)
	out = append(out, obj[i+2:]...)
	return out, nil
}

func setCatalogKeys(obj []byte, metaNum, afNum int) ([]byte, error) {
	var err error
	meta := fmt.Sprintf("%d 0 R", metaNum)
	af := fmt.Sprintf("%d 0 R", afNum)
	if reMetadataRef.Match(obj) {
		obj = reMetadataRef.ReplaceAll(obj, []byte("/Metadata "+meta))
	} else if obj, err = insertInDict(obj, "/Metadata "+meta); err != nil {
		return nil, err
	}
	if reAFArray.Match(obj) {
		obj = reAFArray.ReplaceAll(obj, []byte("/AF["+af+" "))
	} else if obj, err = insertInDict(obj, "/AF["+af+"]"); err != nil {
		return nil, err
	}
	return obj, nil
}

func setFilespecRelationship(obj []byte, rel string) ([]byte, error) {
	if reAFRel.Match(obj) {
		return reAFRel.ReplaceAll(obj, []byte("/AFRelationship/"+rel)), nil
	}
	return insertInDict(obj, "/AFRelationship/"+rel)
}

func setStreamMIME(obj []byte) ([]byte, error) {
	loc := reStreamKw.FindIndex(obj)
	if loc == nil {
		return nil, fmt.Errorf("facturx: flux EmbeddedFile introuvable")
	}
	head := append([]byte{}, obj[:loc[0]+2]...)
	tail := obj[loc[0]+2:]
	var err error
	if reSubtype.Match(head) {
		head = reSubtype.ReplaceAll(head, []byte("/Subtype/text#2Fxml"))
	} else if head, err = insertInDict(head, "/Subtype/text#2Fxml"); err != nil {
		return nil, err
	}
	return append(head, tail...), nil
}

const facturxXMPTemplate = `<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:pdfaid="http://www.aiim.org/pdfa/ns/id/">
   <pdfaid:part>3</pdfaid:part>
   <pdfaid:conformance>B</pdfaid:conformance>
  </rdf:Description>
  <rdf:Description rdf:about="" xmlns:xmp="http://ns.adobe.com/xap/1.0/">
   <xmp:CreatorTool>einvoice-saas</xmp:CreatorTool>
   <xmp:CreateDate>%s</xmp:CreateDate>
   <xmp:ModifyDate>%s</xmp:ModifyDate>
  </rdf:Description>
  <rdf:Description rdf:about="" xmlns:fx="urn:factur-x:pdfa:CrossIndustryDocument:invoice:1p0#">
   <fx:DocumentType>INVOICE</fx:DocumentType>
   <fx:DocumentFileName>factur-x.xml</fx:DocumentFileName>
   <fx:Version>1.0</fx:Version>
   <fx:ConformanceLevel>%s</fx:ConformanceLevel>
  </rdf:Description>
  <rdf:Description rdf:about="" xmlns:pdfaExtension="http://www.aiim.org/pdfa/ns/extension/" xmlns:pdfaSchema="http://www.aiim.org/pdfa/ns/schema#" xmlns:pdfaProperty="http://www.aiim.org/pdfa/ns/property#">
   <pdfaExtension:schemas>
    <rdf:Bag>
     <rdf:li rdf:parseType="Resource">
      <pdfaSchema:schema>Factur-X PDFA Extension Schema</pdfaSchema:schema>
      <pdfaSchema:namespaceURI>urn:factur-x:pdfa:CrossIndustryDocument:invoice:1p0#</pdfaSchema:namespaceURI>
      <pdfaSchema:prefix>fx</pdfaSchema:prefix>
      <pdfaSchema:property>
       <rdf:Seq>
        <rdf:li rdf:parseType="Resource"><pdfaProperty:name>DocumentFileName</pdfaProperty:name><pdfaProperty:valueType>Text</pdfaProperty:valueType><pdfaProperty:category>external</pdfaProperty:category><pdfaProperty:description>name of the embedded XML invoice file</pdfaProperty:description></rdf:li>
        <rdf:li rdf:parseType="Resource"><pdfaProperty:name>DocumentType</pdfaProperty:name><pdfaProperty:valueType>Text</pdfaProperty:valueType><pdfaProperty:category>external</pdfaProperty:category><pdfaProperty:description>INVOICE</pdfaProperty:description></rdf:li>
        <rdf:li rdf:parseType="Resource"><pdfaProperty:name>Version</pdfaProperty:name><pdfaProperty:valueType>Text</pdfaProperty:valueType><pdfaProperty:category>external</pdfaProperty:category><pdfaProperty:description>The actual version of the Factur-X XML schema</pdfaProperty:description></rdf:li>
        <rdf:li rdf:parseType="Resource"><pdfaProperty:name>ConformanceLevel</pdfaProperty:name><pdfaProperty:valueType>Text</pdfaProperty:valueType><pdfaProperty:category>external</pdfaProperty:category><pdfaProperty:description>The conformance level of the embedded Factur-X data</pdfaProperty:description></rdf:li>
       </rdf:Seq>
      </pdfaSchema:property>
     </rdf:li>
    </rdf:Bag>
   </pdfaExtension:schemas>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>
`

func buildFacturXXMP(profile string, now time.Time) []byte {
	ts := now.UTC().Format(time.RFC3339)
	var b bytes.Buffer
	b.WriteString("<?xpacket begin=\"\ufeff\" id=\"W5M0MpCehiHzreSzNTczkc9d\"?>\n")
	fmt.Fprintf(&b, facturxXMPTemplate, ts, ts, profile)
	b.WriteString("<?xpacket end=\"w\"?>")
	return b.Bytes()
}

func afRelationshipFor(profile string) string {
	switch profile {
	case "MINIMUM", "BASIC WL":
		return "Data"
	}
	return "Alternative"
}

// finalizeFacturX ajoute (mise a jour incrementale) : XMP Factur-X, /Metadata
// et /AF dans le catalogue, /AFRelationship sur le filespec, MIME text/xml.
func finalizeFacturX(pdf []byte, profile string, now time.Time) ([]byte, error) {
	if len(pdf) == 0 {
		return nil, fmt.Errorf("facturx: empty pdf")
	}
	tail := pdf
	if len(tail) > 2048 {
		tail = pdf[len(pdf)-2048:]
	}
	sx := reTailStartXref.FindSubmatch(tail)
	if sx == nil {
		return nil, fmt.Errorf("facturx: startxref introuvable")
	}
	prev, _ := strconv.Atoi(string(sx[1]))

	ti := bytes.LastIndex(pdf, []byte("trailer"))
	if ti < 0 {
		return nil, fmt.Errorf("facturx: trailer classique introuvable (xref stream non supporte)")
	}
	tr := pdf[ti:]
	sm := reTrSize.FindSubmatch(tr)
	rm := reTrRoot.FindSubmatch(tr)
	if sm == nil || rm == nil {
		return nil, fmt.Errorf("facturx: /Size ou /Root introuvable dans le trailer")
	}
	size, _ := strconv.Atoi(string(sm[1]))
	rootNum, _ := strconv.Atoi(string(rm[1]))

	// Dernier objet /Filespec du fichier
	fsNum := -1
	for _, loc := range reFilespecType.FindAllIndex(pdf, -1) {
		hdrs := reObjHeader.FindAllSubmatchIndex(pdf[:loc[0]], -1)
		if len(hdrs) == 0 {
			continue
		}
		h := hdrs[len(hdrs)-1]
		fsNum, _ = strconv.Atoi(string(pdf[h[2]:h[3]]))
	}
	if fsNum < 0 {
		return nil, fmt.Errorf("facturx: objet Filespec introuvable (piece jointe absente)")
	}
	fsObj, ok := lastObject(pdf, fsNum)
	if !ok {
		return nil, fmt.Errorf("facturx: objet Filespec %d illisible", fsNum)
	}
	fsRaw := pdf[fsObj.start:fsObj.end]
	em := reEFStream.FindSubmatch(fsRaw)
	if em == nil {
		return nil, fmt.Errorf("facturx: flux EmbeddedFile introuvable dans le Filespec")
	}
	efNum, _ := strconv.Atoi(string(em[1]))
	efObj, ok := lastObject(pdf, efNum)
	if !ok {
		return nil, fmt.Errorf("facturx: objet EmbeddedFile %d illisible", efNum)
	}
	rootObj, ok := lastObject(pdf, rootNum)
	if !ok {
		return nil, fmt.Errorf("facturx: catalogue %d illisible", rootNum)
	}

	newFS, err := setFilespecRelationship(fsRaw, afRelationshipFor(profile))
	if err != nil {
		return nil, err
	}
	newEF, err := setStreamMIME(pdf[efObj.start:efObj.end])
	if err != nil {
		return nil, err
	}
	metaNum := size
	newRoot, err := setCatalogKeys(pdf[rootObj.start:rootObj.end], metaNum, fsNum)
	if err != nil {
		return nil, err
	}

	xmp := buildFacturXXMP(profile, now)
	var metaObj bytes.Buffer
	fmt.Fprintf(&metaObj, "%d 0 obj\n<</Type/Metadata/Subtype/XML/Length %d>>\nstream\n", metaNum, len(xmp))
	metaObj.Write(xmp)
	metaObj.WriteString("\nendstream\nendobj")

	type rev struct {
		num, gen int
		raw      []byte
	}
	revs := []rev{
		{rootObj.num, rootObj.gen, newRoot},
		{fsObj.num, fsObj.gen, newFS},
		{efObj.num, efObj.gen, newEF},
		{metaNum, 0, metaObj.Bytes()},
	}
	sort.Slice(revs, func(i, j int) bool { return revs[i].num < revs[j].num })

	var out bytes.Buffer
	out.Grow(len(pdf) + 8192)
	out.Write(pdf)
	if pdf[len(pdf)-1] != '\n' {
		out.WriteByte('\n')
	}
	offs := make([]int, len(revs))
	for i, r := range revs {
		offs[i] = out.Len()
		out.Write(r.raw)
		out.WriteByte('\n')
	}
	xrefPos := out.Len()
	out.WriteString("xref\n")
	for i, r := range revs {
		fmt.Fprintf(&out, "%d 1\n%010d %05d n \n", r.num, offs[i], r.gen)
	}
	fmt.Fprintf(&out, "trailer\n<</Size %d/Root %d %d R/Prev %d", size+1, rootNum, rootObj.gen, prev)
	if info := reTrInfo.Find(tr); info != nil {
		out.Write(info)
	}
	if id := reTrID.Find(tr); id != nil {
		out.Write(id)
	}
	fmt.Fprintf(&out, ">>\nstartxref\n%d\n%%%%EOF\n", xrefPos)
	return out.Bytes(), nil
}
