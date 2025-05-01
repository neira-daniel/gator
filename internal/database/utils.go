package database

import "fmt"

func (u User) String() string {
	return fmt.Sprintf("* ID        : %v\n* CreatedAt : %v\n* UpdatedAt : %v\n* Name      : %v", u.ID, u.CreatedAt, u.UpdatedAt, u.Name)
}
