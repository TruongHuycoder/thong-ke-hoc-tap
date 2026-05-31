package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

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

	// Smart Math Detection: count math symbols in text
	mathSymbols := []string{"∫", "∑", "∂", "√", "π", "α", "β", "γ", "θ", "λ", "μ", "σ", "φ", "Δ", "≤", "≥", "≠", "≈", "∞", "²", "³", "⁴", "dx", "dy", "lim", "sin", "cos", "tan", "log", "ln"}
	mathScore := 0
	for _, sym := range mathSymbols {
		if strings.Contains(text, sym) {
			mathScore++
		}
	}
	// Also count = signs and numbers ratio
	for _, ch := range text {
		if ch == '=' || ch == '+' || ch == '-' || ch == '/' || ch == '%' {
			mathScore++
		}
	}

	var prompt string

	if mathScore >= 8 {
		// MATH MODE: Specialized prompt for mathematical content
		prompt = fmt.Sprintf(`Bạn là chuyên gia tạo câu hỏi kiểm tra TOÁN HỌC. Phân tích văn bản sau và tạo ĐÚNG %d câu hỏi.

QUY TẮC BẮT BUỘC CHO TOÁN:
1. CHỈ tạo câu MCQ và True/False (TF). KHÔNG tạo FITB vì dễ lỗi với công thức.
2. Hỏi về KHÁI NIỆM, ĐỊNH NGHĨA, ĐỊNH LÝ, ĐIỀU KIỆN ÁP DỤNG — KHÔNG yêu cầu tính toán.
3. Ví dụ câu hỏi TỐT: "Điều kiện để tích phân hội tụ là gì?" / "Định lý nào phát biểu rằng..."
4. Ví dụ câu hỏi XẤU (TRÁNH): "Tính ∫x²dx = ?" (yêu cầu tính toán)
5. Các lựa chọn đáp án phải là CHỮ (khái niệm/tên định lý), KHÔNG phải kết quả số.
6. Phân bổ: 70%% MCQ, 30%% TF.

FORMAT JSON bắt buộc:
{"questions": [
  {"type":"mcq","q":"câu hỏi về khái niệm?","options":["A. khái niệm","B. khái niệm","C. khái niệm","D. khái niệm"],"answer":0,"explanation":"giải thích ngắn"},
  {"type":"tf","q":"Mệnh đề: [phát biểu định lý]","options":["True","False"],"answer":0,"explanation":"giải thích"}
]}

Văn bản toán học: %s`, req.NumQuestions, text)
	} else {
		// NORMAL MODE: General prompt for non-math content
		prompt = fmt.Sprintf(`Bạn là chuyên gia tạo câu hỏi học tập. Tạo ĐÚNG %d câu hỏi từ văn bản sau.
Phân bổ đều: trắc nghiệm (mcq), đúng sai (tf), điền khuyết (fitb).

QUY TẮC:
1. Câu hỏi phải dựa SÁT vào nội dung văn bản, không bịa thêm.
2. answerText của FITB phải là 1-3 từ quan trọng trong văn bản.
3. FITB: đục lỗ từ/cụm từ quan trọng, KHÔNG đục lỗ số hay ký hiệu.
4. Explanation giải thích ngắn gọn tại sao đáp án đúng.

FORMAT JSON bắt buộc:
{"questions": [
  {"type":"mcq","q":"câu hỏi?","options":["A. ...","B. ...","C. ...","D. ..."],"answer":0,"explanation":"..."},
  {"type":"tf","q":"Mệnh đề: ...","options":["True","False"],"answer":0,"explanation":"..."},
  {"type":"fitb","q":"... _______ ...","answerText":"từ khóa","explanation":"..."}
]}

Văn bản: %s`, req.NumQuestions, text)
	}

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
