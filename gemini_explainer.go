package main

import (
	"context"
	"fmt"

	genai "github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiExplainer handles explanation generation using Gemini
type GeminiExplainer struct {
	client *genai.Client
	model  string
}

// NewGeminiExplainer creates a new Gemini explainer client
func NewGeminiExplainer(apiKey string) (*GeminiExplainer, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %v", err)
	}

	return &GeminiExplainer{
		client: client,
		model:  "models/gemini-2.0-flash",
	}, nil
}

// GenerateExplanation sends a prompt to Gemini and returns the explanation text
func (g *GeminiExplainer) GenerateExplanation(prompt string) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	ctx := context.Background()
	model := g.client.GenerativeModel(g.model)

	// Configure model for natural language generation
	model.SetTemperature(0.7) // Higher temperature for more natural explanations
	model.SetTopK(40)
	model.SetTopP(0.95)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate explanation: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	// Extract text response
	explanation := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			explanation += string(text)
		}
	}

	if explanation == "" {
		return "", fmt.Errorf("empty response from Gemini")
	}

	return explanation, nil
}

// Close closes the Gemini client
func (g *GeminiExplainer) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}
