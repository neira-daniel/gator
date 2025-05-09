package main

import (
	"context"
	"fmt"
	"gator/internal/database"
	"time"

	"github.com/google/uuid"
)

// handlerLogin allows a user to log in. It returns a non-nil error if the user
// isn't registered, the configuration file couldn't be updated or the user made
// a mistake when calling the command.
func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("usage: %v <username>", cmd.name)
	}

	userName := cmd.arguments[0]

	ctx := context.Background()
	if _, err := s.db.GetUser(ctx, userName); err != nil {
		return fmt.Errorf("user %q is not registered", userName)
	}

	if err := s.cfg.SetUser(userName); err != nil {
		return err
	}

	fmt.Printf("username has been set to %q\n", userName)

	return nil

}

// handlerRegister allows a user to register themself in the database. It returns a
// non-nil error when the data can't be stored in the database, it wasn't possible
// to set the new user as the active one in the configuration or the user made a
// mistake when calling the command.
func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("usage: %v <username>", cmd.name)
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
		return fmt.Errorf("storing user %q in the database: %w", userParams.Name, err)
	}

	if err = s.cfg.SetUser(userParams.Name); err != nil {
		return fmt.Errorf("setting %q as the active user in the configuration file: %w", userParams.Name, err)
	}

	fmt.Printf("user created with values\n%v\n", user)

	return nil
}

// handleListUsers prints a list of the registered users to the terminal while tagging
// the active one. It returns a non-nil error if it was impossible to query the database
// or the user made a mistake when calling the command.
func handleListUsers(s *state, cmd command) error {
	if len(cmd.arguments) != 0 {
		return fmt.Errorf("%q doesn't take arguments", cmd.name)
	}

	ctx := context.Background()
	users, err := s.db.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("getting all users from the database: %w", err)
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
