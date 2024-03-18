package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var sysPrompt string

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

func callOpenAI(prompt string) (string, error) {
	apiURL := "https://api.openai.com/v1/chat/completions"

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY not set")
	}

	data := OpenAIRequest{
		Model: "gpt-4",
		Messages: []Message{
			{
				Role:    "system",
				Content: sysPrompt,
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

	log.Printf("🤖 Sending request to OpenAI API\n")
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

func processFile(file string) {
	mdFilename := fmt.Sprintf("%s.md", strings.TrimSuffix(file, filepath.Ext(file)))
	cmd := exec.Command("trex", "-i")

	log.Printf("Reading file %s\n", file)
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

	text, err := callOpenAI(string(output))
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
	flag.StringVar(&sysPrompt, "sysprompt", "You are a formatter, spellchecker and grammar checker. Format the provided text without changing the wording, except to correct mistakes. Return the result as markdown.", "Prompt to send to OpenAI")

	flag.Parse()

	args := flag.Args()
	wg := sync.WaitGroup{}

	if len(args) < 1 {
		log.Fatal("You must provide one or more files as input")
	}

	for _, file := range args {
		wg.Add(1)
		go func() {
			defer wg.Done()
			processFile(file)
		}()

	}

	wg.Wait()
}
