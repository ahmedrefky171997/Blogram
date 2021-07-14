package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	_ "github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
	"simple-photoblog-repo/config"
	"simple-photoblog-repo/records"
)

var db *sql.DB
var err error

// var user User
var sessionMap = make(map[string]User) // Session to keep track of active users
var test_chan = make(chan bool)
var test_room  = make([]*Room,0)
var ServerAllRooms = make([]*Room,1000)
var RunningRooms = make([]*Room,1000)
var UnRecordedRooms = make([]*Room,1000)

func main() {
	db = config.InitDB() // Get the db pointer
	err = db.Ping()
	check(err)
	defer db.Close()
	//get list of rooms in room list
	//search in rooms list with 2 hashed ids of the users currently talking to each other
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./dist"))))
	http.HandleFunc("/add", addPost)     // Handle add route
	http.HandleFunc("/home", viewPosts)  // Handle home route
	http.HandleFunc("/signup", signUp)   // Handle signup route
	http.HandleFunc("/login", logIn)     // Handle login route
	http.HandleFunc("/logout", logOut)   // Handle logout route
	http.HandleFunc("/profile", profile) // Handle logout route
	http.HandleFunc("/users", users)     // Handle other users route
	http.HandleFunc("/search", search)   // Handle the search results route
	http.HandleFunc("/chat", chat)       // Handle the chat route
	http.HandleFunc("/c/rooms",chat_room)//handle chat conversation room

	// File Servers to handle the image requests
	http.Handle("/posts_images/", http.StripPrefix("/posts_images", http.FileServer(http.Dir("./posts_images"))))
	http.Handle("/users_images/", http.StripPrefix("/users_images", http.FileServer(http.Dir("./users_images"))))
	http.Handle("/posts_audio/", http.StripPrefix("/posts_audio", http.FileServer(http.Dir("./posts_audio"))))

	http.Handle("/favicon.ico", http.NotFoundHandler()) // Handle the favicon route

	err = http.ListenAndServe(":8080", nil)
	check(err)
}

func signUp(w http.ResponseWriter, r *http.Request) {
	// Check if there is a session already exists
	if alreadyLoggedIn(r) {
		// r.Method = http.MethodGet
		// http.Redirect(w, r, "/#/home", http.StatusSeeOther)
		return
	}

	// enableCors(&w) // Test for other clients to access

	if r.Method == http.MethodPost {
		var userJSON User                               // JSON object of user to get credentials from request
		err = json.NewDecoder(r.Body).Decode(&userJSON) // Decode JSON from the request body
		check(err)

		// fmt.Println("userJson: ", userJSON)

		// Create view user
		user := newUser(userJSON.Email, userJSON.Password, userJSON.FirstName,
			userJSON.LastName, userJSON.VisualImpaired, false)

		// fmt.Println(user.Email)
		// fmt.Println(user.VisualImpaired)

		var userRecord records.UserRecord
		userRecord.Db = db

		// VALIDATE USER EXISTANCE BEFORE CREATING
		var userModel records.User
		userModel, err = userRecord.GetUser(user.Email)
		// If the user already exists send bad request response and return
		if err != nil || userModel.Email != "" {
			http.Error(w, "Email Already Exists!!", http.StatusBadRequest)
			return
		}

		// Hashing Password
		password := []byte(user.Password)
		hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}
		// fmt.Println(string(hashedPassword))

		// Comparing the password with the hash
		err = bcrypt.CompareHashAndPassword(hashedPassword, password)
		// fmt.Println(err) // nil means it is a matc

		userModel.Email = user.Email
		userModel.Id = fmt.Sprintf("%x", sha1.New().Sum([]byte(user.Email))) // Hash the email to get the ID
		userModel.Password = string(hashedPassword)
		userModel.FirstName = user.FirstName
		userModel.LastName = user.LastName
		userModel.VisuallyImpaired = user.VisualImpaired
		userModel.UserImage = "default.jpg"
		userModel.Private = user.Private

		userRecord.InsertUser(userModel)
	}
}

