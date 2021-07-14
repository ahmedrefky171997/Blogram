package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	_ "golang.org/x/crypto/acme"
)

// todo rooms should have unique id
// todo for database
// room interface which will has get room name or id and if it private or not (done)
//RoomRepository  which will add new room and find room by id
// for adding room we just insert room name ,id,privacy

//
// implemented functions
// get name
// register client
// unregister client
// 	broadcasting
// start
// create new room

type Room struct {
	name string `json:"name"`
	ID         string // list of user emails Authors to a room
	IsRunning  bool
	broadcast  chan *Message
	members    map[*User]bool
	email      string
	close      chan bool
	//private    bool `json:"private"`
	// id  string
}

func NewRoom(first_email,second_email string) *Room {
	// create hash for the first email and the second and return their xor
	id := XoredHashed(first_email,second_email)
	return &Room{
		ID: id,// ids of the users
		IsRunning: false,
		email: second_email, //email of the other user
		broadcast:  make(chan *Message,1),
		members:    make(map[*User]bool),
		close: 		make(chan bool),

	}

}

// start the room to recieve on its members
func (room *Room) Run() {
	for {
		select {
		/*
		case myclient := <-room.Register:
			room.newclientjoined(myclient)
		case myclient := <-room.unregister:
			room.clientleft(myclient)
		 */
			case bool :=  <-room.close:
				if bool == true {
					room.Disconnect()
					break
				}
			case message := <-room.broadcast:
			room.broadcastmessage(message.encode())


		}

	}
}

//fesa
func (room *Room) newclientjoined(myclient *User) {
	room.members[myclient] = true
	mymessage := &Message{
		Target:    room.name,
		Message: fmt.Sprintf(" %s welcomeMessage", myclient.FirstName),
	}
	room.broadcastmessage(mymessage.encode())
}

//fesa
func (room *Room) clientleft(myclient *User) {
	if _, ok := room.members[myclient]; ok {
		delete(room.members, myclient)
	}
}

//broadcasting the message to a list of members
func (room *Room) broadcastmessage(message []byte) {
	//fmt.Println("BroadCastMessage --------->",message)
	for member := range room.members {
		fmt.Println("Broadcasting To ",room.email)
		member.send <- message
	}
}
func (room *Room) getname() string {
	return room.name
}
//func (room *Room) getprivacy() bool {
//	return room.private
//}

/*
func (room room)getid()string{
	return room.id
}
*/
//returns an id between two strings
func XoredHashed(s1,s2 string)string  {
	h1 := md5.Sum([]byte(s1))//hash of the first email
	h2 := md5.Sum([]byte(s2))//hash of the second email
	result := make([]byte,len(h1))//result string
	for i:=0;i<len(h1)-1;i++{
		result[i] = h1[i] ^ h2[i] //xor of the two emails
	}
	return hex.EncodeToString(result)
}

func (room *Room) Disconnect()  {
	for member,_ := range room.members {
		member.conn.Close()
		close(room.broadcast)
		close(room.close)
	}
}