package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"gator/internal/database"
	"html"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
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

func scrapeFeeds(s *state) error {
	ctx := context.Background()

	feed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		return fmt.Errorf("getting feed to fetch: %w", err)
	}

	// we mark the feed as fetched before fetching it to account for the chance that
	// we encounter an error while making the GET request
	// this way, we always store the time at which we attempted the fetch
	if err := s.db.MarkFeedFetched(ctx, database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{Time: time.Now().UTC(), Valid: true}, ID: feed.ID}); err != nil {
		return fmt.Errorf("marking feed from %v as fetched: %w", feed.Url, err)
	}

	xmlData, err := fetchFeed(ctx, feed.Url)
	if err != nil {
		return fmt.Errorf("fetching feed from %v: %w", feed.Url, err)
	}

	log.Printf("[OK] %v\n", xmlData.Channel.Title)
	timestampFormat := "Mon, 02 Jan 2006 15:04:05 -0700"
	timestamp := time.Now().UTC()
	for _, post := range xmlData.Channel.Item {
		pubDate, err := time.Parse(timestampFormat, post.PubDate)
		if err != nil {
			pubDate = time.Time{}
		}
		if err := s.db.CreatePost(ctx, database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   timestamp,
			UpdatedAt:   timestamp,
			Title:       post.Title,
			Url:         post.Link,
			Description: post.Description,
			PublishedAt: pubDate,
			FeedID:      feed.ID,
		}); err != nil {
			log.Printf("%v\n", err)
		}
	}

	return nil
}
