package connectors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const (
	defaultGeminiModel       = "gemini-1.5-flash"
	defaultGeminiTemperature = 0.7
)

// GeminiClient handles communication with the Google Gemini API
type GeminiClient struct {
	apiKey             string
	client             *genai.Client
	defaultModel       string
	defaultTemperature float32
	ctx                context.Context
}

// GeminiMessage represents a message in the chat completion request
type GeminiMessage struct {
	Role    string
	Content string
}

// GeminiResponse represents the response from the Gemini API
type GeminiResponse struct {
	Text         string
	FinishReason string
}

// NewGeminiClient creates a new Gemini client instance
func NewGeminiClient() (*GeminiClient, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY environment variable is not set")
	}

	// Validate API key format (basic check)
	if len(apiKey) < 20 {
		return nil, errors.New("GEMINI_API_KEY appears to be invalid (too short)")
	}

	// Read default model from env or use constant
	model := os.Getenv("GEMINI_DEFAULT_MODEL")
	if model == "" {
		model = defaultGeminiModel
		log.Printf("Using default Gemini model: %s", model)
	} else {
		log.Printf("Using configured Gemini model: %s", model)
	}

	// Read default temperature from env or use constant
	var temperature float32 = defaultGeminiTemperature
	if tempStr := os.Getenv("GEMINI_DEFAULT_TEMPERATURE"); tempStr != "" {
		if temp, err := strconv.ParseFloat(tempStr, 32); err == nil {
			temperature = float32(temp)
			log.Printf("Using configured Gemini temperature: %f", temperature)
		} else {
			log.Printf("Error parsing GEMINI_DEFAULT_TEMPERATURE, using default: %f", defaultGeminiTemperature)
		}
	} else {
		log.Printf("Using default Gemini temperature: %f", temperature)
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("error creating Gemini client: %w", err)
	}

	log.Printf("Gemini client initialized successfully")

	return &GeminiClient{
		apiKey:             apiKey,
		client:             client,
		defaultModel:       model,
		defaultTemperature: temperature,
		ctx:                ctx,
	}, nil
}

// CreateUnstructuredChatCompletion sends a chat completion request to the Gemini API
func (c *GeminiClient) CreateUnstructuredChatCompletion(messages []GeminiMessage, model string, temperature *float32) (*GeminiResponse, error) {
	// Use default model if not provided
	if model == "" {
		model = c.defaultModel
	}

	// Use default temperature if not provided
	temp := c.defaultTemperature
	if temperature != nil {
		temp = *temperature
	}

	// Create a generative model
	genModel := c.client.GenerativeModel(model)
	tempFloat := float32(temp)
	genModel.Temperature = &tempFloat

	// Convert messages to Gemini format
	var prompt string
	for _, msg := range messages {
		rolePrefix := ""
		if msg.Role != "" {
			rolePrefix = fmt.Sprintf("%s: ", msg.Role)
		}
		prompt += rolePrefix + msg.Content + "\n"
	}

	// Generate content
	resp, err := genModel.GenerateContent(c.ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("error generating content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("no response generated")
	}

	// Extract text from response
	responseText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			responseText += string(text)
		}
	}

	finishReason := "unknown"
	if resp.Candidates[0].FinishReason != 0 {
		finishReason = resp.Candidates[0].FinishReason.String()
	}

	return &GeminiResponse{
		Text:         responseText,
		FinishReason: finishReason,
	}, nil
}

// cleanMarkdownCodeBlocks removes markdown code blocks and backticks from a string
func cleanMarkdownCodeBlocks(input string) string {
	// Check if the input is empty
	if input == "" {
		return ""
	}

	// Handle various markdown code block formats
	// Case 1: ```json\n...\n```
	if strings.HasPrefix(input, "```") {
		// Find the end of the code block
		endIndex := strings.LastIndex(input, "```")
		if endIndex > 3 {
			// Extract content between the markers
			startContent := strings.Index(input, "\n")
			if startContent != -1 && startContent < endIndex {
				return strings.TrimSpace(input[startContent+1 : endIndex])
			}
		}
	}

	// Case 2: `{...}`
	if strings.HasPrefix(input, "`{") && strings.HasSuffix(input, "}`") {
		return strings.TrimSpace(input[1 : len(input)-1])
	}

	// Case 3: Just remove all backticks
	cleanedText := strings.Replace(input, "`", "", -1)

	// Remove any "json" language identifier that might be present
	cleanedText = strings.Replace(cleanedText, "json", "", -1)

	// Trim whitespace
	cleanedText = strings.TrimSpace(cleanedText)

	return cleanedText
}

// CreateStructuredChatCompletion sends a chat completion request to the Gemini API and expects a JSON response
func (c *GeminiClient) CreateStructuredChatCompletion(messages []GeminiMessage, model string, temperature *float32, responseType interface{}) error {
	// Add system message to ensure JSON response
	jsonFormatMessage := GeminiMessage{
		Role: "system",
		Content: `You are a structured data assistant that ONLY returns raw JSON. Follow these rules strictly:
1. Return ONLY the raw JSON object with no additional text, formatting, or explanation
2. DO NOT use markdown formatting, code blocks, or backticks in your response
3. DO NOT prefix your response with any markdown code block indicators
4. DO NOT wrap your JSON in any special characters
5. For numeric fields, always use numbers not strings (e.g., "temperature": 25 not "temperature": "25")
6. When asked for arrays, always provide at least one item
7. Ensure all required fields are present in the response
8. Your entire response should be valid JSON that can be parsed directly`,
	}

	fullMessages := make([]GeminiMessage, 0, len(messages)+1)
	fullMessages = append(fullMessages, jsonFormatMessage)
	fullMessages = append(fullMessages, messages...)

	// Use default model if not provided
	if model == "" {
		model = c.defaultModel
		log.Printf("Using default model for Gemini structured completion: %s", model)
	} else {
		log.Printf("Using provided model for Gemini structured completion: %s", model)
	}

	// Use default temperature if not provided
	temp := c.defaultTemperature
	if temperature != nil {
		temp = *temperature
		log.Printf("Using provided temperature for Gemini structured completion: %f", temp)
	} else {
		log.Printf("Using default temperature for Gemini structured completion: %f", temp)
	}

	log.Printf("Sending structured completion request to Gemini")

	// Get unstructured response first
	response, err := c.CreateUnstructuredChatCompletion(fullMessages, model, temperature)
	if err != nil {
		log.Printf("Error generating structured content with Gemini: %v", err)
		return fmt.Errorf("error generating structured content: %w", err)
	}

	log.Printf("Successfully received Gemini response, parsing JSON content")

	// Clean the response text by removing markdown code blocks if present
	cleanedResponse := cleanMarkdownCodeBlocks(response.Text)
	log.Printf("Cleaned response: %s", cleanedResponse)

	// Parse the JSON response
	if err := json.Unmarshal([]byte(cleanedResponse), responseType); err != nil {
		log.Printf("Error parsing Gemini JSON response: %v\nResponse content: %s",
			err, response.Text)
		return fmt.Errorf("error parsing JSON response from Gemini: %w\nResponse content: %s",
			err, response.Text)
	}

	log.Printf("Successfully parsed Gemini JSON response")
	return nil
}

// Close closes the Gemini client
func (c *GeminiClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
