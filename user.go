package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"simple-photoblog-repo/records"
	"github.com/gorilla/websocket"
)

type User struct {
	Email          string
	Password       string
	FirstName      string
	LastName       string
	VisualImpaired bool
	// The following are added from the settings
	UserImage     string         // Image path of the user
	Followers     []string       // These strings are the followers users emails
	Following     []string       // These strings are the followed users emails
	Private       bool           // Private profile or not
	ProfilePosts  []records.Post // The user posts
	NewsFeedPosts []records.Post // The posts shown in the user's newsfeed
	//chat attributes
	conn  *websocket.Conn // socket connection for all users
	send  chan []byte     // channel to help in sending the message from a user to another
	rooms map[*Room]bool  // all the rooms that a client has registered to

}

// Initialize a new user
func newUser(email string, password string, firstName string, lastName string, visualyImpaired bool,
	private bool) User {
	return User{
		Email:          email,
		Password:       password,
		FirstName:      firstName,
		LastName:       lastName,
		VisualImpaired: visualyImpaired,
		Followers:      make([]string, 10),
		Following:      make([]string, 10),
		Private:        private,
		ProfilePosts:   make([]records.Post, 10),
		NewsFeedPosts:  make([]records.Post, 10),
		//chat attributes setting
		conn:  nil, // is nil until the conversation room is picked and then it is connected
		send:  make(chan []byte), //channel to recieve and send messages
		rooms: make(map[*Room]bool),
	}
}

// this function reads from the socket connection in a thread and always recieves a message permanently
func (client *User) read() {
	defer func() {
		 //client.disconnect() //this function closes connection with user
	}()
	for {
		_, jsonMessage, err := client.conn.ReadMessage() // connection related function used to read from socket connection
		if err != nil {
			log.Println("hello---from read",err)

			break
		}
		// fmt.Println(p)
		fmt.Println("--------->read User is :  ",client.FirstName)

		client.handleNewMessage(jsonMessage) //this function is used to unmarshal the bytes message sent from a connection to the messages class
		//and then sending the message into the room broadcast channel to be broadcasted for all users
	}
	//..client.chatserver.br
}

//this function is used in read function to broadcast a certain message
func (client *User) handleNewMessage(jsonMessage []byte) {
	var unmarshaledMessage Message
	err := json.NewDecoder(bytes.NewReader(jsonMessage)).Decode(&unmarshaledMessage)
	//push unmarshaled message in a list of all messages in the server
	if err != nil {
		fmt.Println("Error in Handle new message ----->")
	}

		fmt.Println("unmarshalled message",unmarshaledMessage)


	unmarshaledMessage.Sender = client.Email
	roomID := unmarshaledMessage.Target // set the target room for the message

	This_Room := GetTheRoom(client,roomID)
	if This_Room != nil{
		This_Room.broadcast <- &unmarshaledMessage
	}

}

// this function is the send function to each user
// forever this function send messages when there is a message on the Send channel
// there is no work required to this function
func (client *User) write() {
	for {
		select {
		case mymessage, ok := <-client.send: // case when there is a message on the send channel of a user
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{}) // if there is an error close the connection
				return
			}
			w, err := client.conn.NextWriter(websocket.TextMessage) //other wise create a writer
			if err != nil {
				return
			}
			//fmt.Println("----->Write----->",mymessage)
			var unmarshaledMessage Message
			err = json.NewDecoder(bytes.NewReader(mymessage)).Decode(&unmarshaledMessage)

			if err != nil {
				fmt.Println("Error in Handle new message ----->")
			}
			fmt.Println("---->write Writer is : message :  ",client.FirstName,unmarshaledMessage)
			w.Write(mymessage) // use this writer to write to the connection the text message
			w.Close()// close the fucking connection when you are done
		}
	}
}

// func deletPost() TODO
func (client *User) disconnect() {
	//TODO we have to create a way for the server to keep track of the room like global map of rooms to close conn
 	close(client.send)
 	client.conn.Close()
 }

func (user *User) CreateAllRoomsForAGivenUser() {
	//Loop on all followers of this current user and create a room between these 2
	for _,follower := range user.Followers {
		//create a new room
		Room := NewRoom(user.Email,follower)
		//check if there is a room between them

		if user.isRoomBetweenUsers(Room) == true {
			continue
		}

		//append new room to the map rooms
		user.rooms[Room] = true
		fmt.Println("Room Created with User....  ",user.rooms[Room])
		//append this new room to the server list of rooms
		ServerAllRooms = append(ServerAllRooms,Room)
		//append this new room to the un-recorded rooms of the database
		UnRecordedRooms = append(UnRecordedRooms,Room)
	}
}

//check if the room is registered in the rooms map or not
//maybe not needed
func(user *User) isRoomBetweenUsers(room *Room)bool {
	_,ok := user.rooms[room]; if ok {
		return true
	}
	return false
}