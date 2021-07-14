package models

type User interface {
	GetId() string
	GetEmail() string
	GetPassword() string
	GetFirstName() string
	GetLastName() string
	GetVisuallyImpaired() bool
	GetUserImage() string
	GetPrivate() bool
}

type UserRecord interface {
	InsertUser(user User)
	GetUser(email string) (User, error)
	UpdateUser(user User)
	GetUsersFirstLast(userName string) ([]User, error)
	GetUserId(userId string) (User, error)
}
