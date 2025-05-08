package database

import "fmt"

func (u User) String() string {
	return fmt.Sprintf("* ID        : %v\n* CreatedAt : %v\n* UpdatedAt : %v\n* Name      : %v", u.ID, u.CreatedAt, u.UpdatedAt, u.Name)
}

func (f Feed) String() string {
	return fmt.Sprintf("* ID        : %v\n* CreatedAt : %v\n* UpdatedAt : %v\n* Name      : %v\n* Url       : %v\n* UserID    : %v\n", f.ID, f.CreatedAt, f.UpdatedAt, f.Name, f.Url, f.UserID)
}

func (f CreateFeedFollowRow) String() string {
	return fmt.Sprintf(`* ID        : %v
* CreatedAt : %v
* UpdatedAt : %v
* UserID    : %v
* FeedID    : %v
* UserName  : %v
* FeedName  : %v
`, f.ID, f.CreatedAt, f.UpdatedAt, f.UserID, f.FeedID, f.UserName, f.FeedName)
}
