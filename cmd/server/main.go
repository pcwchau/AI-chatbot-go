package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type ChatRequest struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

type ExternalAPIRequest struct {
	Model    string `json:"model"`
	Store    bool   `json:"store"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

type ExternalAPIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
}

func main() {
	http.HandleFunc("/chat", chatHandler)

	// Start the HTTP server on port 8080
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	// Set the content type to application/json
	w.Header().Set("Content-Type", "application/json")

	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse the JSON body into ChatRequest struct
	var chatRequest ChatRequest
	if err := json.Unmarshal(body, &chatRequest); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Prepare the JSON body for the external API request
	apiRequest := ExternalAPIRequest{
		Model: "gpt-4o-mini",
		Store: true,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    "user",
				Content: chatRequest.Message,
			},
		},
	}

	// Encode the API request into JSON
	apiRequestBody, err := json.Marshal(apiRequest)
	if err != nil {
		http.Error(w, "Failed to create request body for external API", http.StatusInternalServerError)
		return
	}

	// Send the HTTP POST request to the external API
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(apiRequestBody))
	if err != nil {
		http.Error(w, "Failed to create request to external API", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer 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")

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request to external API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response from the external API
	apiResponseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response from external API", http.StatusInternalServerError)
		return
	}

	// Parse the response from the external API to extract the content field
	var apiResponse ExternalAPIResponse
	if err := json.Unmarshal(apiResponseBody, &apiResponse); err != nil {
		http.Error(w, "Failed to parse response from external API", http.StatusInternalServerError)
		return
	}

	// Extract the content field from the response
	if len(apiResponse.Choices) > 0 {
		content := apiResponse.Choices[0].Message.Content

		// Return the extracted content in the response
		json.NewEncoder(w).Encode(ChatResponse{Reply: content})
	} else {
		http.Error(w, "No response content from external API", http.StatusInternalServerError)
	}
}
