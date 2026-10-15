package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func llmGenerate(ctx context.Context, s Settings, system, user string) (string, error) {
	if !s.LLMEnabled || s.LLMAPIKey == "" {
		return "", errors.New("llm disabled")
	}
	base := strings.TrimRight(s.LLMBaseURL, "/")
	if base == "" {
		base = "https://api.deepseek.com"
	}
	url := base
	if !strings.HasSuffix(base, "/chat/completions") {
		url = base + "/chat/completions"
	}

	payload := map[string]any{
		"model": s.LLMModel,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"temperature": 0.2,
		"stream":      false,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.LLMAPIKey)

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", errors.New("llm http " + resp.Status + ": " + truncate(string(b), 300))
	}
	var cr chatResponse
	if err := json.Unmarshal(b, &cr); err != nil {
		return "", err
	}
	if len(cr.Choices) == 0 {
		return "", errors.New("llm empty choices")
	}
	return cr.Choices[0].Message.Content, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func extractJSON(s string) string {
	i := strings.Index(s, "{")
	j := strings.LastIndex(s, "}")
	if i >= 0 && j > i {
		return s[i : j+1]
	}
	return s
}

type optimizeResp struct {
	Content ResumeContent `json:"content"`
	Diff    []DiffItem    `json:"diff"`
}

type evaluateResp struct {
	OverallScore int         `json:"overallScore"`
	Dimensions   []Dimension `json:"dimensions"`
	Suggestions  []string    `json:"suggestions"`
}

func optimizeWithLLM(ctx context.Context, s Settings, resume ResumeContent, jd JD) (ResumeContent, []DiffItem, error) {
	sys := "你是一位资深简历顾问，负责把简历针对目标 JD 做动态调整。只输出 JSON，不要输出任何其他文字。"
	user := "目标 JD：\n" + jdText(jd) + "\n\n原始简历：\n" + mustJSON(resume) +
		"\n\n请输出 JSON：{\"content\":{\"name\":\"\",\"title\":\"\",\"contact\":\"\",\"summary\":\"\",\"sections\":[{\"title\":\"\",\"bullets\":[]}]},\"diff\":[{\"section\":\"\",\"original\":\"\",\"adjusted\":\"\",\"matched\":[]}]}。" +
		"content 为调整后的简历内容（保留原始信息，针对 JD 突出匹配点）；diff 列出每个有改动的条目。"
	raw, err := llmGenerate(ctx, s, sys, user)
	if err != nil {
		return ResumeContent{}, nil, err
	}
	var out optimizeResp
	if err := json.Unmarshal([]byte(extractJSON(raw)), &out); err != nil {
		return ResumeContent{}, nil, err
	}
	if out.Content.Name == "" && len(out.Content.Sections) == 0 {
		return ResumeContent{}, nil, errors.New("llm returned empty content")
	}
	return out.Content, out.Diff, nil
}

func evaluateWithLLM(ctx context.Context, s Settings, resume ResumeContent, jd JD) (int, []Dimension, []string, error) {
	sys := "你是一位招聘负责人，负责针对目标 JD 评价简历。只输出 JSON，不要输出任何其他文字。"
	user := "目标 JD：\n" + jdText(jd) + "\n\n简历：\n" + mustJSON(resume) +
		"\n\n请输出 JSON：{\"overallScore\":0,\"dimensions\":[{\"name\":\"\",\"score\":0,\"comment\":\"\"}],\"suggestions\":[]}。" +
		"overallScore 为 0~100 整数；dimensions 至少包含技能匹配、经历相关、表达清晰、结构完整四项。"
	raw, err := llmGenerate(ctx, s, sys, user)
	if err != nil {
		return 0, nil, nil, err
	}
	var out evaluateResp
	if err := json.Unmarshal([]byte(extractJSON(raw)), &out); err != nil {
		return 0, nil, nil, err
	}
	if out.OverallScore == 0 {
		return 0, nil, nil, errors.New("llm returned empty evaluation")
	}
	return out.OverallScore, out.Dimensions, out.Suggestions, nil
}

func prepWithLLM(ctx context.Context, s Settings, jd JD) (InterviewPrepContent, error) {
	sys := "你是面试教练，负责根据 JD 输出面试备战建议。只输出 JSON，不要输出任何其他文字。"
	user := "目标 JD：\n" + jdText(jd) +
		"\n\n请输出 JSON：{\"summary\":\"\",\"questions\":[{\"category\":\"\",\"items\":[{\"question\":\"\",\"hint\":\"\"}]}],\"studyTopics\":[],\"projectTips\":[]}。" +
		"questions 覆盖自我介绍、项目深挖、技术基础、行为面试、反问环节。"
	raw, err := llmGenerate(ctx, s, sys, user)
	if err != nil {
		return InterviewPrepContent{}, err
	}
	var out InterviewPrepContent
	if err := json.Unmarshal([]byte(extractJSON(raw)), &out); err != nil {
		return InterviewPrepContent{}, err
	}
	if out.Summary == "" && len(out.Questions) == 0 {
		return InterviewPrepContent{}, errors.New("llm returned empty prep")
	}
	return out, nil
}

func jdText(jd JD) string {
	return "职位：" + jd.Title + "\n公司：" + jd.Company + "\n描述：" + jd.Description +
		"\n职责：" + jd.Responsibilities + "\n要求：" + jd.Requirements + "\n加分项：" + jd.PlusPoints
}

func mustJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
