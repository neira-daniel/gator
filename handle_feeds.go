package main

import (
	"context"
	"errors"
	"fmt"
	"gator/internal/database"
	"time"

	"github.com/google/uuid"
)

func handlerAgg(s *state, cmd command) error {
	if len(cmd.arguments) != 0 {
		return errors.New("'agg' doesn't take arguments")
	}

	ctx := context.Background()
	url := "https://www.wagslane.dev/index.xml"
	xmlData, err := fetchFeed(ctx, url)
	if err != nil {
		return fmt.Errorf("fetching feed in %v: %w", url, err)
	}

	fmt.Printf("%v\n", xmlData)

	return nil
}

// handlerAddFeed allows the user to add a feed to the database and start following it.
//
// It takes a name for the new feed and its URL.
//
// It returns a non-nil error if there was a problem adding the feed to the database
// (for example, when the feed already exists), if it wasn't possible to make the user
// follow it or the user made a mistake when calling the command.
func handlerAddFeed(s *state, cmd command, userData database.User) error {
	if len(cmd.arguments) != 2 {
		return fmt.Errorf("usage: %v <feed name> <feed URL>", cmd.name)
	}

	ctx := context.Background()
	timestamp := time.Now().UTC()
	feedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
		Name:      cmd.arguments[0],
		Url:       cmd.arguments[1],
		UserID:    userData.ID,
	}

	if _, err := s.db.CreateFeed(ctx, feedParams); err != nil {
		return fmt.Errorf("storing feed data to the database: %w", err)
	}

	if err := handlerFollow(s, command{name: cmd.name, arguments: []string{cmd.arguments[1]}}, userData); err != nil {
		return fmt.Errorf("following feed after adding it: %w", err)
	}

	return nil
}

// handlerListAllFeeds lists all feeds registered in the database. It prints the name
// of the feed, the URL, and the username of the user that added it to the database.
//
// This function doesn't take arguments.
//
// It returns a non-nil error if there was a problem querying the database or the user
// made a mistake when calling the command.
func handlerListAllFeeds(s *state, cmd command) error {
	if len(cmd.arguments) != 0 {
		return fmt.Errorf("%q doesn't take arguments", cmd.name)
	}

	ctx := context.Background()
	feeds, err := s.db.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("getting feed information from the database: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("%q\t%v\t%v\n", feed.FeedName, feed.FeedUrl, feed.UserName)
	}

	return nil
}

// handlerFollow allows the user to follow a registered feed.
//
// It takes the feed's URL as argument.
//
// It returns a non-nil error if there was a problem querying the database to save or
// load records or if the user made a mistake when calling the command.
func handlerFollow(s *state, cmd command, userData database.User) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("usage: %v <url>", cmd.name)
	}

	ctx := context.Background()
	feedID, err := s.db.GetFeedIdByURL(ctx, cmd.arguments[0])
	if err != nil {
		return fmt.Errorf("getting feed record from the database: %w", err)
	}

	timestamp := time.Now().UTC()
	feedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
		UserID:    userData.ID,
		FeedID:    feedID,
	}

	feedFollowData, err := s.db.CreateFeedFollow(ctx, feedFollowParams)
	if err != nil {
		return fmt.Errorf("storing feed follow record in the database: %w", err)
	}

	fmt.Printf("following feed with values\n%v", feedFollowData)

	return nil
}

// handlerFollowing lists all feeds followed by the current user.
//
// It doesn't take arguments.
//
// It returns a non-nil error if there was a problem querying the database or the user
// made a mistake when calling the command.
func handlerFollowing(s *state, cmd command, userData database.User) error {
	if len(cmd.arguments) != 0 {
		return fmt.Errorf("%q doesn't take arguments", cmd.name)
	}

	ctx := context.Background()
	feeds, err := s.db.GetFeedFollowsForUser(ctx, userData.ID)
	if err != nil {
		return fmt.Errorf("getting feed follows: %w", err)
	}

	for _, feedRecord := range feeds {
		fmt.Printf("---\n%v", feedRecord)
	}

	return nil
}

func handlerUnfollowFeeds(s *state, cmd command, userData database.User) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("usage: %v <url>", cmd.name)
	}

	ctx := context.Background()

	url := cmd.arguments[0]
	feedID, err := s.db.GetFeedIdByURL(ctx, url)
	if err != nil {
		return fmt.Errorf("getting feed ID: %w", err)
	}

	if err := s.db.UnfollowFeed(ctx, database.UnfollowFeedParams{UserID: userData.ID, FeedID: feedID}); err != nil {
		return fmt.Errorf("deleting feed-follow from the database: %w", err)
	}

	return nil
}
