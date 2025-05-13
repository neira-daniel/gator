package main

import (
	"context"
	"fmt"
	"gator/internal/database"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// handlerAgg starts a loop that fetches feeds indefinitely. It blocks the program
// execution, and can only be terminated pressing CTRL-C. The users need to use the
// program in a different instance (shell) to the one that is running `handlerAgg`.
//
// The function takes a human readable time interval expressed in "ms", "s", "m" or
// "h" for miliseconds, seconds, minutes and hours, respectively. They can be mixed
// with each other. For example, "1h10m20s" is a valid interval of 1 hour, 10 minutes
// and 20 seconds.
//
// handlerAgg returns a non-nil error only when the received time interval couldn't
// be parsed. In case there was a problem fetching a feed, the error will be logged
// without killing the program so it keeps working on fetching the next feed in the
// queue.
func handlerAgg(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("usage: %v <time between requests>", cmd.name)
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.arguments[0])
	if err != nil {
		return fmt.Errorf("parsing the time between requests parameter: %w", err)
	}
	fmt.Printf("Collecting feeds every %v starting right now\n", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)
	// block execution by not using a goroutine and start fetching feeds immediately
	// by using an empty for-condition (we should have used `for range ticker.C` or
	// `for t := range ticker.C` if we want the ticker `t` timestamp to wait for the
	// first tick to start fetching feeds)
	t := time.Now()
	for {
		if err := scrapeFeeds(s); err != nil {
			log.Printf("%v - found error while scraping feeds: %v", t.UTC(), err)
		}
		// we reassign the value of the `t` we declared and assigned before the for-block
		t = <-ticker.C
	}

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
		return fmt.Errorf("getting feed follows from the database: %w", err)
	}

	for _, feedRecord := range feeds {
		fmt.Printf("---\n%v", feedRecord)
	}

	return nil
}

// handlerUnfollowFeeds allows the user to unfollow a feed.
//
// It takes the url of the feed to unfollow as parameter.
//
// It returns a non-nil error when it wasn't possible to query the database for the
// feed id or there was an error updating the feed_follows database.
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

func handlerBrowse(s *state, cmd command, userData database.User) error {
	if len(cmd.arguments) > 1 {
		return fmt.Errorf("usage: %v [number of posts]", cmd.name)
	}

	var limit int32
	limit = 2
	if len(cmd.arguments) == 1 {
		limit64, err := strconv.ParseInt(cmd.arguments[0], 10, 32)
		if err != nil {
			return fmt.Errorf("parsing optional number of posts parameter: %w", err)
		}
		limit = int32(limit64)
	}

	ctx := context.Background()
	feedFollows, err := s.db.GetFeedFollowsForUser(ctx, userData.ID)
	if err != nil {
		return fmt.Errorf("getting feed follows from the database: %w", err)
	}

	for _, feedFollow := range feedFollows {
		posts, err := s.db.GetPostsForUser(ctx, database.GetPostsForUserParams{
			FeedID: feedFollow.FeedID, Limit: limit,
		})
		if err != nil {
			log.Printf("[NOT OK] getting posts from %q for %v: %v", feedFollow.FeedName, userData.Name, err)
		}

		fmt.Printf("---\n%v\n", feedFollow.FeedName)
		for _, post := range posts {
			fmt.Printf("- %v\n", post.Title)
		}

	}

	return nil
}
