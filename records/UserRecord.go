package records

//                       IMPLEMENTS MODELS

import (
	"database/sql"
	"path"
	"strconv"
	"strings"
	"simple-photoblog-repo/models"
)

type User struct {
	Id               string
	Email            string
	Password         string
	FirstName        string
	LastName         string
	VisuallyImpaired bool
	UserImage        string
	Private          bool
}

type UserRecord struct {
	Db *sql.DB
}

func (user User) GetId() string {
	return user.Id
}

func (user User) GetEmail() string {
	return user.Email
}

func (user User) GetPassword() string {
	return user.Password
}

func (user User) GetFirstName() string {
	return user.FirstName
}

func (user User) GetLastName() string {
	return user.LastName
}

func (user User) GetVisuallyImpaired() bool {
	return user.VisuallyImpaired
}

func (user User) GetUserImage() string {
	return user.UserImage
}

func (user User) GetPrivate() bool {
	return user.Private
}

// UserRecord Methods: ...

func (userRecord *UserRecord) InsertUser(user models.User) {
	// user implements user interface from package models
	email := user.GetEmail()
	userid := user.GetId()
	password := user.GetPassword()
	firstName := user.GetFirstName()
	lastName := user.GetLastName()
	visuallyImpaired := user.GetVisuallyImpaired()
	userImage := user.GetUserImage()
	private := user.GetPrivate()

	stmt, err := userRecord.Db.Prepare(`INSERT INTO users (email, userid, password, firstname, lastname, visuallyimpaired,
		userimage, private) 
		VALUES ("` + email + `", "` + userid + `", "` + password + `", "` + firstName + `", "` + lastName + `",
		 "` + strconv.Itoa(boolToInt(visuallyImpaired)) + `",
		  "` + userImage + `", "` + strconv.Itoa(boolToInt(private)) + `")`)
	check(err)
	_, err = stmt.Exec() // Execute the statement
	check(err)
}

func (userRecord *UserRecord) GetUser(email string) (User, error) {
	var user User
	var tempVisuallyImpaired int // To get the integer value of visuallyimpaired from query
	var tempPrivate int          // To get the integer value of private from query
	row, err := userRecord.Db.Query(`SELECT * FROM users WHERE email = "` +
		email + `" ;`) // Get the required row from posts table by it's id
	check(err)

	for row.Next() {
		err = row.Scan(&user.Email, &user.Id, &user.Password, &user.FirstName, &user.LastName,
			&tempVisuallyImpaired, &user.UserImage, &tempPrivate) // Get the image name
		check(err)
		user.UserImage = path.Join("/", "users_images", user.UserImage)
	}

	user.VisuallyImpaired = intToBool(tempVisuallyImpaired)
	user.Private = intToBool(tempPrivate)

	return user, err
}

func (userRecord *UserRecord) GetUserId(userId string) (User, error) {
	var user User
	var tempVisuallyImpaired int // To get the integer value of visuallyimpaired from query
	var tempPrivate int          // To get the integer value of private from query
	row, err := userRecord.Db.Query(`SELECT * FROM users WHERE userid = "` +
		userId + `" ;`) // Get the required row from posts table by it's id
	check(err)

	for row.Next() {
		err = row.Scan(&user.Email, &user.Id, &user.Password, &user.FirstName, &user.LastName,
			&tempVisuallyImpaired, &user.UserImage, &tempPrivate) // Get the image name
		check(err)
		user.UserImage = path.Join("/", "users_images", user.UserImage)
	}

	user.VisuallyImpaired = intToBool(tempVisuallyImpaired)
	user.Private = intToBool(tempPrivate)

	return user, err
}

func (userRecord *UserRecord) UpdateUser(user models.User) {

	email := user.GetEmail()
	password := user.GetPassword()
	firstName := user.GetFirstName()
	lastName := user.GetLastName()
	visuallyImpaired := user.GetVisuallyImpaired()
	userImage := user.GetUserImage()
	private := user.GetPrivate()

	_, err := userRecord.Db.Exec(`update users set email = ?, password = ?, firstname = ?,
	lastname = ?, visuallyimpaired = ?, userimage = ?,
	private = ? where email = ?`, email, password, firstName, lastName, strconv.Itoa(boolToInt(visuallyImpaired)),
		userImage, strconv.Itoa(boolToInt(private)), email)

	check(err)
}

// Get all users by first name or last name
func (userRecord *UserRecord) GetUsersFirstLast(userName string) ([]User, error) {
	names := strings.Split(userName, " ") // List of search paramter strings splitted by blank space
	var err error
	var user User
	var users []User
	var tempVisuallyImpaired int // To get the integer value of visuallyimpaired from query
	var tempPrivate int          // To get the integer value of private from query
	if len(names) >= 2 {
		// Get all users from the search result as firstname lastname and vice versa
		rows, err := userRecord.Db.Query(`SELECT email, userid, firstname, lastname,
		visuallyimpaired, userimage, private FROM users WHERE firstname LIKE "%` +
			names[0] + `%" OR firstname LIKE "%` + names[1] + `%" OR lastname LIKE "%` +
			names[0] + `%" OR lastname LIKE "%` + names[1] + `%" ;`)
		check(err)

		for rows.Next() {
			err = rows.Scan(&user.Email, &user.Id, &user.FirstName, &user.LastName,
				&tempVisuallyImpaired, &user.UserImage, &tempPrivate)
			check(err)
			user.UserImage = path.Join("/", "users_images", user.UserImage)
			user.VisuallyImpaired = intToBool(tempVisuallyImpaired)
			user.Private = intToBool(tempPrivate)

			users = append(users, user)
		}
	} else {
		// Get all users from the search result as firstname lastname and vice versa
		rows, err := userRecord.Db.Query(`SELECT email, userid, firstname, lastname,
		visuallyimpaired, userimage, private FROM users WHERE firstname LIKE "%` +
			names[0] + `%" OR lastname LIKE "%` + names[0] + `%";`)
		check(err)

		for rows.Next() {
			err = rows.Scan(&user.Email, &user.Id, &user.FirstName, &user.LastName,
				&tempVisuallyImpaired, &user.UserImage, &tempPrivate)
			check(err)
			user.UserImage = path.Join("/", "users_images", user.UserImage)
			user.VisuallyImpaired = intToBool(tempVisuallyImpaired)
			user.Private = intToBool(tempPrivate)

			users = append(users, user)
		}
	}

	return users, err
}

// Get all the followed users info
func (userRecord *UserRecord) GetFollowersInfo(followers []string) []User {
	var user User
	var users []User
	// var tempVisuallyImpaired int // To get the integer value of visuallyimpaired from query
	// var tempPrivate int          // To get the integer value of private from query

	for _, follower := range followers {
		row, err := userRecord.Db.Query(`SELECT email, userid, firstname, lastname, userimage FROM users 
		WHERE email = "` + follower + `" ;`) // Get the required row from posts table by it's id
		check(err)

		for row.Next() {
			err = row.Scan(&user.Email, &user.Id, &user.FirstName, &user.LastName,
				&user.UserImage) // Get the image name
			check(err)
			user.UserImage = path.Join("/", "users_images", user.UserImage)
		}

		users = append(users, user)
	}

	return users
}

// Helper function to get integer value of boolean
func boolToInt(bitSet bool) int {
	var bitSetVar int
	if bitSet {
		bitSetVar = 1
	}
	return bitSetVar
}

// Helper function to get boolean value of integer
func intToBool(bitSet int) bool {
	var bitSetVar bool
	if bitSet == 1 {
		bitSetVar = true
	}
	return bitSetVar
}
