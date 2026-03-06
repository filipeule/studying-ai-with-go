package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// load env variables, ignoring errors
	_ = godotenv.Load()

	// define and parse command line flags
	apiMode := flag.Bool("api", false, "run in api mode")
	port := flag.String("port", "8080", "port to run the api server on")
	filePath := flag.String("file", "", "file path to summarize")

	// summary configuration
	summaryRate := flag.Float64("rate", 0.3, "summary rate for extractive summarization")
	targetPercent := flag.Float64("percent", 30.0, "target summary percentage for abstractive summarization")
	summarizerType := flag.String(
		"type",
		"extractive",
		"summarizer type (extractive, abstractive_openai, abstractive_huggingface, hybrid)",
	)

	// openai configuration
	openAIKey := flag.String("openai-key", os.Getenv("OPENAI_API_KEY"), "openai api key")
	openAIModel := flag.String("openai-model", os.Getenv("OPENAI_MODEL"), "openai model to use")
	openAIBaseURL := flag.String("openai-url", os.Getenv("OPENAI_URL"), "base url for openai api requests")

	// hugging face configuration
	huggingFaceKey := flag.String("hf-key", os.Getenv("HUGGING_FACE_KEY"), "hugging face api key")
	huggingFaceModel := flag.String("hf-model", "", "hugging face model to use")
	huggingFaceURL := flag.String("hf-url", "", "custom url for hugging face inference api")
	maxLength := flag.Int("max-length", 0, "maximum length for hugging face summary (0 for auto)")
	minLength := flag.Int("min-length", 0, "minimum length for hugging face summary (0 for auto)")

	flag.Parse()

	// create a configuration for the app from pased flags
	config := Config{
		FilePath: *filePath,
		SummaryRate: *summaryRate,
		TargetPercent: *targetPercent,
		SummarizerType: *summarizerType,
		OpenAIKey: *openAIKey,
		OpenAIModel: *openAIModel,
		OpenAIBaseURL: *openAIBaseURL,
		HuggingFaceKey: *huggingFaceKey,
		HuggingFaceModel: *huggingFaceModel,
		HuggingFaceURL: *huggingFaceURL,
		MaxLength: *maxLength,
		MinLength: *minLength,
	}

	if *apiMode {
		fmt.Printf("starting api server on port %s...\n", *port)
	} else {
		if config.FilePath == "" {
			fmt.Println("error: please provide a filepath with -flag file")
			flag.Usage()
			os.Exit(1)
		}
	}
}
