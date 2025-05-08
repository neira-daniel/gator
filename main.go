package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	handler map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handler[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	f, ok := c.handler[cmd.name]
	if !ok {
		return fmt.Errorf("invalid command: '%v' not found", cmd.name)
	}
	err := f(s, cmd)
	return err
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return errors.New("didn't pass the user name to the 'login' command")
	}
	if len(cmd.arguments) > 1 {
		return errors.New("too many arguments for the 'login' command")
	}

	userName := cmd.arguments[0]

	ctx := context.Background()
	if _, err := s.db.GetUser(ctx, userName); err != nil {
		return fmt.Errorf("user '%v' is not registered", userName)
	}

	if err := s.cfg.SetUser(userName); err != nil {
		return err
	}

	fmt.Printf("username has been set to %v\n", userName)

	return nil

}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return errors.New("didn't pass the user name to the 'register' command")
	}
	if len(cmd.arguments) > 1 {
		return errors.New("too many arguments for the 'register' command")
	}

	timestamp := time.Now().UTC()
	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
		Name:      cmd.arguments[0],
	}

	ctx := context.Background()

	user, err := s.db.CreateUser(ctx, userParams)
	if err != nil {
		return fmt.Errorf("couldn't create user: %w", err)
	}

	if err = s.cfg.SetUser(userParams.Name); err != nil {
		return fmt.Errorf("couldn't save the configuration: %w", err)
	}

	fmt.Printf("user %v created with values\n%v\n", userParams.Name, user)

	return nil
}

func handlerNukeUserData(s *state, cmd command) error {
	if len(cmd.arguments) != 0 {
		return errors.New("'reset' doesn't take arguments")
	}

	if err := s.cfg.SetUser(""); err != nil {
		return fmt.Errorf("reset table 'users' :: couldn't update config. file: %w", err)
	}

	ctx := context.Background()
	if err := s.db.NukeData(ctx); err != nil {
		return fmt.Errorf("(reset user table) couldn't delete records: %w", err)
	}
	fmt.Println("table 'users' was reset successfully")

	return nil
}

func handleUsers(s *state, cmd command) error {
	if len(cmd.arguments) != 0 {
		return errors.New("'users' doesn't take arguments")
	}

	ctx := context.Background()
	users, err := s.db.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("list users :: couldn't list registered users: %w", err)
	}

	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %v (current)\n", user.Name)
		} else {
			fmt.Printf("* %v\n", user.Name)
		}

	}

	return nil
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
	feedStr := fmt.Sprintf("{\n  Title       : %v,\n  Link        : %v,\n  Description : %v,\n", r.Channel.Title, r.Channel.Link, r.Channel.Description)
	for _, v := range r.Channel.Item {
		feedStr += fmt.Sprintf("%v", &v)
	}
	return feedStr + "}"
}

func (r *RSSItem) String() string {
	return fmt.Sprintf("  {\n    Title       : %v,\n    Link        : %v,\n    Description : %v,\n    PubDate     : %v\n  },\n", r.Title, r.Link, r.Description, r.PubDate)
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

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.arguments) != 2 {
		return fmt.Errorf("usage: %v <feed name> <feed URL>", cmd.name)
	}

	ctx := context.Background()

	userData, err := s.db.GetUser(ctx, s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("getting %v's data from the database: %w", s.cfg.CurrentUserName, err)
	}

	timestamp := time.Now().UTC()
	feedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
		Name:      cmd.arguments[0],
		Url:       cmd.arguments[1],
		UserID:    userData.ID,
	}

	_, err = s.db.CreateFeed(ctx, feedParams)
	if err != nil {
		return fmt.Errorf("storing feed data to the database: %w", err)
	}

	if err := handlerFollow(s, command{name: cmd.name, arguments: []string{cmd.arguments[1]}}); err != nil {
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

func handlerFollow(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("usage: %v <url>", cmd.name)
	}

	ctx := context.Background()

	userData, err := s.db.GetUser(ctx, s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("getting user data: %w", err)
	}

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

func handlerFollowing(s *state, cmd command) error {
	if len(cmd.arguments) != 0 {
		return fmt.Errorf("%q doesn't take arguments", cmd.name)
	}

	ctx := context.Background()

	userData, err := s.db.GetUser(ctx, s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("getting user data: %w", err)
	}

	feeds, err := s.db.GetFeedFollowsForUser(ctx, userData.ID)
	if err != nil {
		return fmt.Errorf("getting feed follows: %w", err)
	}

	fmt.Printf("%+v\n", feeds)

	return nil
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
	c.register("users", handleUsers)
	c.register("agg", handlerAgg)
	c.register("addfeed", handlerAddFeed)
	c.register("feeds", handlerFeeds)
	c.register("follow", handlerFollow)
	c.register("following", handlerFollowing)

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
