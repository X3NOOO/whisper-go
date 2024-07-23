package main

import (
	"fmt"
	"os"

	"github.com/X3NOOO/whisper-go"
)

const FILENAME = "example.mp3"

func main() {
	file, err := os.ReadFile(FILENAME)
	if err != nil {
		panic(err)
	}

	client := whisper.New(os.Getenv("GROQ_API_KEY"))

	req := whisper.Request{
		File: whisper.File{
			Data: file,
			Name: FILENAME,
		},
		Model:          "whisper-large-v3",
		Temperature:    0.1,
		ResponseFormat: "text",
		Language:       "en",
	}

	resp, err := client.Transcribe(req)
	if err != nil {
		panic(err)
	}

	fmt.Println((*resp)["text"])
}
