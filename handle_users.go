package main

import (
	"context"
	"errors"
	"fmt"
	"gator/internal/database"
	"time"

	"github.com/google/uuid"
)

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

func handleListUsers(s *state, cmd command) error {
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
