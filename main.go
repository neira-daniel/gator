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

	ctx := context.Background()
	if err := s.db.NukeData(ctx); err != nil {
		return fmt.Errorf("(reset user table) couldn't delete records: %w", err)
	}
	fmt.Println("table 'users' was reset successfully")

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
