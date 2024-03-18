package main

import (
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

var (
	sysPrompt string
	model     string
	ocrMutex  sync.Mutex
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

func callOpenAI(prompt string) (string, error) {
	apiURL := "https://api.openai.com/v1/chat/completions"

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY not set")
	}

	data := OpenAIRequest{
		Model: model,
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

	log.Printf("🤖 Sending request to OpenAI API with OCR'd text:\n%s\n", prompt)
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

func openInputFile(filePath string) (*os.File, error) {
	inputFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", filePath, err)
	}

	return inputFile, nil
}

func executeOCR(inputFile *os.File) ([]byte, error) {
	ocrMutex.Lock()
	defer ocrMutex.Unlock()
	cmd := exec.Command("trex", "-i")
	cmd.Stdin = inputFile

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error executing OCR: %w", err)
	}
	return output, nil
}

func writeToMarkdownFile(mdFilename string, content []byte) error {
	mdFile, err := os.Create(mdFilename)
	if err != nil {
		return fmt.Errorf("error creating markdown file %s: %w", mdFilename, err)
	}
	defer mdFile.Close()

	if _, err := mdFile.Write(content); err != nil {
		return fmt.Errorf("error writing markdown file %s: %w", mdFilename, err)
	}

	return nil
}

func processFile(file string, wg *sync.WaitGroup) {
	defer wg.Done()
	inputFile, err := openInputFile(file)
	if err != nil {
		log.Println(err)
		return
	}
	defer inputFile.Close()

	output, err := executeOCR(inputFile)
	if err != nil {
		log.Println(err)
		return
	}

	text, err := callOpenAI(string(output))
	if err != nil {
		log.Printf("Error calling OpenAI API: %s\n", err)
		return
	}

	mdFilename := fmt.Sprintf("%s.md", strings.TrimSuffix(file, filepath.Ext(file)))
	if err := writeToMarkdownFile(mdFilename, []byte(text)); err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("Created %s from %s\n", mdFilename, file)
}

func main() {
	flag.StringVar(&sysPrompt, "sysprompt", "You are a formatter, spellchecker and grammar checker. Format the provided text without changing the wording, except to correct mistakes. Return the result as markdown.", "Prompt to send to OpenAI")
	flag.StringVar(&model, "model", "gpt-4", "Which GPT model to use. Recommended: 'gpt-4', 'gpt-3.5-turbo'")
	flag.Parse()

	args := flag.Args()
	var wg sync.WaitGroup

	if len(args) < 1 {
		log.Fatal("You must provide one or more files as input")
	}

	for _, file := range args {
		wg.Add(1)
		go processFile(file, &wg)
	}

	wg.Wait()
}
