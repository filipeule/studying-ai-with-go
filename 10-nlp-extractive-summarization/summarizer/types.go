package summarizer

import (
	"sync"

	"github.com/jdkato/prose/v2"
	"github.com/sashabaranov/go-openai"
)

type Type int

const (
	Extractive             Type = iota // selects existing sentences from the original text
	AbstractiveOpenAI                  // generates new text using openai models
	AbstractiveHuggingFace             // generates new text hugging face models
	Hybrid                             // use both extractive and abstractive techniques
)

type SummaryResponse struct {
	OriginalText          string   `json:"original_text"`
	Summary               string   `json:"summary"`
	Keywords              []string `json:"keywords"`
	OriginalSentenceCount int      `json:"original_sentence_count"`
	SummarySentenceCount  int      `json:"summary_sentence_count"`
	OriginalWordCount     int      `json:"original_word_count"`
	SummaryWordCount      int      `json:"summary_word_count"`
	CompressionRatio      float64  `json:"compression_ratio"`
	SummaryPercentage     float64  `json:"summary_percentage"`
	TargetPercentage      float64  `json:"target_percentage"`
	AbstractiveSummary    bool     `json:"abstractive_summary"`
	RequestedMethod       string   `json:"requested_method"`
	ActualMethod          string   `json:"actual_method"`
	FallbackReason        string   `json:"fallback_reason,omitempty"`
}

type Summarizer interface {
	GenerateSummary() string
	GetResponse() SummaryResponse
	ExtractKeywords(count int) []string
}

type TextSummarizer struct {
	doc              *prose.Document
	text             string
	summary          string
	keywords         []string
	summaryRate      float64
	targetPercentage float64
	summarizerType   Type

	openAIKey     string
	openAIBaseURL string
	openAIModel   string
	openAIClient  *openai.Client
	clientmu      sync.Mutex

	huggingFaceConfig HuggingFaceConfig

	requestedMethod string
	actualMethod    string
	fallbackReason  string
}

type Options func(*TextSummarizer)