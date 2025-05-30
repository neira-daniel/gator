package main

import (
	"context"
	"database/sql"

	"fmt"
	"gator/internal/config"
	"gator/internal/database"

	"log"

	"os"

	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
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
		log.Fatal(fmt.Errorf("couldn't read configuration file from disk: %w", err))
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
	// log in as an existing user
	c.register("login", handlerLogin)
	// register a new user
	c.register("register", handlerRegister)
	// delete all database records
	c.register("reset", handlerNukeUserData)
	// list all registered users
	c.register("users", handleListUsers)
	// starts the infinite fetching loop of feeds
	c.register("agg", handlerAgg)
	// add and follow a feed
	c.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	// list all registered feeds
	c.register("feeds", handlerListAllFeeds)
	// follow a registered feed
	c.register("follow", middlewareLoggedIn(handlerFollow))
	// list feeds followed by current user
	c.register("following", middlewareLoggedIn(handlerFollowing))
	// unfollow a feed followed by current user
	c.register("unfollow", middlewareLoggedIn(handlerUnfollowFeeds))
	// print post titles from feeds followed by active user
	c.register("browse", middlewareLoggedIn(handlerBrowse))

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