func logIn(w http.ResponseWriter, r *http.Request) {
	// Check if there is a session already exists
	if alreadyLoggedIn(r) {
		return
	}

	if r.Method == http.MethodPost {
		var userJSON User
		err = json.NewDecoder(r.Body).Decode(&userJSON) // Decode JSON from the request body
		check(err)

		var userRecord records.UserRecord
		userRecord.Db = db

		// VALIDATE USER EXISTANCE BEFORE CREATING
		var userModel records.User
		userModel, err = userRecord.GetUser(userJSON.Email)
		check(err)

		// If the email doesn't exist send bad request response and return
		if err != nil || userModel.Email == "" {
			http.Error(w, "Wrong Email and/or Password", http.StatusBadRequest)
			return
		}

		// Compare the given password after hashing with the stored password
		err = bcrypt.CompareHashAndPassword([]byte(userModel.Password), []byte(userJSON.Password))
		correctPassword := true // Boolean to check validity of password

		// If not correct send error response and return
		if err != nil {
			http.Error(w, "Wrong Email and/or Password", http.StatusBadRequest)
			correctPassword = false
			return
		}

		// Compare given email and stored email
		if userJSON.Email == userModel.Email && correctPassword {
			// fmt.Println("CORRECT INFO")
			cookie := createCookie(r)
			// fmt.Println(cookie.Value)
			// Create view user from package main and add it to the sessions map
			user := newUser(userModel.Email, userModel.Password, userModel.FirstName,
				userModel.LastName, userModel.VisuallyImpaired, userModel.Private)

			// TODO populate user.Followers, user.Following, and user.ProfilePosts
			// before admiting the user to the session map

			sessionMap[cookie.Value] = user
			// Create a user cookie
			http.SetCookie(w, &cookie)

			// For local storage on client side send the user info
			userJSON.Password = "" // Clear the password before encoding
			json.NewEncoder(w).Encode(userJSON)

		} else {
			// Send not bad request response
			http.Error(w, "Wrong Email and/or Password", http.StatusBadRequest)
		}
	}
}

func logOut(w http.ResponseWriter, r *http.Request) {
	// DELETE COOKIE
	// Check if the user is logged in
	if !alreadyLoggedIn(r) {
		// r.Method = http.MethodGet
		// http.Redirect(w, r, "/#/login", http.StatusSeeOther)
		return
	}

	// Get the session key to delete it
	c, err := r.Cookie("session")
	check(err)
	sessionKey := c.Value
	// Delete the user entry from the map
	delete(sessionMap, sessionKey)

	// Delete the user cookie
	c = &http.Cookie{
		Name:   "session",
		MaxAge: -1,
	}
	http.SetCookie(w, c)

}

func addPost(w http.ResponseWriter, r *http.Request) {
	// Check if the user is logged in
	if !alreadyLoggedIn(r) {
		// r.Method = http.MethodGet
		// http.Redirect(w, r, "/#/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {

		username := r.FormValue("user")          // Get the username of the post creator
		postCaption := r.FormValue("caption")    // Get the caption of the post
		filePath, fileName := createPostImage(r) // Get the file path of the post's image and the name

		post := newPost(username, postCaption, filePath)
		//go post.imageCaption()         // Run the image captioning in a go routine
		imageCap := <-post.PostChannel // Get the image caption result from the channel
		fmt.Println(imageCap)

		// Used to implement PostRecord interface from models to use it's methods
		var postRecord records.PostRecord
		var userRecord records.UserRecord

		postRecord.Db = db // DB pointer
		userRecord.Db = db // DB pointer

		// postModel and userModel from records implements post and user from models
		// for a post object to be inserted and user object to get the author info
		var postModel records.Post
		var userModel records.User

		postModel.PostAuthor = sessionMap[getUuid(r)].Email       // Get the user from his cookie using sessionMap
		userModel, err = userRecord.GetUser(postModel.PostAuthor) // Used to get the username

		// Preparing for database entry
		username = userModel.FirstName + " " + userModel.LastName
		postModel.UserName = username
		postModel.Caption = postCaption
		postModel.ImagePath = fileName
		postModel.ImageCaption = imageCap
		t := time.Now() // Current time for timestamp of the post
		ts := t.Format("2006-01-02 15:04:05")
		postModel.TimeStamp = ts
		audioName := strings.Split(fileName, ".")[0] + ".wav" // edit the extention for tts
		ts = strings.Replace(ts, ":", "-", -1)                // Replace all the colons by dashes to write the file properly
		postModel.PostAudio = ts + audioName                  // Audio file name is timestamp with no colons + image name

		// Preparing TTS

		// post.TimeStamp = ts
		post.AudioPath = ts + audioName
		// The Text passed to TTS module
		post.TextToSpeech = username + ", at " + ts + " posted: " + postCaption + ", image description: " + imageCap
		//go post.postTTS() // Run TTS module

		// Performing record insertion in the database
		postRecord.InsertPost(postModel)

	}
}

