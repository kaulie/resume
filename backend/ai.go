package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var techKeywords = []string{
	"react", "vue", "angular", "typescript", "javascript", "node", "node.js",
	"python", "java", "go", "golang", "c++", "c#", "php", "ruby", "rust", "swift",
	"kotlin", "sql", "mysql", "postgresql", "mongodb", "redis", "elasticsearch",
	"kafka", "rabbitmq", "docker", "kubernetes", "k8s", "linux", "nginx",
	"webpack", "vite", "babel", "sass", "less", "tailwind", "css", "html",
	"小程序", "uni-app", "flutter", "android", "ios",
	"机器学习", "深度学习", "自然语言处理", "nlp", "计算机视觉", "cv", "大模型", "llm",
	"pytorch", "tensorflow", "算法", "数据结构", "图论", "动态规划",
	"微服务", "分布式", "高并发", "性能优化", "数据库", "缓存", "消息队列",
	"前端", "后端", "全栈", "架构", "测试", "运维", "devops", "ci/cd",
	"产品", "数据分析", "数据挖掘", "可视化", "etl", "hadoop", "spark", "flink",
	"网络安全", "加密", "区块链", "云计算", "aws", "azure", "阿里云", "腾讯云",
	"git", "敏捷", "scrum", "api", "restful", "graphql", "grpc", "http", "tcp",
}

var stopWords = map[string]bool{
	"的": true, "了": true, "和": true, "与": true, "及": true, "或": true, "等": true,
	"有": true, "在": true, "是": true, "为": true, "负责": true,
	"以上": true, "优先": true, "熟悉": true, "精通": true, "了解": true, "掌握": true,
	"能够": true, "具备": true, "相关": true, "经验": true, "能力": true, "要求": true,
	"岗位": true, "职责": true, "任职": true, "描述": true,
	"the": true, "and": true, "or": true, "with": true, "experience": true,
	"skills": true, "requirements": true, "responsibilities": true, "a": true,
	"an": true, "of": true, "to": true, "in": true, "for": true, "on": true,
	"is": true, "are": true, "you": true, "we": true, "will": true, "be": true, "as": true,
}

var enTokenRe = regexp.MustCompile(`[a-z][a-z0-9+#./-]{1,}`)
var cnChunkRe = regexp.MustCompile(`^[\p{Han}a-z0-9+#./-]+$`)

func tokenize(text string) []string {
	lower := strings.ToLower(text)
	var tokens []string
	tokens = append(tokens, enTokenRe.FindAllString(lower, -1)...)
	for _, chunk := range strings.FieldsFunc(lower, func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune("，。；、：！？()（）【】[],;:!?/", r)
	}) {
		chunk = strings.TrimSpace(chunk)
		n := len([]rune(chunk))
		if cnChunkRe.MatchString(chunk) && n >= 2 && n <= 8 {
			tokens = append(tokens, strings.ToLower(chunk))
		}
	}
	return tokens
}

func extractKeywords(jd JD) []string {
	blob := strings.Join([]string{jd.Title, jd.Company, jd.Description, jd.Responsibilities, jd.Requirements, jd.PlusPoints}, "\n")
	tokens := tokenize(blob)
	found := map[string]bool{}
	techSet := map[string]bool{}
	for _, k := range techKeywords {
		techSet[k] = true
	}
	for _, t := range tokens {
		if stopWords[t] {
			continue
		}
		if techSet[t] {
			found[t] = true
		}
		if enTokenRe.MatchString(t) && len(t) >= 2 && !stopWords[t] {
			found[t] = true
		}
	}
	lowerBlob := strings.ToLower(blob)
	for _, kw := range techKeywords {
		if strings.Contains(lowerBlob, kw) {
			found[kw] = true
		}
	}
	out := make([]string, 0, len(found))
	for k := range found {
		out = append(out, k)
	}
	sort.Strings(out)
	if len(out) > 60 {
		out = out[:60]
	}
	return out
}

func resumeText(c ResumeContent) string {
	parts := []string{c.Name, c.Title, c.Contact, c.Summary}
	for _, s := range c.Sections {
		parts = append(parts, s.Title)
		parts = append(parts, s.Bullets...)
	}
	return strings.Join(parts, "\n")
}

