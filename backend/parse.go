package main

import (
	"archive/zip"
	"bytes"
	"context"
	"html"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
)

var (
	emailRe = regexp.MustCompile(`[\w.+-]+@[\w-]+\.[\w.-]+`)
	phoneRe = regexp.MustCompile(`(\+?86[- ]?)?1[3-9]\d{9}`)
)

var sectionKeywords = map[string][]string{
	"工作经历": {"工作经历", "工作经验", "职业经历", "work experience", "employment"},
	"项目经验": {"项目经验", "项目经历", "projects", "project experience"},
	"教育背景": {"教育背景", "教育经历", "education"},
	"专业技能": {"专业技能", "技能", "技术栈", "skills"},
	"自我评价": {"自我评价", "个人简介", "个人总结", "summary", "profile"},
	"证书奖项": {"证书", "获奖", "荣誉", "奖项", "awards", "certificate"},
}

func normalizeLine(line string) string {
	line = strings.ReplaceAll(line, "\u00a0", " ")
	return strings.TrimSpace(strings.Join(strings.Fields(line), " "))
}

func detectMime(originalName, mime string) string {
	lower := strings.ToLower(originalName)
	switch {
	case strings.HasSuffix(lower, ".docx"):
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case strings.HasSuffix(lower, ".pdf"):
		return "application/pdf"
	case strings.HasSuffix(lower, ".md"), strings.HasSuffix(lower, ".markdown"):
		return "text/markdown"
	case strings.HasSuffix(lower, ".txt"):
		return "text/plain"
	}
	return mime
}

func extractDocxText(data []byte) (string, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	for _, f := range r.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		b, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return "", err
		}
		s := string(b)
		s = strings.ReplaceAll(s, "</w:p>", "\n")
		s = strings.ReplaceAll(s, "<w:tab/>", "\t")
		s = strings.ReplaceAll(s, "<w:br/>", "\n")
		s = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(s, "")
		return html.UnescapeString(s), nil
	}
	return "", nil
}

func findPDFTextHelper() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	p := filepath.Join(filepath.Dir(exe), "pdftext")
	if st, err := os.Stat(p); err == nil && !st.IsDir() {
		return p
	}
	return ""
}

func extractPDFText(data []byte) (string, error) {
	// 优先用 macOS 原生 PDFKit（对中文/CID 字体 PDF 更可靠），失败再退回纯 Go 解析器。
	if helper := findPDFTextHelper(); helper != "" {
		if text, err := extractPDFViaHelper(helper, data); err == nil && strings.TrimSpace(text) != "" {
			return text, nil
		}
	}
	return extractPDFViaGo(data)
}

func extractPDFViaHelper(helper string, data []byte) (string, error) {
	tmp, err := os.CreateTemp("", "resume-*.pdf")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return "", err
	}
	tmp.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, helper, tmp.Name()).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func extractPDFViaGo(data []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	numPage := r.NumPage()
	for i := 1; i <= numPage; i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		sb.WriteString(text)
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

func parseFileToText(data []byte, mime string) (string, error) {
	switch mime {
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return extractDocxText(data)
	case "application/pdf":
		return extractPDFText(data)
	default:
		return string(data), nil
	}
}

func isSectionHeader(line string) string {
	lower := strings.ToLower(line)
	for title, keys := range sectionKeywords {
		for _, k := range keys {
			if strings.Contains(lower, k) {
				return title
			}
		}
	}
	stripped := regexp.MustCompile(`^[一二三四五六七八九十\d]+[、.．:：\s]+`).ReplaceAllString(line, "")
	lower = strings.ToLower(stripped)
	for title, keys := range sectionKeywords {
		for _, k := range keys {
			if strings.Contains(lower, k) {
				return title
			}
		}
	}
	return ""
}

func splitLines(text string) []string {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		n := normalizeLine(l)
		if n != "" {
			out = append(out, n)
		}
	}
	return out
}

func textToContent(text string) ResumeContent {
	lines := splitLines(text)
	content := ResumeContent{Name: "", Title: "", Contact: "", Summary: "", Sections: []Section{}}

	var contactParts []string
	sections := []Section{}
	var active *Section

	for _, line := range lines {
		if header := isSectionHeader(line); header != "" {
			active = &Section{Title: header, Bullets: []string{}}
			sections = append(sections, *active)
			continue
		}
		if content.Name == "" && len(line) <= 20 && len(sections) == 0 {
			content.Name = line
			continue
		}
		if content.Title == "" && len(sections) == 0 && len(line) <= 40 {
			content.Title = line
			continue
		}
		if emailRe.MatchString(line) || phoneRe.MatchString(line) {
			contactParts = append(contactParts, line)
			continue
		}
		if active != nil {
			active.Bullets = append(active.Bullets, line)
			sections[len(sections)-1] = *active
		} else if content.Summary == "" {
			content.Summary = line
		} else {
			content.Summary += "\n" + line
		}
	}

	content.Contact = strings.Join(contactParts, " · ")
	if content.Name == "" {
		content.Name = "未命名简历"
	}
	if content.Summary != "" && !hasSection(sections, "自我评价") {
		summary := Section{Title: "自我评价", Bullets: splitLines(content.Summary)}
		sections = append([]Section{summary}, sections...)
	}
	content.Sections = sections
	return content
}

func hasSection(sections []Section, title string) bool {
	for _, s := range sections {
		if s.Title == title {
			return true
		}
	}
	return false
}