func viewPosts(w http.ResponseWriter, r *http.Request) {
	// Check if the user is not logged in
	if !alreadyLoggedIn(r) {
		// r.Method = http.MethodGet
		// http.Redirect(w, r, "/#/login", http.StatusSeeOther)
		return
	}

	var userRecord records.UserRecord
	userRecord.Db = db // DB pointer

	var userModel records.User
	userModel, err = userRecord.GetUser(sessionMap[getUuid(r)].Email) // Used to get the currently logged in user

	userModel.Password = "" // Clear the password

	w.Header().Set("Content-Type", "application/json") // Set the header type to json

	// Used to implement PostRecord interface from models to use it's methods
	var postRecord records.PostRecord
	postRecord.Db = db

	posts := postRecord.GetAllPosts()

	var followRecord records.FollowRecord
	followRecord.Db = db

	// Create user of the currently logged in user
	user := newUser(userModel.Email, "", userModel.FirstName, userModel.LastName,
		userModel.VisuallyImpaired, userModel.Private)

	// Get this user's followers
	user.Followers = followRecord.GetAllFollowers(userModel.Email)
	//TODO load rooms from database
	user.UserImage = userModel.UserImage
	//create conversation chat rooms for all the given followers
	user.CreateAllRoomsForAGivenUser()
	fmt.Println("main.go--->home page done creating all rooms for a user successfully...",user.rooms)
	//update current active user
	sessionMap[getUuid(r)]= user

	var followedUsers []records.User // The followed users info
	followedUsers = userRecord.GetFollowersInfo(user.Followers)

	// Struct to send followed users to frontend
	type followers struct {
		UserName  string
		UserImage string
		UserEmail string
		UserId    string
	}

	var tmp followers
	var tmps []followers

	// Iterate over followed users to create an encodable JSON for frontend
	for _, follower := range followedUsers {
		tmp.UserName = follower.FirstName + " " + follower.LastName
		tmp.UserImage = follower.UserImage
		tmp.UserEmail = follower.Email
		tmp.UserId = follower.Id

		tmps = append(tmps, tmp)
	}

	// temp struct to send the posts and the currently active user info
	type temp struct {
		Posts     []records.Post
		User      records.User
		Followers []followers
	}

	// Create the object to be sent
	var obj temp
	obj.Posts = posts
	obj.User = userModel
	obj.Followers = tmps

	// Encode posts list and currently active user info and send it to the client
	err = json.NewEncoder(w).Encode(obj)
	check(err)
}

func profile(w http.ResponseWriter, r *http.Request) {
	// Check if the user is not logged in
	if !alreadyLoggedIn(r) {
		// r.Method = http.MethodGet
		// http.Redirect(w, r, "/#/login", http.StatusSeeOther)
		return
	}

	var userRecord records.UserRecord
	userRecord.Db = db // DB pointer

	var userModel records.User
	userModel, err = userRecord.GetUser(sessionMap[getUuid(r)].Email) // Used to get the currently logged in user

	// To update the user's profile picture
	// TODO other user attributes
	if r.Method == http.MethodPost {
		w.Header().Set("Content-Type", "application/json") // Set the header type to json
		fileName := createUserImage(r)                     // Get the file path of the post's image and the name

		// Used to implement PostRecord interface from models to use it's methods

		userModel.UserImage = fileName // Set the user's image
		userRecord.UpdateUser(userModel)
		userModel, err = userRecord.GetUser(sessionMap[getUuid(r)].Email)
		err = json.NewEncoder(w).Encode(userModel.UserImage)
		check(err)
		return
	}

	w.Header().Set("Content-Type", "application/json") // Set the header type to json

	// Used to implement PostRecord interface from models to use it's methods
	var postRecord records.PostRecord
	postRecord.Db = db

	var postModel records.Post
	postModel.PostAuthor = sessionMap[getUuid(r)].Email // Get the user from his cookie using sessionMap

	posts := postRecord.GetAllPostsAuthor(postModel.PostAuthor) // Get this user's posts

	// temp struct to send the posts and the user's profile image
	type temp struct {
		Posts     []records.Post
		UserImage string
		UserName  string
	}
	var obj temp
	obj.Posts = posts
	obj.UserImage = userModel.UserImage
	obj.UserName = userModel.FirstName + " " + userModel.LastName
	// Encode posts list and send it to the client
	err = json.NewEncoder(w).Encode(obj)
	check(err)
}

