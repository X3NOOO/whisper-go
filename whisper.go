package whisper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)


const (
	defaultApiURL = "https://api.groq.com/openai/v1/audio/transcriptions"
)

type File struct {
	Data []byte
	Name string
}

type Request struct {
	File 		  	File
	Model           string
	Temperature     float64
	ResponseFormat  string
	Language        string
}

type Response map[string]any // As some APIs return custom parameters alongside the response (see "x_groq") we cannot statically define this type.

type Whisper struct {
	apiKey string
	apiURL string
}

func New(apiKey string) *Whisper {
	return &Whisper{
		apiKey: apiKey,
		apiURL: defaultApiURL,
	}
}

func (w *Whisper) SetApiURL(apiURL string) {
	w.apiURL = apiURL
}

/*
Transcribe the given audio file using the specified model and parameters.

Warning: even if the "response_format" is set to "text", the return type is still a map. The result is then stored under the ["text"] key.
*/
func (w *Whisper) Transcribe(req Request) (*Response, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", req.File.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	n, err := part.Write(req.File.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to write file data: %w", err)
	}

	if n != len(req.File.Data) {
		return nil, fmt.Errorf("failed to write all file data")
	}

	_ = writer.WriteField("model", req.Model)
	_ = writer.WriteField("temperature", fmt.Sprintf("%f", req.Temperature))
	_ = writer.WriteField("response_format", req.ResponseFormat)
	_ = writer.WriteField("language", req.Language)

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	httpReq, err := http.NewRequest("POST", w.apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+w.apiKey)
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if len(response) == 0 {
		return nil, fmt.Errorf("empty response")
	}
	
	if response[0] != '{' { // dumb json check
		return &Response{
			"text": string(response),
		}, nil
	}

	var transcriptionResp Response
	err = json.NewDecoder(resp.Body).Decode(&transcriptionResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &transcriptionResp, nil
}