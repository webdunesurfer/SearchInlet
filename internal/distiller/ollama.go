package distiller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type OllamaClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewOllamaClient(baseURL string) *OllamaClient {
	return &OllamaClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second, // Models take time to think
		},
	}
}

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type GenerateResponse struct {
	Response string `json:"response"`
}

type PullRequest struct {
	Model string `json:"model"`
}

type ListResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func (c *OllamaClient) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama list returned status %d", resp.StatusCode)
	}

	var listResp ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}

	var names []string
	for _, m := range listResp.Models {
		names = append(names, m.Name)
	}
	return names, nil
}

type PullProgress struct {
	Status    string `json:"status"`
	Digest    string `json:"digest"`
	Total     int64  `json:"total"`
	Completed int64  `json:"completed"`
}

func (c *OllamaClient) PullModel(ctx context.Context, model string, onProgress func(PullProgress)) error {
	reqBody, err := json.Marshal(PullRequest{Model: model})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/api/pull", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama pull returned status %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		var progress PullProgress
		if err := decoder.Decode(&progress); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}
		if onProgress != nil {
			onProgress(progress)
		}
	}

	return nil
}

func (c *OllamaClient) DeleteModel(ctx context.Context, model string) error {
	reqBody, err := json.Marshal(PullRequest{Model: model})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", c.BaseURL+"/api/delete", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama delete returned status %d", resp.StatusCode)
	}

	return nil
}

func (c *OllamaClient) Distill(ctx context.Context, model, systemPrompt, content string) (string, error) {
	fullPrompt := fmt.Sprintf("%s\n\nCONTENT TO DISTILL:\n%s", systemPrompt, content)

	reqBody, err := json.Marshal(GenerateRequest{
		Model:  model,
		Prompt: fullPrompt,
		Stream: false,
	})
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var genResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return "", err
	}

	return genResp.Response, nil
}
