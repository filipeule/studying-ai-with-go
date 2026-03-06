package main

type Config struct {
	Text           string  `json:"text"`
	FilePath       string  `json:"-"`
	SummaryRate    float64 `json:"summaryRate,omitempty"`
	TargetPercent  float64 `json:"targetPercent,omitempty"`
	SummarizerType string  `json:"summarizerType,omitempty"`

	// openai parameters
	OpenAIKey     string `json:"openAIKey,omitempty"`
	OpenAIModel   string `json:"openAIModel,omitempty"`
	OpenAIBaseURL string `json:"openAIBaseURL,omitempty"`

	// hugging face parameters
	HuggingFaceKey   string `json:"huggingFaceKey,omitempty"`
	HuggingFaceModel string `json:"huggingFaceModel,omitempty"`
	HuggingFaceURL   string `json:"huggingFaceURL,omitempty"`
	MaxLength        int    `json:"maxLength,omitempty"`
	MinLength        int    `json:"minLength,omitempty"`
}

type AppResponse struct {
	Summary               string   `json:"summary"`
	Keywords              []string `json:"keywords"`
	OriginalSentenceCount int      `json:"originalSentenceCount"`
	SummarySentenceCount  int      `json:"summarySentenceCount"`
	OriginalWordCount     int      `json:"originalWordCount"`
	SummaryWordCount      int      `json:"summaryWordCount"`
	CompressionRatio      float64  `json:"compressionRatio"`
	SummaryPercentage     float64  `json:"summaryPercentage"`
	Error                 string   `json:"error,omitempty"`
}
