package main

type Section struct {
	Title   string   `json:"title"`
	Bullets []string `json:"bullets"`
}

type ResumeContent struct {
	Name     string    `json:"name"`
	Title    string    `json:"title"`
	Contact  string    `json:"contact"`
	Summary  string    `json:"summary"`
	Sections []Section `json:"sections"`
}

type Resume struct {
	ID           string        `json:"id"`
	Title        string        `json:"title"`
	OriginalName string        `json:"originalName"`
	Mime         string        `json:"mime"`
	SourceText   string        `json:"sourceText"`
	Content      ResumeContent `json:"content"`
	CreatedAt    string        `json:"createdAt"`
	UpdatedAt    string        `json:"updatedAt"`
}

type JD struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Company          string `json:"company"`
	Description      string `json:"description"`
	Responsibilities string `json:"responsibilities"`
	Requirements     string `json:"requirements"`
	PlusPoints       string `json:"plusPoints"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

type DiffItem struct {
	Section  string   `json:"section"`
	Original string   `json:"original"`
	Adjusted string   `json:"adjusted"`
	Matched  []string `json:"matched"`
}

type ResumeVersion struct {
	ID        string        `json:"id"`
	ResumeID  string        `json:"resumeId"`
	JDID      *string       `json:"jdId"`
	Label     string        `json:"label"`
	Content   ResumeContent `json:"content"`
	Diff      []DiffItem    `json:"diff"`
	CreatedAt string        `json:"createdAt"`
}

type Dimension struct {
	Name    string `json:"name"`
	Score   int    `json:"score"`
	Comment string `json:"comment"`
}

type Evaluation struct {
	ID           string      `json:"id"`
	ResumeID     string      `json:"resumeId"`
	JDID         string      `json:"jdId"`
	OverallScore int         `json:"overallScore"`
	Dimensions   []Dimension `json:"dimensions"`
	Suggestions  []string    `json:"suggestions"`
	CreatedAt    string      `json:"createdAt"`
}

type QuestionItem struct {
	Question string `json:"question"`
	Hint     string `json:"hint"`
}

type QuestionCategory struct {
	Category string         `json:"category"`
	Items    []QuestionItem `json:"items"`
}

type InterviewPrepContent struct {
	Summary     string             `json:"summary"`
	Questions   []QuestionCategory `json:"questions"`
	StudyTopics []string           `json:"studyTopics"`
	ProjectTips []string           `json:"projectTips"`
}

type InterviewPrep struct {
	ID        string               `json:"id"`
	JDID      string               `json:"jdId"`
	Content   InterviewPrepContent `json:"content"`
	CreatedAt string               `json:"createdAt"`
}

type Settings struct {
	LLMEnabled  bool   `json:"llmEnabled"`
	LLMBaseURL  string `json:"llmBaseUrl"`
	LLMModel    string `json:"llmModel"`
	LLMAPIKey   string `json:"llmApiKey"`
	PDFFontPath string `json:"pdfFontPath"`
}

// normalizeContent 兜底：历史数据里 sections 可能是 null（旧版解析），
// 统一成空数组，避免前端对 null 调 map 报错。
func normalizeContent(c *ResumeContent) {
	if c.Sections == nil {
		c.Sections = []Section{}
	}
}
