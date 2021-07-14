package models

type Follow interface {
	GetUser() string
	GetFollows() string
	GetDate() string
}

type FollowRecord interface {
	InsertFollow(follow Follow)
	IsFollowed(user string, follows string) bool
	Unfollow(follow Follow)
	GetAllFollowers(follow Follow)
}
