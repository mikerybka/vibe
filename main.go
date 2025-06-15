package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const openaiURL = "https://api.openai.com/v1/chat/completions"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float32   `json:"temperature"`
	Stop        []string  `json:"stop"`
}

type Choice struct {
	Message Message `json:"message"`
}

type ResponseBody struct {
	Choices []Choice `json:"choices"`
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: vibe <output.go> <prompt>")
		os.Exit(1)
	}
	outputFile := os.Args[1]
	userPrompt := os.Args[2]

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENAI_API_KEY is not set")
		os.Exit(1)
	}

	reqBody := RequestBody{
		Model: "gpt-4.1",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant that ONLY returns valid, self-contained Go code. Do not include explanations or markdown formatting."},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0,
		Stop:        []string{"```", "<!--"},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Println("Failed to serialize request body:", err)
		os.Exit(1)
	}

	req, err := http.NewRequest("POST", openaiURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		fmt.Println("Failed to create HTTP request:", err)
		os.Exit(1)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("HTTP request failed:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		fmt.Printf("API error: %s\n", data)
		os.Exit(1)
	}

	var respBody ResponseBody
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		fmt.Println("Failed to decode response:", err)
		os.Exit(1)
	}

	if len(respBody.Choices) == 0 {
		fmt.Println("No choices returned")
		os.Exit(1)
	}

	code := strings.TrimSpace(respBody.Choices[0].Message.Content)

	err = os.WriteFile(outputFile, []byte(code + "\n"), 0644)
	if err != nil {
		fmt.Printf("Failed to write file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Go code written to", outputFile)
}


