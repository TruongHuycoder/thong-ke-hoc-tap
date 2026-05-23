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
const groqModel = "llama3-70b-8192"

func init() {
	godotenv.Load()
	apiKey := "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"

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.groq.com/openai/v1"
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
			Model:    groqModel,
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
	if req.NumQuestions > 100 {
	    req.NumQuestions = 100
	}

	prompt := fmt.Sprintf(`Dựa vào nội dung văn bản sau, hãy tạo ra chính xác %d câu hỏi bài tập. 
Cố gắng phân bổ đều các loại câu hỏi: trắc nghiệm (mcq), đúng sai (tf), và điền khuyết (fitb).

LƯU Ý QUAN TRỌNG ĐỂ ĐẢM BẢO CHẤT LƯỢNG CÂU HỎI:
1. ĐỐI VỚI MÔN TOÁN / KHOA HỌC / KỸ THUẬT: 
   - Hãy dùng các ký hiệu văn bản tiêu chuẩn để tránh lỗi font (ví dụ: dùng "!=" hoặc chữ "khác" thay cho ký hiệu ≠, dùng ">=" thay cho ≥, "<=" thay cho ≤).
   - Tuyệt đối KHÔNG đục lỗ (tạo câu hỏi điền khuyết) vào giữa một công thức toán học, ký hiệu, hoặc các số liệu. 
   - Ví dụ TỐT: "______ bậc nhất một ẩn có dạng ax + b = 0" (đáp án: Phương trình).

2. ĐỐI VỚI TỪ VỰNG NGOẠI NGỮ:
   - TUYỆT ĐỐI KHÔNG bắt người dùng điền phiên âm hoặc đưa phiên âm vào phần đục lỗ "______".
   - Câu điền khuyết (fitb) phải yêu cầu điền từ đúng chính tả dựa vào nghĩa tiếng Việt hoặc ngữ cảnh.
   
3. QUY TẮC CHUNG CHO CÂU ĐIỀN KHUYẾT (FITB):
   - Chỉ đục lỗ những "từ khóa quan trọng" (khái niệm, định nghĩa, tên gọi). KHÔNG đục lỗ các từ nối, từ chỉ số lượng rời rạc.
   - Câu hỏi đục lỗ phải có ngữ cảnh rõ ràng, đủ để suy luận ra từ cần điền.
   - Đáp án cần điền (answerText) phải là cụm từ ngắn gọn (1-3 từ).

Hãy trả về một danh sách (mảng) JSON. Với mỗi câu hỏi, cấu trúc JSON cần chính xác như sau:
- Câu trắc nghiệm: { "type": "mcq", "q": "Câu hỏi?", "options": ["A", "B", "C", "D"], "answer": 0 } (answer là vị trí index đúng)
- Câu đúng sai: { "type": "tf", "q": "Câu hỏi?", "options": ["True", "False"], "answer": 1 } (answer là vị trí index đúng)
- Câu điền khuyết: { "type": "fitb", "q": "Nội dung câu hỏi yêu cầu điền từ ______", "answerText": "từ_cần_điền" }

Nội dung văn bản: %s`, req.NumQuestions, req.Text)

	resp, err := openaiClient.CreateChatCompletion(
		r.Context(),
		openai.ChatCompletionRequest{
			Model: groqModel,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(resp.Choices) > 0 {
		w.Write([]byte(resp.Choices[0].Message.Content))
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

	log.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
