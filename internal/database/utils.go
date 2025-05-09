package database

import "fmt"

func (u User) String() string {
	return fmt.Sprintf(`* ID        : %v
* CreatedAt : %v
* UpdatedAt : %v
* Name      : %v
`, u.ID, u.CreatedAt, u.UpdatedAt, u.Name)
}

func (f Feed) String() string {
	return fmt.Sprintf(`* ID        : %v
* CreatedAt : %v
* UpdatedAt : %v
* Name      : %v
* Url       : %v
* UserID    : %v
`, f.ID, f.CreatedAt, f.UpdatedAt, f.Name, f.Url, f.UserID)
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

func (f GetFeedFollowsForUserRow) String() string {
	return fmt.Sprintf(`* FeedID   : %v
* FeedName : %v
* Username : %v
`, f.FeedID, f.FeedName, f.UserName)
}
