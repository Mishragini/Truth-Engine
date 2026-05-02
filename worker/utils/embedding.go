package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pgvector/pgvector-go"
)

type JinaRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type JinaEmbedding struct {
	Embedding pgvector.Vector `json:"embedding"`
}
type JinaResponse struct {
	Data []JinaEmbedding `json:"data"`
}

func GetEmbedding(ctx context.Context, text string, httpClient *http.Client, jinaApiKey string) (*pgvector.Vector, error) {
	body, _ := json.Marshal(JinaRequest{
		Input: []string{text},
		Model: "jina-embeddings-v2-base-en",
	})

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.jina.ai/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+jinaApiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jina api error: status %d", resp.StatusCode)
	}

	var jinaRes JinaResponse

	if err := json.NewDecoder(resp.Body).Decode(&jinaRes); err != nil {
		return nil, err
	}

	if len(jinaRes.Data) == 0 {
		return nil, fmt.Errorf("jina returned no embeddings")
	}

	return &jinaRes.Data[0].Embedding, nil
}
