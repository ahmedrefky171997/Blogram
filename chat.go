package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var tpl *template.Template

// func init() {
// 	tpl = template.Must(template.ParseGlob("chatroom/chatroomwithsockets/templates/*"))
// }
// func temp() {
// 	ws := newserver() // start the server
// 	// http.HandleFunc("/", intro)

// 	// http.ListenAndServe(":8080", nil)
// }

// func intro(w http.ResponseWriter, req *http.Request) {
// 	tpl.ExecuteTemplate(w, "intro.html", nil)

// }

func connectsocket( w http.ResponseWriter, req *http.Request) { //ws fesa

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	} //type upgrader used to upgrade a certain request to a full duplex connection

	upgrader.CheckOrigin = func(r *http.Request) bool {
		fmt.Println("Inside upgrader")
		return true
	}

	Current_active_user := GetUserFromCookie(req) // get a user by using the cookie and before it gets upgraded
	conn, err := upgrader.Upgrade(w, req, nil) // upgrade request to full duplex connection
	if err != nil {
		log.Print("hello",err)
	}

	Current_active_user.conn = conn // setting connection to that of the user's connection
	//fmt.Println("User",Current_active_user.FirstName+" "+Current_active_user.LastName,"Connected Successfully")
	go Current_active_user.write()
	go Current_active_user.read()


}
