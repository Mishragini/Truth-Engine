package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

const SystemPrompt = `
You are an intelligent search assistant. Your job is to answer the user's question using the provided sources.

The sources may be noisy, partially relevant, or incomplete (e.g., links, comments, discussions).

Follow this reasoning process:

1. Identify which parts of the sources are relevant to the user's question.
2. Ignore irrelevant or unrelated content.
3. Extract useful facts, opinions, or signals from the relevant parts.
4. Combine them into a clear and helpful answer.

Rules:
- Use ONLY the provided sources. Do not introduce external knowledge.
- If the sources are partially relevant, provide the best possible answer using what is available.
- Do NOT immediately give up just because the answer is not explicitly stated.
- If multiple sources suggest trends or opinions, summarize them.
- If sources contradict each other, mention it briefly.
- If the sources are completely irrelevant, say:
  "I don't have enough relevant information to answer this."

Style:
- Be concise and direct.
- Do not mention "sources" or "context" explicitly.
- Do not hallucinate facts, URLs, or data.

Sources:
%s
`

func FormatText(resources []*Resource) string {
	var b strings.Builder
	for _, resource := range resources {
		fmt.Fprintf(&b, "Title: %s \n Source:%v", resource.Title, resource.Content)
	}
	return b.String()
}

func PromptLlm(ctx context.Context, aiClient *openai.Client, resources []*Resource, query string) (*openai.ChatCompletionStream, error) {
	formattedResources := FormatText(resources)
	sysPrompt := fmt.Sprintf(SystemPrompt, formattedResources)
	stream, err := aiClient.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model: "llama-3.3-70b-versatile",
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: sysPrompt},
			{Role: openai.ChatMessageRoleUser, Content: query},
		},
		Stream: true,
	})
	if err != nil {
		return nil, err
	}
	return stream, nil
}

func ExtractKeywords(ctx context.Context, aiClient *openai.Client, query string) (string, error) {
	resp, err := aiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "llama-3.3-70b-versatile",
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleSystem,
				Content: `Extract 2-3 broad technical keywords from the user query for searching HackerNews.
							Prefer general terms over specific ones. For example:
							- "how to learn golang efficiently" → "golang learning"
							- "best rust async runtime" → "rust async"  
							- "kubernetes pod networking issues" → "kubernetes networking"
							Return ONLY the keywords separated by spaces, nothing else.`,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: query,
			},
		},
		MaxTokens: 20,
	})

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from AI")
	}

	return resp.Choices[0].Message.Content, nil
}
