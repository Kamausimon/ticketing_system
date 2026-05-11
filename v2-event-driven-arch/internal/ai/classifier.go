package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ClassifierService handles AI-based ticket classification
type ClassifierService struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
}

// NewClassifierService creates a new AI classifier service
func NewClassifierService() *ClassifierService {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("⚠️  OPENAI_API_KEY not set - AI classification will be disabled")
		return nil
	}

	return &ClassifierService{
		apiKey: apiKey,
		apiURL: "https://api.openai.com/v1/chat/completions",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TicketClassification represents the AI's classification result
type TicketClassification struct {
	Priority   string  `json:"priority"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

// OpenAIRequest represents a request to OpenAI API
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents a response from OpenAI API
type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a response choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents API usage statistics
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ClassifyTicket analyzes a support ticket and suggests priority level
func (s *ClassifierService) ClassifyTicket(subject, description, category string, orderID, eventID *uint) (*TicketClassification, error) {
	if s == nil {
		return nil, fmt.Errorf("AI classifier not available - OPENAI_API_KEY not configured")
	}

	// Build context information
	context := s.buildContext(category, orderID, eventID)

	// Create the prompt
	prompt := s.buildPrompt(subject, description, context)

	// Call OpenAI API
	response, err := s.callOpenAI(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	// Parse the response
	classification, err := s.parseResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return classification, nil
}

// buildContext creates contextual information for classification
func (s *ClassifierService) buildContext(category string, orderID, eventID *uint) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("Category: %s", category))

	if orderID != nil {
		parts = append(parts, fmt.Sprintf("Related to Order #%d", *orderID))
	}

	if eventID != nil {
		parts = append(parts, fmt.Sprintf("Related to Event #%d", *eventID))
	}

	return strings.Join(parts, " | ")
}

// buildPrompt creates the classification prompt
func (s *ClassifierService) buildPrompt(subject, description, context string) string {
	return fmt.Sprintf(`You are a support ticket classifier. Analyze the following support ticket and classify its priority level.

Context: %s
Subject: %s
Description: %s

Priority Levels:
- critical: Payment failures, security issues, event cancellations, complete service outage (requires immediate action within 15 minutes)
- high: Login problems, booking errors, issues affecting upcoming events, partial service disruption (requires response within 2 hours)
- medium: Feature questions, minor bugs, account changes, general inquiries (response within 24 hours)
- low: Suggestions, feedback, documentation requests, non-urgent questions (response within 72 hours)

Respond ONLY with valid JSON in this exact format (no markdown, no code blocks):
{
  "priority": "critical|high|medium|low",
  "confidence": 0.95,
  "reasoning": "Brief explanation of why this priority was chosen"
}`, context, subject, description)
}

// callOpenAI makes the API call to OpenAI
func (s *ClassifierService) callOpenAI(prompt string) (*OpenAIResponse, error) {
	requestBody := OpenAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a highly accurate support ticket classifier. Always respond with valid JSON only, no additional text or formatting.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.3, // Lower temperature for more consistent results
		MaxTokens:   200,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &openAIResp, nil
}

// parseResponse extracts classification from OpenAI response
func (s *ClassifierService) parseResponse(response *OpenAIResponse) (*TicketClassification, error) {
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := response.Choices[0].Message.Content

	// Clean the response - remove markdown code blocks if present
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var classification TicketClassification
	if err := json.Unmarshal([]byte(content), &classification); err != nil {
		return nil, fmt.Errorf("failed to parse classification JSON: %w (content: %s)", err, content)
	}

	// Validate priority
	validPriorities := map[string]bool{
		"critical": true,
		"high":     true,
		"medium":   true,
		"low":      true,
	}

	if !validPriorities[classification.Priority] {
		return nil, fmt.Errorf("invalid priority: %s", classification.Priority)
	}

	// Ensure confidence is between 0 and 1
	if classification.Confidence < 0 {
		classification.Confidence = 0
	}
	if classification.Confidence > 1 {
		classification.Confidence = 1
	}

	return &classification, nil
}

// ClassifyAsync runs classification in the background (non-blocking)
func (s *ClassifierService) ClassifyAsync(subject, description, category string, orderID, eventID *uint, callback func(*TicketClassification, error)) {
	go func() {
		classification, err := s.ClassifyTicket(subject, description, category, orderID, eventID)
		callback(classification, err)
	}()
}
