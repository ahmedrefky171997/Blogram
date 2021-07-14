package records

//                       IMPLEMENTS MODELS

import (
	"database/sql"

	// "fmt"
	// "path"
	// "strconv"

	"simple-photoblog-repo/models"
)

type Follow struct {
	User_    string
	Follows_ string
	Date     string
}

type FollowRecord struct {
	Db *sql.DB
}

func (follow Follow) GetUser() string {
	return follow.User_
}

func (follow Follow) GetFollows() string {
	return follow.Follows_
}

func (follow Follow) GetDate() string {
	return follow.Date
}

// FollowRecord Methods: ...

// Inserts a follow in the DB
func (followRecord *FollowRecord) InsertFollow(follow models.Follow) {
	// post implements post interface from package models
	userId := follow.GetUser()
	follows := follow.GetFollows()

	// The insertion query statment
	stmt, err := followRecord.Db.Prepare(`INSERT INTO followers (user_, follows_) 
		VALUES ("` + userId + `", "` + follows + `")`)
	check(err)
	_, err = stmt.Exec() // Execute the statement
	check(err)
}

func (followRecord *FollowRecord) IsFollowed(user string, follows string) bool {

	// The insertion query statment
	rows, err := followRecord.Db.Query(`SELECT * FROM followers
		WHERE user_ = "` + user + `" AND follows_ = "` + follows + `";`)
	check(err)

	// If such relation exists then return true
	if rows.Next() == true {
		return true
	}

	return false
}

func (followRecord *FollowRecord) Unfollow(follow models.Follow) {
	// The deletion query statment
	user := follow.GetUser()
	follows := follow.GetFollows()
	_, err := followRecord.Db.Exec(`DELETE from followers 
	WHERE user_ = ? AND follows_ = ?;`, user, follows)
	check(err)
}

func (followRecord *FollowRecord) GetAllFollowers(user string) []string {
	// Get all the followers of this users
	// The insertion query statment
	var followers []string
	var follower string
	rows, err := followRecord.Db.Query(`SELECT follows_ FROM followers
		WHERE user_ = "` + user + `";`)
	check(err)

	for rows.Next() {
		err = rows.Scan(&follower)
		check(err)
		followers = append(followers, follower)
	}

	return followers
}
