package config

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"os"
)

func InitDB() *sql.DB {

	// Load the environment variables from env_var.env file
	err := godotenv.Load("config/env_var.env")
	check(err)
	dbUser := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbIP := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Access the database using string s
	s := dbUser + ":" + dbPassword + "@tcp(" + dbIP + ":" + dbPort + ")/" + dbName + "?charset=utf8"
	db, err := sql.Open("mysql", s)

	// Create the users table if doesn't exist
	stmt, err := db.Prepare(`CREATE TABLE IF NOT EXISTS users (
		email VARCHAR(256) NOT NULL,
		userid VARCHAR(256) NOT NULL,
		password VARCHAR(256) NOT NULL,
		firstname VARCHAR(256) NOT NULL,
		lastname VARCHAR(256) NOT NULL,
		visuallyimpaired BOOLEAN,
		userimage VARCHAR(1024) NOT NULL,
		private BOOLEAN,
		PRIMARY KEY (email)
		);`)
	check(err)

	_, err = stmt.Exec() // Execute the statement
	check(err)

	// Create the posts table if doesn't exist
	stmt, err = db.Prepare(`CREATE TABLE IF NOT EXISTS posts (
		id INT NOT NULL AUTO_INCREMENT,
		author VARCHAR(256) NOT NULL,
		username VARCHAR(256) NOT NULL,
		date TIMESTAMP NOT NULL,
		caption VARCHAR(256) NOT NULL,
		imagecaption VARCHAR(2048) NOT NULL,
		imagename VARCHAR(1024) NOT NULL,
		postaudio VARCHAR(1024) NOT NULL,
		PRIMARY KEY (id),
		FOREIGN KEY (author) REFERENCES users(email)
		);`)
	check(err)

	_, err = stmt.Exec() // Execute the statement
	check(err)

	// Create the followers table if doesn't exist
	stmt, err = db.Prepare(`CREATE TABLE IF NOT EXISTS followers (
		user_ VARCHAR(100) NOT NULL,
		follows_ VARCHAR(100) NOT NULL,
		date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_,follows_),
		FOREIGN KEY (user_) REFERENCES users(email),
		FOREIGN KEY (follows_) REFERENCES users(email)
		);`)
	check(err)

	_, err = stmt.Exec() // Execute the statement
	check(err)
	// db.SetConnMaxLifetime(time.Second*30)
	return db
}

// Helper function to print out errors
func check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
