package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func main() {
	url := "https://api.groq.com/openai/v1/chat/completions"
	payload := []byte(`{"model":"llama3-70b-8192","messages":[{"role":"user","content":"hello"}]}`)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer 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")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Status:", resp.Status)
	fmt.Println("Body:", string(body))
}
