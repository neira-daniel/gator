package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"_ link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func (r *RSSFeed) String() string {
	feedStr := fmt.Sprintf(`{
  Title       : %v,
  Link        : %v,
  Description : %v,
`, r.Channel.Title, r.Channel.Link, r.Channel.Description)
	for _, v := range r.Channel.Item {
		feedStr += fmt.Sprintf("%v", &v)
	}
	return feedStr + "}"
}

func (r *RSSItem) String() string {
	return fmt.Sprintf(`  {
    Title       : %v,
    Link        : %v,
    Description : %v,
    PubDate     : %v
  },
  `, r.Title, r.Link, r.Description, r.PubDate)
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building GET request to fetch feed: %w", err)
	}
	req.Header.Set("User-Agent", "gator")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making GET request to fetch %v: %w", feedURL, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body of GET request to %v: %w", feedURL, err)
	}

	reader := bytes.NewReader(body)
	decoder := xml.NewDecoder(reader)
	decoder.DefaultSpace = "_"

	var rss RSSFeed
	if err := decoder.Decode(&rss); err != nil {
		return nil, fmt.Errorf("decoding response body to GET request to %v: %w", feedURL, err)
	}

	rss.Channel.Title = html.UnescapeString(rss.Channel.Title)
	rss.Channel.Description = html.UnescapeString(rss.Channel.Description)
	for i := range rss.Channel.Item {
		rss.Channel.Item[i].Title = html.UnescapeString(rss.Channel.Item[i].Title)
		rss.Channel.Item[i].Description = html.UnescapeString(rss.Channel.Item[i].Description)
	}

	return &rss, nil
}
