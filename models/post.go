package models

type Post interface {
	GetId() int
	GetUserName() string
	GetPostAuthor() string
	GetCaption() string
	GetImageCaption() string
	GetImagePath() string
	GetTimeStamp() string
	GetPostAudio() string
}

type PostRecord interface {
	InsertPost(post Post)
	GetAllPosts() []Post
	SearchPostId(id int) Post
	DeletePostId(id int)
	GetAllPostsAuthor(author string) []Post
}
