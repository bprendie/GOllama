package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Request struct defines the structure of a request sent to the Ollama API.
type Request struct {
	Model             string    `json:"model"`               // The name of the model to use for generating responses.
	Messages          []Message `json:"messages"`            // A slice of Message structs representing the conversation history.
	Stream            bool      `json:"stream"`              // A boolean flag indicating whether to stream the response.
	ContextWindowSize int       `json:"context_window_size"` // The size of the context window for the conversation.
}

// Message struct defines the structure of a single message in the chat.
type Message struct {
	Role    string `json:"role"`    // The role of the message sender, e.g., "user" or "system".
	Content string `json:"content"` // The content of the message.
}

// Response struct defines the structure of a response received from the Ollama API.
type Response struct {
	Model              string    `json:"model"`                // The name of the model used.
	CreatedAt          time.Time `json:"created_at"`           // The timestamp when the response was created.
	Message            Message   `json:"message"`              // The generated message.
	Done               bool      `json:"done"`                 // A boolean flag indicating whether the response is complete.
	TotalDuration      int64     `json:"total_duration"`       // The total duration of the response generation.
	LoadDuration       int       `json:"load_duration"`        // The duration of loading the model.
	PromptEvalCount    int       `json:"prompt_eval_count"`    // The number of prompt evaluations.
	PromptEvalDuration int       `json:"prompt_eval_duration"` // The duration of prompt evaluations.
	EvalCount          int       `json:"eval_count"`           // The number of evaluations.
	EvalDuration       int64     `json:"eval_duration"`        // The duration of evaluations.
}

// Config struct defines the structure of the configuration file.
type Config struct {
	OllamaURL         string `json:"ollamaURL"`
	OllamaPort        int    `json:"ollamaPort"`
	ModelName         string `json:"modelName"`
	ContextWindowSize int    `json:"contextWindowSize"`
	HumanName         string `json:"humanName"`
	AIName            string `json:"AIName"`
	SystemPrompt      string `json:"systemPrompt"`
}

// LoadConfig function reads the configuration from a JSON file.
func LoadConfig(filename string) (Config, error) {
	var config Config
	configFile, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	return config, err
}

// talkToOllama function sends a request to the Ollama API and returns the response.
func talkToOllama(url string, ollamaReq Request) (*Response, error) {
	// Marshal the request struct into JSON format.
	js, err := json.Marshal(&ollamaReq)
	if err != nil {
		return nil, err // Return an error if marshaling fails.
	}

	// Create a new HTTP client.
	client := http.Client{}

	// Create a new HTTP POST request with the JSON data.
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(js))
	if err != nil {
		return nil, err // Return an error if creating the request fails.
	}

	// Send the HTTP request and get the response.
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, err // Return an error if sending the request fails.
	}
	defer httpResp.Body.Close() // Ensure the response body is closed after reading.

	// Decode the JSON response into the Response struct.
	ollamaResp := Response{}
	err = json.NewDecoder(httpResp.Body).Decode(&ollamaResp)
	return &ollamaResp, err // Return the response and any error that occurred.
}

func main() {
	// Define a command-line flag for the configuration file path.
	configPath := flag.String("config", "llama.json", "Path to the configuration file")
	flag.Parse()

	// Load the configuration from the config file.
	config, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Press 'Enter' to use default values from config or 'y' to enter custom values.")
	if scanner.Scan() {
		input := scanner.Text()
		if strings.ToLower(input) == "y" {
			// Prompt for custom human name.
			fmt.Print("Enter custom human name: ")
			if scanner.Scan() {
				config.HumanName = scanner.Text()
			}

			// Prompt for custom AI name.
			fmt.Print("Enter custom AI name: ")
			if scanner.Scan() {
				config.AIName = scanner.Text()
			}

			// Prompt for custom system prompt.
			fmt.Print("Enter custom system prompt: ")
			if scanner.Scan() {
				config.SystemPrompt = scanner.Text()
			}
		}
	}

	// Define the system prompt message with the human name and AI name.
	systemPromptWithNames := fmt.Sprintf("%s Your name is %s. My name is %s.", config.SystemPrompt, config.HumanName, config.AIName)
	systemMsg := Message{
		Role:    "system",
		Content: systemPromptWithNames,
	}

	// Initialize a slice to store the conversation history.
	var conversationHistory []Message
	conversationHistory = append(conversationHistory, systemMsg)

	// Enter an infinite loop to continuously read user input.
	for {
		fmt.Printf("%s: ", config.HumanName)
		if !scanner.Scan() {
			break // Exit the loop if reading input fails.
		}
		input := scanner.Text() // Get the user input from the scanner.
		if input == "exit" {
			break // Exit the loop if the user types "exit".
		}

		// Define the user message and add it to the conversation history.
		userMsg := Message{
			Role:    "user",
			Content: input,
		}
		conversationHistory = append(conversationHistory, userMsg)

		// Create a new request with the model name, messages, streaming flag, and context window size.
		req := Request{
			Model:             config.ModelName,
			Stream:            false,
			Messages:          conversationHistory, // Use the conversation history as context.
			ContextWindowSize: config.ContextWindowSize,
		}

		// Send the request to the Ollama API and get the response.
		resp, err := talkToOllama(config.OllamaURL, req)
		if err != nil {
			fmt.Printf("Error sending message: %v\n", err) // Print an error message if sending the request fails.
			continue                                       // Continue to the next iteration of the loop.
		}

		// Print the AI's response and add it to the conversation history.
		fmt.Printf("%s: %s\n", config.AIName, resp.Message.Content)
		conversationHistory = append(conversationHistory, resp.Message)
	}
}
