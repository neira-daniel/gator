package main

import (
	"fmt"
)

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
