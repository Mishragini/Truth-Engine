package utils

import "github.com/pgvector/pgvector-go"

type Query struct {
	QueryId     string `json:"query_id"`
	SearchQuery string `json:"search_query"`
}

type Resource struct {
	Title   string
	Content []string
}

type SaveResponseDb struct {
	Query     string
	Embedding pgvector.Vector
	Response  string
	Resources []*Resource
}

type WSMessage struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}
