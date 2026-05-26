//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func main() {
	godotenv.Load()
	client, err := genai.NewClient(context.Background(), option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	iter := client.ListModels(context.Background())
	for {
		m, err := iter.Next()
		if err != nil {
			break
		}
		fmt.Println(m.Name)
	}
}