func keywordsInText(keywords []string, text string) []string {
	lower := strings.ToLower(text)
	var out []string
	for _, k := range keywords {
		if strings.Contains(lower, strings.ToLower(k)) {
			out = append(out, k)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// 简历优化（动态调整）—— 规则引擎
// ---------------------------------------------------------------------------

func buildOptimization(resume ResumeContent, jd JD) (ResumeContent, []DiffItem) {
	keywords := extractKeywords(jd)
	matched := keywordsInText(keywords, resumeText(resume))
	missing := diffKeywords(keywords, matched)

	var diff []DiffItem
	sections := make([]Section, 0, len(resume.Sections))
	for _, sec := range resume.Sections {
		bullets := make([]string, len(sec.Bullets))
		copy(bullets, sec.Bullets)
		for _, b := range sec.Bullets {
			hit := keywordsInText(keywords, b)
			if len(hit) > 0 {
				diff = append(diff, DiffItem{Section: sec.Title, Original: b, Adjusted: b, Matched: hit})
			}
		}
		if isSkillSection(sec.Title) {
			sort.SliceStable(bullets, func(i, j int) bool {
				return len(keywordsInText(keywords, bullets[i])) > len(keywordsInText(keywords, bullets[j]))
			})
		}
		sections = append(sections, Section{Title: sec.Title, Bullets: bullets})
	}

	notes := []string{}
	if len(matched) > 0 {
		notes = append(notes, "已覆盖 JD 关键词："+strings.Join(matched, "、"))
	}
	if len(missing) > 0 {
		notes = append(notes, "建议补充/突出："+strings.Join(missing, "、"))
	}
	if len(notes) == 0 {
		notes = append(notes, "未提取到明显关键词，建议人工核对 JD 要求。")
	}
	sections = append(sections, Section{Title: "JD 匹配说明", Bullets: notes})

	target := jd.Title
	if jd.Company != "" {
		target = target + " @ " + jd.Company
	}
	content := ResumeContent{
		Name:     resume.Name,
		Title:    resume.Title,
		Contact:  resume.Contact,
		Summary:  "（面向「" + target + "」优化版）" + resume.Summary,
		Sections: sections,
	}
	return content, diff
}

func isSkillSection(title string) bool {
	return strings.Contains(title, "技能") || strings.Contains(strings.ToLower(title), "skill") || strings.Contains(title, "技术栈")
}

func diffKeywords(keywords, matched []string) []string {
	m := map[string]bool{}
	for _, k := range matched {
		m[k] = true
	}
	var out []string
	for _, k := range keywords {
		if !m[k] {
			out = append(out, k)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// 简历评价 —— 规则引擎
// ---------------------------------------------------------------------------

func buildEvaluation(resume ResumeContent, jd JD) (int, []Dimension, []string) {
	keywords := extractKeywords(jd)
	matched := keywordsInText(keywords, resumeText(resume))
	missing := diffKeywords(keywords, matched)
	coverage := 0.0
	if len(keywords) > 0 {
		coverage = float64(len(matched)) / float64(len(keywords))
	}

	skillScore := int(coverage * 100)
	hasProjects := false
	hasWork := false
	for _, s := range resume.Sections {
		if strings.Contains(s.Title, "项目") {
			hasProjects = true
		}
		if strings.Contains(s.Title, "工作") || strings.Contains(s.Title, "职业") {
			hasWork = true
		}
	}
	expScore := int(coverage * 70)
	if hasWork {
		expScore += 15
	}
	if hasProjects {
		expScore += 15
	}
	if expScore > 100 {
		expScore = 100
	}

	var allBullets []string
	for _, s := range resume.Sections {
		allBullets = append(allBullets, s.Bullets...)
	}
	avgLen := 0
	if len(allBullets) > 0 {
		sum := 0
		for _, b := range allBullets {
			sum += len([]rune(b))
		}
		avgLen = sum / len(allBullets)
	}
	clarityScore := 40
	switch {
	case avgLen >= 15 && avgLen <= 80:
		clarityScore = 80
	case avgLen > 80:
		clarityScore = 60
	case avgLen > 0:
		clarityScore = 55
	}

	structureScore := 0
	if resume.Name != "" {
		structureScore += 20
	}
	if resume.Contact != "" {
		structureScore += 20
	}
	if resume.Summary != "" {
		structureScore += 20
	}
	structureScore += len(resume.Sections) * 10
	if structureScore > 100 {
		structureScore = 100
	}

	overall := int(float64(skillScore)*0.4 + float64(expScore)*0.3 + float64(clarityScore)*0.15 + float64(structureScore)*0.15)

	dimensions := []Dimension{
		{Name: "技能匹配", Score: skillScore, Comment: fmt.Sprintf("命中 %d/%d 个 JD 关键词", len(matched), len(keywords))},
		{Name: "经历相关", Score: expScore, Comment: boolComment(hasWork || hasProjects, "已包含工作/项目经历", "缺少工作或项目经历")},
		{Name: "表达清晰", Score: clarityScore, Comment: clarityComment(avgLen)},
		{Name: "结构完整", Score: structureScore, Comment: fmt.Sprintf("共 %d 个模块", len(resume.Sections))},
	}

	suggestions := []string{}
	if len(missing) > 0 {
		limit := 15
		if len(missing) < limit {
			limit = len(missing)
		}
		suffix := ""
		if len(missing) > 15 {
			suffix = " 等"
		}
		suggestions = append(suggestions, "建议补充与以下关键词相关的经历或技能："+strings.Join(missing[:limit], "、")+suffix)
	}
	if !hasWork {
		suggestions = append(suggestions, "建议增加「工作经历」模块，突出与 JD 相关的职责与成果。")
	}
	if !hasProjects {
		suggestions = append(suggestions, "建议增加「项目经验」模块，用 STAR 结构描述项目。")
	}
	if resume.Summary == "" {
		suggestions = append(suggestions, "建议补充个人简介/自我评价，概述与 JD 的匹配度。")
	}
	if resume.Contact == "" {
		suggestions = append(suggestions, "建议补充联系方式（邮箱/电话）。")
	}
	if avgLen > 100 {
		suggestions = append(suggestions, "部分条目过长，建议拆分为更聚焦的要点。")
	}

	return overall, dimensions, suggestions
}

func boolComment(cond bool, yes, no string) string {
	if cond {
		return yes
	}
	return no
}

func clarityComment(avgLen int) string {
	if avgLen == 0 {
		return "内容偏少"
	}
	return fmt.Sprintf("条目平均长度约 %d 字", avgLen)
}

// ---------------------------------------------------------------------------
// 面试备战 —— 规则引擎
// ---------------------------------------------------------------------------

func buildInterviewPrep(jd JD) InterviewPrepContent {
	keywords := extractKeywords(jd)
	top := keywords
	if len(top) > 10 {
		top = top[:10]
	}
	first := func(n int) []string {
		if len(top) < n {
			return top
		}
		return top[:n]
	}

	questions := []QuestionCategory{
		{
			Category: "自我介绍",
			Items: []QuestionItem{{
				Question: "请做一个 1 分钟自我介绍",
				Hint:     fmt.Sprintf("围绕「%s」组织：我是谁 → 与该岗位最相关的 2~3 段经历/技能（%s）→ 为什么匹配。", jd.Title, joinOr(first(4), "核心技能")),
			}},
		},
		{Category: "项目/经历深挖", Items: keywordQuestions(first(5), "请介绍一个你使用 %s 的项目经历", "用 STAR 结构：背景 → 目标 → 你做了什么（突出 %s）→ 结果/数据。")},
		{Category: "技术基础", Items: keywordQuestions(first(6), "谈谈你对 %s 的理解与常见考点", "梳理 %s 的核心概念、原理与易踩坑点，准备 2~3 个可展开的例子。")},
		{
			Category: "行为面试",
			Items: []QuestionItem{
				{Question: "讲一次你解决难题的经历", Hint: "突出目标、阻力、行动、结果与复盘。"},
				{Question: "你和团队意见不一致时怎么处理", Hint: "强调沟通、数据/事实驱动、最终达成共识。"},
				{Question: "你未来的职业规划是什么", Hint: fmt.Sprintf("结合 %s 的发展方向表达稳定性与成长意愿。", jd.Title)},
			},
		},
		{
			Category: "反问环节",
			Items: []QuestionItem{
				{Question: "这个岗位目前最大的挑战是什么？", Hint: "体现你对岗位本身的关注。"},
				{Question: "团队的技术栈与协作流程是怎样的？", Hint: "了解真实工作环境，便于后续准备。"},
			},
		},
	}

	studyTopics := top
	if len(studyTopics) == 0 {
		studyTopics = []string{"JD 核心职责", "岗位所需技术栈", "公司业务背景"}
	}

	projectTips := []string{
		fmt.Sprintf("准备 1~2 个能体现 %s 的完整项目，按 STAR 打磨。", joinOr(first(5), "岗位核心技能")),
		"把项目中的量化结果（性能提升、业务增长、效率提升）提炼成一句话。",
		"提前研究公司业务与 JD 中的加分项，准备对应话术。",
	}

	target := jd.Title
	if jd.Company != "" {
		target = target + " @ " + jd.Company
	}
	return InterviewPrepContent{
		Summary:     fmt.Sprintf("针对「%s」的面试备战建议：先围绕核心关键词（%s）准备自我介绍与项目案例，再逐项补齐技术基础，最后准备行为面试与反问。", target, joinOr(first(6), "见下")),
		Questions:   questions,
		StudyTopics: studyTopics,
		ProjectTips: projectTips,
	}
}

func keywordQuestions(items []string, qFmt, hintFmt string) []QuestionItem {
	out := make([]QuestionItem, 0, len(items))
	for _, k := range items {
		out = append(out, QuestionItem{
			Question: fmt.Sprintf(qFmt, k),
			Hint:     fmt.Sprintf(hintFmt, k),
		})
	}
	return out
}

func joinOr(items []string, fallback string) string {
	if len(items) == 0 {
		return fallback
	}
	return strings.Join(items, "、")
}
