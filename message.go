package main

import (
	"encoding/json"
	"log"
)

const SendMessageAction = "send-message"

/*const JoinRoomAction = "join-room"
const LeaveRoomAction = "leave-room"
const UserJoinedAction = "user-join"
const UserLeftAction = "user-left"
*/

//const JoinRoomPrivateAction = "join-room-private"
const RoomJoinedAction = "room-joined"

type Message struct {
	Message   string  `json:"message"`
	Sender    string `json:"sender"` //sender must be a string
	Target    string  `json:"target"` // id room
	//TimeStamp string
}

type temp struct { // receive frontend message
	Message string //`json:"message"`
	User    string //`json:"user"`
	Target  string //`json:"target"`
}

func (message *Message) encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	return json
}
