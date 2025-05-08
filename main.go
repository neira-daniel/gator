package main

import (
	"context"
	"database/sql"

	"errors"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"

	"log"

	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

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

	_, err := s.db.CreateFeed(ctx, feedParams)
	if err != nil {
		return fmt.Errorf("storing feed data to the database: %w", err)
	}

	if err := handlerFollow(s, command{name: cmd.name, arguments: []string{cmd.arguments[1]}}, userData); err != nil {
		return fmt.Errorf("following feed after adding it: %w", err)
	}

	return nil
}

func handlerFeeds(s *state, cmd command) error {
	if len(cmd.arguments) != 0 {
		return fmt.Errorf("%q doesn't take arguments", cmd.name)
	}

	ctx := context.Background()
	feeds, err := s.db.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("getting feed info. from db: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("%q\t%v\t%v\n", feed.FeedName, feed.FeedUrl, feed.UserName)
	}

	return nil
}

func handlerFollow(s *state, cmd command, userData database.User) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("usage: %v <url>", cmd.name)
	}

	ctx := context.Background()
	feedID, err := s.db.GetFeedIdByURL(ctx, cmd.arguments[0])
	if err != nil {
		return fmt.Errorf("getting feed data: %w", err)
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
		return fmt.Errorf("creating feed follow record: %w", err)
	}

	fmt.Printf("%v", feedFollowData)

	return nil
}

func handlerFollowing(s *state, cmd command, userData database.User) error {
	if len(cmd.arguments) != 0 {
		return fmt.Errorf("%q doesn't take arguments", cmd.name)
	}

	ctx := context.Background()
	feeds, err := s.db.GetFeedFollowsForUser(ctx, userData.ID)
	if err != nil {
		return fmt.Errorf("getting feed follows: %w", err)
	}

	fmt.Printf("%+v\n", feeds)

	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		ctx := context.Background()
		user, err := s.db.GetUser(ctx, s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("getting user data: %w", err)
		}

		return handler(s, cmd, user)
	}
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(fmt.Errorf("couldn't read original configuration file from disk: %w", err))
	}

	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		log.Fatal(fmt.Errorf("couldn't prepare the database abstraction: %w", err))
	}
	defer db.Close()

	dbQueries := database.New(db)

	s := &state{
		db:  dbQueries,
		cfg: &cfg,
	}

	c := commands{
		handler: make(map[string]func(*state, command) error),
	}
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", handlerNukeUserData)
	c.register("users", handleListUsers)
	c.register("agg", handlerAgg)
	c.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	c.register("feeds", handlerFeeds)
	c.register("follow", middlewareLoggedIn(handlerFollow))
	c.register("following", middlewareLoggedIn(handlerFollowing))

	cliArgs := os.Args
	if len(cliArgs) < 2 {
		log.Fatal("no command given")
	}

	cmd := command{
		name:      cliArgs[1],
		arguments: cliArgs[2:],
	}
	if err := c.run(s, cmd); err != nil {
		log.Fatal(fmt.Errorf("failed to run '%v' command: %w", cmd.name, err))
	}

}
