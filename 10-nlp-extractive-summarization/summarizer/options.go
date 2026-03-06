package summarizer

import (
	"fmt"

	"github.com/jdkato/prose/v2"
	"github.com/sashabaranov/go-openai"
)

func WithOpenAIKey(apiKey string) Options {
	return func(ts *TextSummarizer) {
		ts.openAIKey = apiKey
	}
}

func WithOpenAIBaseURL(baseURL string) Options {
	return func(ts *TextSummarizer) {
		ts.openAIBaseURL = baseURL
	}
}

func WithOpenAIModel(model string) Options {
	return func(ts *TextSummarizer) {
		ts.openAIModel = model
	}
}

func WithSummarizerType(sType Type) Options {
	return func(ts *TextSummarizer) {
		ts.summarizerType = sType
	}
}

func WithAbstractiveHuggingFace() Options {
	return func(ts *TextSummarizer) {
		ts.summarizerType = AbstractiveHuggingFace
	}
}

func WithAbstractiveOpenAI() Options {
	return func(ts *TextSummarizer) {
		ts.summarizerType = AbstractiveOpenAI
	}
}

func NewTextSummarizer(
	text string, summaryRate float64, targetPercentage float64, opts ...Options,
) (Summarizer, error) {
	doc, err := prose.NewDocument(text)
	if err != nil {
		return nil, fmt.Errorf("error creating NLP document: %w", err)
	}

	ts := &TextSummarizer{
		doc: doc,
		text: text,
		summaryRate: summaryRate,
		targetPercentage: targetPercentage,
		summarizerType: Extractive,
		openAIModel: openai.GPT3Dot5Turbo,
	}

	// apply options
	for _, opt := range opts {
		opt(ts)
	}

	return ts, nil
}
