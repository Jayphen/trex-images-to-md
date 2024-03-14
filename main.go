package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type OpenAIResponse struct {
	Choices []struct {
		Message Message
	} `json:"choices"`
}

type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float32   `json:"temperature"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func callOpenAI(prompt string, apiKey string) (string, error) {
	apiURL := "https://api.openai.com/v1/chat/completions"

	data := OpenAIRequest{
		Model: "gpt-4",
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a formatter, spellchecker and grammar checker. Format the provided text without changing the wording, except to correct mistakes. Return the result as markdown.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.5,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	body := bytes.NewReader(payload)

	client := &http.Client{}
	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading API response: %s\n", err)
		return "", err
	}

	var openAIResp OpenAIResponse
	err = json.Unmarshal(respBody, &openAIResp)
	if err != nil {
		return "", err
	}

	if len(openAIResp.Choices) > 0 {
		return openAIResp.Choices[0].Message.Content, nil
	}

	return "", nil
}

func processFile(file string, apiKey string) {
	mdFilename := fmt.Sprintf("%s.md", strings.TrimSuffix(file, filepath.Ext(file)))

	cmd := exec.Command("trex", "-i")
	inputFile, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()

	cmd.Stdin = inputFile

	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	text, err := callOpenAI(string(output), apiKey)
	if err != nil {
		log.Printf("Error calling OpenAI API: %s\n", err)
		return
	}

	mdFile, err := os.Create(mdFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer mdFile.Close()

	writer := bufio.NewWriter(mdFile)
	defer writer.Flush()

	_, err = writer.WriteString(text)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created %s from %s\n", mdFilename, file)
}

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatal("You must provide one or more files as input")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY not set")
	}

	for _, file := range args {
		processFile(file, apiKey)
	}
}
