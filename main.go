package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

var openaiClient *openai.Client
var systemPrompt string
var aiModel string

func init() {
	godotenv.Load()
	
	var apiKey string
	var baseURL string

	if xaiKey := os.Getenv("XAI_API_KEY"); xaiKey != "" {
		apiKey = xaiKey
		baseURL = "https://api.x.ai/v1"
		aiModel = "grok-beta"
	} else if groqKey := os.Getenv("GROQ_API_KEY"); groqKey != "" {
		apiKey = groqKey
		baseURL = "https://api.groq.com/openai/v1"
		aiModel = "llama-3.3-70b-versatile"
	} else {
		apiKey = "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"
		baseURL = "https://api.groq.com/openai/v1"
		aiModel = "llama-3.3-70b-versatile"
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	openaiClient = openai.NewClientWithConfig(config)

	knowledge, _ := os.ReadFile("knowledge.txt")
	rules, _ := os.ReadFile("rules.txt")
	systemPrompt = string(rules) + "\n\nNGỮ CẢNH DỰ ÁN:\n" + string(knowledge)
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type ChatMessage struct {
		Role string `json:"role"`
		Text string `json:"text"`
	}

	var req struct {
		History []ChatMessage `json:"history"`
		Message string        `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
	}

	for _, msg := range req.History {
		role := openai.ChatMessageRoleUser
		if msg.Role == "model" || msg.Role == "assistant" {
			role = openai.ChatMessageRoleAssistant
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Text,
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.Message,
	})

	stream, err := openaiClient.CreateChatCompletionStream(
		r.Context(),
		openai.ChatCompletionRequest{
			Model:    aiModel,
			Messages: messages,
			Stream:   true,
		},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)
	flusher, ok := w.(http.Flusher)

	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Printf("Chat stream error: %v", err)
			fmt.Fprintf(w, " [Lỗi mạng hoặc API Limit: %v]", err)
			break
		}
		fmt.Fprint(w, resp.Choices[0].Delta.Content)
		if ok {
			flusher.Flush()
		}
	}
}

func handleGenerateQuiz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Text string `json:"text"`
		NumQuestions int `json:"num_questions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if req.NumQuestions <= 0 {
	    req.NumQuestions = 3
	}
	if req.NumQuestions > 20 {
	    req.NumQuestions = 20
	}

	// Truncate text to avoid token overflow on Groq Free tier
	text := req.Text
	if len(text) > 3000 {
		text = text[:3000]
	}

	prompt := fmt.Sprintf(`Tạo %d câu hỏi từ văn bản sau. Phân bổ đều: trắc nghiệm (mcq), đúng sai (tf), điền khuyết (fitb).

Trả về JSON object có key "questions" chứa mảng câu hỏi. Mỗi câu hỏi theo format:
- MCQ: {"type":"mcq","q":"...","options":["A","B","C","D"],"answer":0,"explanation":"..."}
- TF: {"type":"tf","q":"...","options":["True","False"],"answer":0,"explanation":"..."}
- FITB: {"type":"fitb","q":"... ______","answerText":"từ","explanation":"..."}

Quy tắc: Không đục lỗ công thức toán. Không dùng phiên âm. answerText phải 1-3 từ.

Văn bản: %s`, req.NumQuestions, text)

	resp, err := openaiClient.CreateChatCompletion(
		r.Context(),
		openai.ChatCompletionRequest{
			Model: aiModel,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)
	if err != nil {
		log.Printf("Quiz generation error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(resp.Choices) > 0 {
		content := resp.Choices[0].Message.Content
		// Extract the "questions" array from JSON Object response
		var result map[string]json.RawMessage
		if err := json.Unmarshal([]byte(content), &result); err == nil {
			if questions, ok := result["questions"]; ok {
				w.Write(questions)
				return
			}
		}
		// Fallback: return raw content
		w.Write([]byte(content))
	} else {
		w.Write([]byte(`[]`))
	}
}

func main() {
	initAuth()

	fs := http.FileServer(http.Dir("./freshman-bridge-dashboard"))
	http.Handle("/", fs)

	http.HandleFunc("/api/chat", handleChat)
	http.HandleFunc("/api/generate_quiz", handleGenerateQuiz)

	// Auth & Data Endpoints
	http.HandleFunc("/api/register", handleRegister)
	http.HandleFunc("/api/login", handleLogin)
	http.HandleFunc("/api/data", handleData)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Server is running on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
