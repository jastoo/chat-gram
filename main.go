package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

const apiURL = "https://cheapest-gpt-4-turbo-gpt-4-vision-chatgpt-openai-ai-api.p.rapidapi.com/v1/chat/completions"

type OpenAIRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	openAIKey := os.Getenv("RAPIDAPI_KEY")
	if openAIKey == "" {
		log.Fatal("RAPIDAPI_KEY is not set")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			responseText := getAIResponse(update.Message.Text, openAIKey)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseText)
			bot.Send(msg)
		}
	}
}

func getAIResponse(inputText, openAIKey string) string {
	messages := []Message{
		{
			Role:    "user",
			Content: inputText,
		},
	}

	reqBody := OpenAIRequest{
		Messages:    messages,
		Model:       "gpt-4-turbo-2024-04-09",
		MaxTokens:   100,
		Temperature: 0.9,
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("Failed to marshal request body: %v", err)
		return "Failed to process request"
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return "Failed to process request"
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-RapidAPI-Key", openAIKey)
	req.Header.Set("X-RapidAPI-Host", "cheapest-gpt-4-turbo-gpt-4-vision-chatgpt-openai-ai-api.p.rapidapi.com")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Request failed: %v", err)
		return "Failed to process request"
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return "Failed to process request"
	}

	log.Printf("Response body: %s", body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Non-OK HTTP status: %s", resp.Status)
		return "Failed to process request"
	}

	var aiResponse OpenAIResponse
	err = json.Unmarshal(body, &aiResponse)
	if err != nil {
		log.Printf("Failed to unmarshal response body: %v", err)
		return "Failed to process request"
	}

	if len(aiResponse.Choices) > 0 {
		return aiResponse.Choices[0].Message.Content
	}
	return "No response from AI"
}