func users(w http.ResponseWriter, r *http.Request) {
	// Check if the user is not logged in
	if !alreadyLoggedIn(r) {
		// r.Method = http.MethodGet
		// http.Redirect(w, r, "/#/login", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "application/json") // Set the header type to json

	userId := r.URL.Query().Get("id") // Get the ID from query parameter

	// Used to implement UserRecord interface from models to use it's methods
	var userRecord records.UserRecord
	userRecord.Db = db // DB pointer

	var userModel records.User // Get the visited user
	userModel, err = userRecord.GetUserId(userId)
	check(err)

	var currentUser records.User // Get the currently logged in user
	currentUser, err = userRecord.GetUser(sessionMap[getUuid(r)].Email)
	check(err)

	// Used to implement PostRecord interface from models to use it's methods
	var postRecord records.PostRecord
	postRecord.Db = db

	var postModel records.Post
	postModel.PostAuthor = userModel.Email // Get the user from his email referenced by his ID

	posts := postRecord.GetAllPostsAuthor(postModel.PostAuthor) // Get this user's posts

	var followRecord records.FollowRecord
	followRecord.Db = db

	followed := followRecord.IsFollowed(currentUser.Email, userModel.Email) // Is this user followed or not
	fmt.Println("Followed ? ", followed)
	// Follow request sent case
	if r.Method == http.MethodPost {
		// Get the currently logged in user (Must be done again in POST method)
		currentUser, err = userRecord.GetUser(sessionMap[getUuid(r)].Email)

		var follow records.Follow
		follow.User_ = currentUser.Email // Get the currently logged in user Email

		type temp struct {
			Id     string
			Follow bool // True to follow False to unfollow
		}
		var tmp temp                               // Temporary struct to decode JSON of follow request
		err = json.NewDecoder(r.Body).Decode(&tmp) // Decode JSON from the request body
		check(err)

		// Get the visited user (Must be done again in POST method)
		userModel, err = userRecord.GetUserId(tmp.Id)

		follow.Follows_ = userModel.Email // Set the user to be followed

		if tmp.Follow == true { // Follow case
			fmt.Println("IN FOLLOW CASE")
			followRecord.InsertFollow(follow) // Perform query
		} else { // Unfollow case
			fmt.Println("IN UNFOLLOW CASE")
			followRecord.Unfollow(follow)
		}

		return
	}

	// temp struct to send the posts and the user's profile image
	type temp struct {
		Posts     []records.Post
		UserImage string
		UserName  string
		Followed  bool
	}
	var obj temp
	obj.Posts = posts
	obj.UserImage = userModel.UserImage
	obj.UserName = userModel.FirstName + " " + userModel.LastName
	obj.Followed = followed
	// Encode posts list and send it to the client
	err = json.NewEncoder(w).Encode(obj)
	check(err)
}

func search(w http.ResponseWriter, r *http.Request) {
	// Check if the user is not logged in
	if !alreadyLoggedIn(r) {
		// r.Method = http.MethodGet
		// http.Redirect(w, r, "/#/login", http.StatusSeeOther)
		return
	}

	// Decode JSON object from request

	type temp struct { // temp struct to decode JSON object from request
		SearchString string
	}

	var searchQuery temp                               // JSON object of user to get credentials from request
	err = json.NewDecoder(r.Body).Decode(&searchQuery) // Decode JSON from the request body
	check(err)

	var userRecord records.UserRecord
	userRecord.Db = db // DB pointer

	users, err := userRecord.GetUsersFirstLast(searchQuery.SearchString)
	check(err)

	// Write response with all users from the result query
	w.Header().Set("Content-Type", "application/json") // Set the header type to json
	err = json.NewEncoder(w).Encode(users)
	check(err)
}

func chat(w http.ResponseWriter, r *http.Request) {

	if !alreadyLoggedIn(r) {
		// r.Method = http.MethodGet
		// http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		current_active_user := GetUserFromCookie(r)// current active user
		 GetResponseToRoom(&current_active_user,w)
		return
	}
}

func chat_room(w http.ResponseWriter, r *http.Request)  {

	if !alreadyLoggedIn(r) {
		// r.Method = http.MethodGet
		// http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}


	//get user from currently active req
	//get second user  from hashed id
	second_user_email := r.FormValue("email") // get id from url hashed
	room_id := r.FormValue("id")//Room id
	fmt.Println("Delivered from Front end ---> id : ",room_id)
	fmt.Println("Delivered from Front end ---> email :",second_user_email)
	user := GetUserFromCookie(r)// current active user
	current_active_user := &user
	if len(second_user_email) == 0 || len(room_id) == 0 || current_active_user == nil {
		return
	}
	//fmt.Println(current_active_user)
	second_active_user := GetUserFromSessionMap(second_user_email)
	if second_active_user == nil{
		fmt.Println("No second user till now")
		//http.Redirect(w,r,"/chat?email="+key,http.StatusSeeOther)//redirection to the same route
	}

		//TODO Load All Messages for the room and send it back to the front end

		connectsocket(w, r) //upgrading connection
		// search for the room's id and start the room
		This_Room := GetTheRoom(current_active_user,room_id)
		This_Room.members[current_active_user]=true
		This_Room.members[second_active_user]=true
			if This_Room.IsRunning == false {
				// pushing this room to the Running Rooms
				RunningRooms = append(RunningRooms,This_Room)
				This_Room.Run()
			}

	// TODO: TAKE THIS FUNCTION OUT AND PUT IT HERE
	// <-test_chan
	// TODO: get all the rooms the current user has
	// TODO: If the user has no rooms, suggest to user to search for other users to chat with
	// TODO: If the user has some rooms, show them from DB and display last message sent
	// TODO: Display the chat room content when the user selects a certain room
	// TODO: If user is online make a socket connection (must be threaded)
	// TODO: If not send the message to DB

}
// Helper function to print out errors
func check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

// Helper function to get rows count
func checkCount(rows *sql.Rows) (count int) {
	var name string
	for rows.Next() {
		err := rows.Scan(&count, &name)
		check(err)
	}
	return count
}

// Helper function to create a session cookie
func createCookie(r *http.Request) http.Cookie {
	uuid, err := uuid.NewUUID()
	uuidS := uuid.String()
	check(err)
	return http.Cookie{
		Name:  "session",
		Value: uuidS,
	}
}

// Helper function to check if user is loged in or not
func alreadyLoggedIn(r *http.Request) bool {
	cookie, err := r.Cookie("session")
	if cookie == nil {
		return false
	}
	uuid := cookie.Value
	check(err)

	if _, ok := sessionMap[uuid]; ok {
		return true
	}
	return false
}

func getUuid(r *http.Request) string {
	cookie, err := r.Cookie("session")
	if cookie == nil {
		return ""
	}
	uuid := cookie.Value
	check(err)
	return uuid
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func GetUserFromCookie(r *http.Request) User {
	cookie, err := r.Cookie("session")
	check(err)
	return sessionMap[cookie.Value]
}
// for testing only
func TEST_Get_other_user(uuid string) *User {
	for _,val := range sessionMap{
		if val.Email != sessionMap[uuid].Email {
			return &val
		}
	}
	useer := sessionMap[uuid]
	return &useer
}
func GetUserFromSessionMap(email string) *User {
	for _,user := range sessionMap {
		if user.Email == email {
			return &user
		}
	}
	return nil
}

// Sends the Json data of the rooms to the front end ***
func GetResponseToRoom(user *User,w http.ResponseWriter)  {
	//data to be sent back
	type Room_data struct  {
		Room_ID string
		FirstName string //Of the second user to this user
		LastName string // Same
		Email string // Same
		Image string // profile image of the second user
		Online bool // Online
	}
	// list of rooms for a given user
	var data_list []Room_data
	var data Room_data //object to be appended when iterating
	var user_record records.UserRecord // calling database and getting a user that is not online and is in the followers
	user_record.Db = db // DB pointer
	//looping for all rooms for a given user

	for room,_ := range user.rooms {
		data.Room_ID = room.ID// setting room id to that of the room
		su := GetUserFromSessionMap(room.email) //trial to get a user from session of online
		if su == nil {
			//offline user who is a follower to the current user
			us_record,error := user_record.GetUser(room.email)
			if error != nil {
				fmt.Println("can't get the user from GETRESPONSE TOO ROOM")
			}

			data.Online = false
			data.Email = us_record.Email
			data.FirstName = us_record.FirstName
			data.LastName = us_record.LastName
			//case that there is a room between user and himself
			//this case is forbidden
			data.Image = us_record.UserImage

			data_list = append(data_list,data)
			continue
		}
		//online user who is follower to the current user
		data.Online = true
		data.FirstName = su.FirstName
		data.LastName = su.LastName
		data.Email = su.Email
		data.Image = su.UserImage
		data_list = append(data_list,data)
	}
	// remove bugged data from the data list
	var right_list []Room_data
	for _,data :=range data_list {
		//checking for the corrupt data
		if data.Email == "" ||data.Image == "" || data.Room_ID == "00000000000000000000000000000000"{
			continue
		}
		right_list = append(right_list,data)
	}
	err := json.NewEncoder(w).Encode(right_list)
	check(err)
}
//for a certain user get the room to be initiated
func GetTheRoom(user *User,Target string)*Room  {
	for room,_ := range user.rooms {
		if room.ID == Target {
			return room
		}
	}
	return nil
}