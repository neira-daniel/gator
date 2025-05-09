package main

import (
	"context"
	"fmt"
)

// handlerNukeUserData deletes the contents of the `users` table in the database. This
// effectively deletes all data as a consequence of the rules declared in the database's
// schema. The function returns a non-nil error if the operation was unsuccessful.
func handlerNukeUserData(s *state, cmd command) error {
	if len(cmd.arguments) != 0 {
		return fmt.Errorf("%q doesn't take arguments", cmd.name)
	}

	if err := s.cfg.SetUser(""); err != nil {
		return fmt.Errorf("update the configuration file while resetting database: %w", err)
	}

	ctx := context.Background()
	if err := s.db.NukeData(ctx); err != nil {
		return fmt.Errorf("resetting database: %w", err)
	}
	fmt.Println("the database was reset successfully")

	return nil
}
