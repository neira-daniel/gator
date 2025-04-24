package main

import (
	"errors"
	"fmt"
	"gator/internal/config"
	"log"
	"os"
)

type state struct {
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
	err := c.handler[cmd.name](s, cmd)
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

	if err := s.cfg.SetUser(userName); err != nil {
		return err
	}

	fmt.Printf("username has been set to %v\n", userName)

	return nil

}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(fmt.Errorf("couldn't read original configuration file from disk: %w", err))
	}

	s := &state{
		cfg: &cfg,
	}

	c := commands{
		handler: make(map[string]func(*state, command) error),
	}
	c.register("login", handlerLogin)

	cliArgs := os.Args
	if len(cliArgs) < 3 {
		if len(cliArgs) < 2 {
			log.Fatal("no command given")
		} else {
			log.Fatal("no argument given")

		}
	}

	cmd := command{
		name:      cliArgs[1],
		arguments: cliArgs[2:],
	}
	if err := c.run(s, cmd); err != nil {
		log.Fatal(fmt.Errorf("failed to run command '%v': %w", cmd.name, err))
	}

}
