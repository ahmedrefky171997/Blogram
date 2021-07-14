package records

//                       IMPLEMENTS MODELS

import (
	"database/sql"
	"fmt"
	"path"
	"strconv"

	"simple-photoblog-repo/models"
)

type Post struct {
	Id           int
	UserName     string
	PostAuthor   string // The user who created this post (email)
	TimeStamp    string
	Caption      string
	ImageCaption string
	ImagePath    string
	PostAudio    string
}

type PostRecord struct {
	Db *sql.DB
}

func (post Post) GetId() int {
	return post.Id
}

func (post Post) GetUserName() string {
	return post.UserName
}

func (post Post) GetPostAuthor() string {
	return post.PostAuthor
}

func (post Post) GetTimeStamp() string {
	return post.TimeStamp
}
func (post Post) GetCaption() string {
	return post.Caption
}

func (post Post) GetImageCaption() string {
	return post.ImageCaption
}

func (post Post) GetImagePath() string {
	return post.ImagePath
}

func (post Post) GetPostAudio() string {
	return post.PostAudio
}

// PostRecord Methods: ...

// Inserts a post in the DB
func (postRecord *PostRecord) InsertPost(post models.Post) {
	// post implements post interface from package models
	userName := post.GetUserName()
	postAuthor := post.GetPostAuthor()
	caption := post.GetCaption()
	imageCaption := post.GetImageCaption()
	imagePath := post.GetImagePath()
	timeStamp := post.GetTimeStamp()
	postAudio := post.GetPostAudio()

	// The insertion query statment
	stmt, err := postRecord.Db.Prepare(`INSERT INTO posts (username, author, date, caption, imagecaption,
		 imagename, postaudio) 
		VALUES ("` + userName + `", "` + postAuthor + `", "` + timeStamp + `",
		 "` + caption + `", "` + imageCaption + `", "` + imagePath + `", "` + postAudio + `")`)
	check(err)
	_, err = stmt.Exec() // Execute the statement
	check(err)
}

// Returns all posts created
func (postRecord *PostRecord) GetAllPosts() []Post {

	rows, err := postRecord.Db.Query(`SELECT * FROM posts;`) // Get all rows from posts table
	check(err)

	var post Post
	var posts []Post

	for rows.Next() {
		err = rows.Scan(&post.Id, &post.PostAuthor, &post.UserName, &post.TimeStamp, &post.Caption, &post.ImageCaption,
			&post.ImagePath, &post.PostAudio) // Get the image name
		check(err)
		post.ImagePath = path.Join("/", "posts_images", post.ImagePath)
		post.PostAudio = path.Join("/", "posts_audio", post.PostAudio)
		// fmt.Println(post.UserName)
		posts = append(posts, post) // Append the image name to slice of names
	}

	return posts
}

// Returns a single user posts using their email
func (postRecord *PostRecord) GetAllPostsAuthor(author string) []Post {

	rows, err := postRecord.Db.Query(`SELECT * FROM posts WHERE author = "` + author + `";`) // Get all rows from posts table
	check(err)

	var post Post
	var posts []Post

	for rows.Next() {
		err = rows.Scan(&post.Id, &post.PostAuthor, &post.UserName, &post.TimeStamp, &post.Caption, &post.ImageCaption,
			&post.ImagePath, &post.PostAudio) // Get the image name
		check(err)
		post.ImagePath = path.Join("/", "posts_images", post.ImagePath)
		// fmt.Println(post.UserName)
		posts = append(posts, post) // Append the image name to slice of names
	}

	return posts
}

//                  NEEDS TO BE TESTED !!!!!!!
func (postRecord *PostRecord) SearchPostId(id int) Post {

	var post Post
	row, err := postRecord.Db.Query(`SELECT * FROM posts WHERE posts.id = ` +
		strconv.Itoa(id) + `;`) // Get the required row from posts table by it's id
	check(err)

	err = row.Scan(&post.Id, &post.UserName, &post.Caption, &post.ImageCaption,
		&post.ImagePath, &post.PostAudio) // Get the image name

	return post
}

//                  NEEDS TO BE TESTED !!!!!!!
func (postRecord *PostRecord) DeletePostId(id int) {

	_, err := postRecord.Db.Query(`DELETE FROM posts WHERE posts.id = ` +
		strconv.Itoa(id) + `;`) // Get the required row from posts table by it's id
	check(err)

}

// 								TODO
func (PostRecord *PostRecord) SortPosts() {
	//							TODO
}

// Helper function to print out errors
func check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
