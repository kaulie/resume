package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

func contentToMarkdown(c ResumeContent) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n", c.Name)
	if c.Title != "" {
		fmt.Fprintf(&b, "## %s\n", c.Title)
	}
	if c.Contact != "" {
		fmt.Fprintf(&b, "> %s\n", c.Contact)
	}
	b.WriteString("\n")
	if c.Summary != "" {
		b.WriteString("## 自我评价\n\n")
		b.WriteString(c.Summary)
		b.WriteString("\n\n")
	}
	for _, s := range c.Sections {
		fmt.Fprintf(&b, "## %s\n\n", s.Title)
		for _, bullet := range s.Bullets {
			fmt.Fprintf(&b, "- %s\n", bullet)
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func contentToText(c ResumeContent) string {
	var b strings.Builder
	b.WriteString(c.Name + "\n")
	if c.Title != "" {
		b.WriteString(c.Title + "\n")
	}
	if c.Contact != "" {
		b.WriteString(c.Contact + "\n")
	}
	b.WriteString("\n")
	if c.Summary != "" {
		b.WriteString("【自我评价】\n" + c.Summary + "\n\n")
	}
	for _, s := range c.Sections {
		fmt.Fprintf(&b, "【%s】\n", s.Title)
		for _, bullet := range s.Bullets {
			fmt.Fprintf(&b, "- %s\n", bullet)
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func xmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;", "'", "&apos;")
	return r.Replace(s)
}

func wPara(text string, bold bool, size int) string {
	b := ""
	if bold {
		b = "<w:b/>"
	}
	sz := size * 2 // half-points
	return fmt.Sprintf(`<w:p><w:pPr><w:rPr>%s<w:sz w:val="%d"/></w:rPr></w:pPr><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, b, sz, xmlEscape(text))
}

func exportDocx(c ResumeContent) ([]byte, error) {
	var doc strings.Builder
	doc.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	doc.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	doc.WriteString(wPara(c.Name, true, 24))
	if c.Title != "" {
		doc.WriteString(wPara(c.Title, true, 14))
	}
	if c.Contact != "" {
		doc.WriteString(wPara(c.Contact, false, 11))
	}
	if c.Summary != "" {
		doc.WriteString(wPara("自我评价", true, 14))
		doc.WriteString(wPara(c.Summary, false, 11))
	}
	for _, s := range c.Sections {
		doc.WriteString(wPara(s.Title, true, 14))
		for _, bullet := range s.Bullets {
			doc.WriteString(wPara("• "+bullet, false, 11))
		}
	}
	doc.WriteString(`</w:body></w:document>`)

	contentTypes := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` +
		`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>` +
		`<Default Extension="xml" ContentType="application/xml"/>` +
		`<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>` +
		`</Types>`

	rels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>` +
		`</Relationships>`

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	files := map[string]string{
		"[Content_Types].xml": contentTypes,
		"_rels/.rels":         rels,
		"word/document.xml":   doc.String(),
	}
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte(content)); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func resolveFontPath(configured string) (string, error) {
	candidates := []string{}
	if configured != "" {
		candidates = append(candidates, configured)
	}
	if v := os.Getenv("PDF_FONT_PATH"); v != "" {
		candidates = append(candidates, v)
	}
	candidates = append(candidates,
		"/System/Library/Fonts/Supplemental/Arial Unicode.ttf",
		"/System/Library/Fonts/STHeiti Light.ttc",
		"/usr/share/fonts/truetype/wqy/wqy-zenhei.ttc",
		"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/truetype/arphic/uming.ttc",
	)
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("未找到可用的中文字体文件，请在管理页面配置 PDF 字体路径")
}

func exportPDF(c ResumeContent, fontPath string) ([]byte, error) {
	font, err := resolveFontPath(fontPath)
	if err != nil {
		return nil, err
	}
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFontLocation(filepath.Dir(font))
	pdf.AddUTF8Font("cjk", "", filepath.Base(font))
	pdf.AddUTF8Font("cjk", "B", filepath.Base(font))
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	pdf.SetFont("cjk", "B", 22)
	pdf.Cell(0, 10, c.Name)
	pdf.Ln(10)

	if c.Title != "" {
		pdf.SetFont("cjk", "", 12)
		pdf.SetTextColor(37, 99, 235)
		pdf.Cell(0, 7, c.Title)
		pdf.Ln(7)
	}
	if c.Contact != "" {
		pdf.SetFont("cjk", "", 10)
		pdf.SetTextColor(107, 114, 128)
		pdf.MultiCell(0, 5, c.Contact, "", "", false)
		pdf.Ln(3)
	}
	pdf.SetTextColor(31, 41, 55)

	if c.Summary != "" {
		pdf.SetFont("cjk", "B", 13)
		pdf.SetTextColor(29, 78, 216)
		pdf.Cell(0, 8, "自我评价")
		pdf.Ln(8)
		pdf.SetTextColor(31, 41, 55)
		pdf.SetFont("cjk", "", 10.5)
		pdf.MultiCell(0, 5.5, c.Summary, "", "", false)
		pdf.Ln(3)
	}

	for _, s := range c.Sections {
		pdf.SetFont("cjk", "B", 13)
		pdf.SetTextColor(29, 78, 216)
		pdf.Cell(0, 8, s.Title)
		pdf.Ln(8)
		pdf.SetTextColor(31, 41, 55)
		pdf.SetFont("cjk", "", 10.5)
		for _, bullet := range s.Bullets {
			pdf.MultiCell(0, 5.5, "•  "+bullet, "", "", false)
		}
		pdf.Ln(3)
	}

	tmp, err := os.CreateTemp("", "resume-*.pdf")
	if err != nil {
		return nil, err
	}
	path := tmp.Name()
	tmp.Close()
	_ = os.Remove(path)
	if err := pdf.OutputFileAndClose(path); err != nil {
		return nil, err
	}
	defer os.Remove(path)
	return os.ReadFile(path)
}
