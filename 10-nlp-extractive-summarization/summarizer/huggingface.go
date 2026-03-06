package summarizer

import (
	"net/http"
	"sync"
)

type HuggingFaceConfig struct {
	APIKey string
	ModelName string
	InferenceURL string
	MaxLength int
	MinLength int
	Client *http.Client
	clientmu sync.Mutex
}