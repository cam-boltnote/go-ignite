package connectors

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	defaultModel       = "gpt-3.5-turbo"
	defaultTemperature = 0.7
)

// OpenAIClient handles communication with the OpenAI API
type OpenAIClient struct {
	apiKey             string
	httpClient         *http.Client
	baseURL            string
	defaultModel       string
	defaultTemperature float32
}

// ChatMessage represents a message in the chat completion request
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents the request structure for chat completions
type ChatCompletionRequest struct {
	Model          string          `json:"model"`
	Messages       []ChatMessage   `json:"messages"`
	Temperature    float32         `json:"temperature"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

// ChatCompletionResponse represents the response from the chat completion API
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Add a new struct for the response format
type ResponseFormat struct {
	Type string `json:"type"`
}

// NewOpenAIClient creates a new OpenAI client instance
func NewOpenAIClient() (*OpenAIClient, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY environment variable is not set")
	}

	// Validate API key format (basic check)
	if len(apiKey) < 20 {
		return nil, errors.New("OPENAI_API_KEY appears to be invalid (too short)")
	}

	// Read default model from env or use constant
	model := os.Getenv("OPENAI_DEFAULT_MODEL")
	if model == "" {
		model = defaultModel
		log.Printf("Using default OpenAI model: %s", model)
	} else {
		log.Printf("Using configured OpenAI model: %s", model)
	}

	// Read default temperature from env or use constant
	var temperature float32 = defaultTemperature
	if tempStr := os.Getenv("OPENAI_DEFAULT_TEMPERATURE"); tempStr != "" {
		if temp, err := strconv.ParseFloat(tempStr, 32); err == nil {
			temperature = float32(temp)
			log.Printf("Using configured OpenAI temperature: %f", temperature)
		} else {
			log.Printf("Error parsing OPENAI_DEFAULT_TEMPERATURE, using default: %f", defaultTemperature)
		}
	} else {
		log.Printf("Using default OpenAI temperature: %f", temperature)
	}

	log.Printf("OpenAI client initialized successfully")

	return &OpenAIClient{
		apiKey:             apiKey,
		httpClient:         &http.Client{},
		baseURL:            "https://api.openai.com/v1",
		defaultModel:       model,
		defaultTemperature: temperature,
	}, nil
}

// CreateChatCompletion sends a chat completion request to the OpenAI API
func (c *OpenAIClient) CreateChatCompletion(messages []ChatMessage, model string, temperature float32) (*ChatCompletionResponse, error) {
	if model == "" {
		model = "gpt-4" // Default to GPT-4
	}

	reqBody := ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", c.baseURL), bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result ChatCompletionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &result, nil
}

// CreateUnstructuredChatCompletion sends a chat completion request to the OpenAI API
func (c *OpenAIClient) CreateUnstructuredChatCompletion(messages []ChatMessage, model string, temperature *float32) (*ChatCompletionResponse, error) {
	// Use default model if not provided
	if model == "" {
		model = c.defaultModel
	}

	// Use default temperature if not provided
	temp := c.defaultTemperature
	if temperature != nil {
		temp = *temperature
	}

	reqBody := ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temp,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", c.baseURL), bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result ChatCompletionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &result, nil
}

// CreateStructuredChatCompletion sends a chat completion request and expects a JSON response
func (c *OpenAIClient) CreateStructuredChatCompletion(messages []ChatMessage, model string, temperature *float32, responseType interface{}) error {
	// Add system message to ensure JSON response
	jsonFormatMessage := ChatMessage{
		Role: "system",
		Content: `You are a structured data assistant. Follow these rules strictly:
1. Always respond with valid JSON that matches the expected response type
2. Never include explanatory text - only return the JSON object
3. For numeric fields, always use numbers not strings (e.g., "temperature": 25 not "temperature": "25")
4. When asked for arrays, always provide at least the minimum number requested
5. Ensure all required fields are present in the response`,
	}

	fullMessages := make([]ChatMessage, 0, len(messages)+1)
	fullMessages = append(fullMessages, jsonFormatMessage)
	fullMessages = append(fullMessages, messages...)

	// Use default model if not provided
	if model == "" {
		model = c.defaultModel
		log.Printf("Using default model for structured completion: %s", model)
	}

	// Use default temperature if not provided
	temp := c.defaultTemperature
	if temperature != nil {
		temp = *temperature
		log.Printf("Using provided temperature for structured completion: %f", temp)
	} else {
		log.Printf("Using default temperature for structured completion: %f", temp)
	}

	// Create request with JSON response format
	reqBody := ChatCompletionRequest{
		Model:       model,
		Messages:    fullMessages,
		Temperature: temp,
		ResponseFormat: &ResponseFormat{
			Type: "json_object",
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("Error marshaling OpenAI request: %v", err)
		return fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", c.baseURL), bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("Error creating OpenAI request: %v", err)
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	log.Printf("Sending structured completion request to OpenAI")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("Error making OpenAI request: %v", err)
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading OpenAI response body: %v", err)
		return fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("OpenAI API request failed with status %d: %s", resp.StatusCode, string(body))
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		log.Printf("Error unmarshaling OpenAI chat response: %v", err)
		return fmt.Errorf("error unmarshaling chat response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		log.Printf("OpenAI returned no response choices")
		return errors.New("no response choices returned")
	}

	log.Printf("Successfully received OpenAI response, parsing JSON content")

	// The response should already be valid JSON since we used response_format
	if err := json.Unmarshal([]byte(chatResp.Choices[0].Message.Content), responseType); err != nil {
		log.Printf("Error parsing OpenAI JSON response: %v\nResponse content: %s",
			err, chatResp.Choices[0].Message.Content)
		return fmt.Errorf("error parsing JSON response: %w\nResponse content: %s",
			err, chatResp.Choices[0].Message.Content)
	}

	log.Printf("Successfully parsed OpenAI JSON response")
	return nil
}

// CreateEmbedding generates embeddings for the given text using the OpenAI API
func (c *OpenAIClient) CreateEmbedding(input string) ([]float32, error) {
	reqBody := map[string]interface{}{
		"model": "text-embedding-ada-002",
		"input": input,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/embeddings", c.baseURL), bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	// Extract embeddings from the response
	data, ok := result["data"].([]interface{})
	if !ok || len(data) == 0 {
		return nil, errors.New("invalid embedding response format")
	}

	embedding, ok := data[0].(map[string]interface{})["embedding"].([]interface{})
	if !ok {
		return nil, errors.New("invalid embedding data format")
	}

	// Convert the embedding values to float32
	embeddings := make([]float32, len(embedding))
	for i, v := range embedding {
		f, ok := v.(float64)
		if !ok {
			return nil, errors.New("invalid embedding value format")
		}
		embeddings[i] = float32(f)
	}

	return embeddings, nil
}
