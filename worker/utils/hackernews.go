package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
)

type Hits struct {
	ObjectID string   `json:"objectID"`
	Title    string   `json:"title"`
	Url      *string  `json:"url"`
	Tags     []string `json:"_tags"`
	Children []int    `json:"children"`
}
type HackernewsStoryData struct {
	Hits []Hits `json:"hits"`
}

type HackernewsItemChildren struct {
	Text string `json:"text"`
}

type HackerNewsCommentData struct {
	Title    string                   `json:"title"`
	Children []HackernewsItemChildren `json:"children"`
}

func isAskOrShow(tags []string) bool {
	return slices.Contains(tags, "ask_hn") || slices.Contains(tags, "show_hn")
}

func StoryComment(ctx context.Context, objectId string, httpClient *http.Client) (*HackerNewsCommentData, error) {
	itemsUrl := url.URL{
		Scheme: "http",
		Host:   "hn.algolia.com",
		Path:   fmt.Sprintf("/api/v1/items/%s", objectId),
	}

	req, err := http.NewRequestWithContext(ctx, "GET", itemsUrl.String(), nil)

	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Err fetching data from Hackernews! %v", resp.StatusCode)
		return nil, err
	}

	var itemRes HackerNewsCommentData

	err = json.NewDecoder(resp.Body).Decode(&itemRes)

	if err != nil {
		return nil, err
	}

	for i, _ := range itemRes.Children {
		sanitizedText, err := Sanitize(itemRes.Children[i].Text)
		if err != nil {
			return nil, err
		}
		itemRes.Children[i].Text = sanitizedText
	}
	return &itemRes, nil
}

func FetchHackerNews(ctx context.Context, query string, httpClient *http.Client) ([]*Resource, error) {
	var resources []*Resource
	urlQuery := url.Values{}
	urlQuery.Set("query", query)
	urlQuery.Add("tags", "story")
	urlQuery.Add("hitsPerPage", "5")
	urlQuery.Add("numericFilters", "points>50,num_comments>10")
	hackernewsUrl := url.URL{
		Scheme:   "http",
		Host:     "hn.algolia.com",
		Path:     "/api/v1/search",
		RawQuery: urlQuery.Encode(),
	}
	req, err := http.NewRequestWithContext(ctx, "GET", hackernewsUrl.String(), nil)

	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Err fetching data from Hackernews! %v", resp.StatusCode)
		return nil, err
	}
	var stories HackernewsStoryData
	err = json.NewDecoder(resp.Body).Decode(&stories)
	if err != nil {
		return nil, err
	}

	for _, hit := range stories.Hits {
		resource := &Resource{}
		if isAskOrShow(hit.Tags) && len(hit.Children) > 0 {
			comments, err := StoryComment(ctx, hit.ObjectID, httpClient)
			if err != nil {
				return nil, err
			}
			resource.Title = comments.Title
			var contents []string
			for _, comment := range comments.Children {
				contents = append(contents, comment.Text)
			}
			resource.Content = contents
			resources = append(resources, resource)
		} else {
			resource.Title = hit.Title
			if hit.Url != nil {
				content := []string{*hit.Url}
			    resource.Content = content
			}
			resources = append(resources, resource)
		}
	}
	return resources, nil
}
