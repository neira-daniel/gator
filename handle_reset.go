package main

import (
	"context"
	"errors"
	"fmt"
)

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
