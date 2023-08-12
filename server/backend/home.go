package main

import (
	"fmt"
	"myflagsubmitter/common"
	"sync"

	"github.com/gorilla/websocket"
)

func updateNewFlags(flags []common.Flag) {
	//fmt.Println("Updating websocket clients with flags:", flags)
	var wg sync.WaitGroup
	for _, client := range clients {
		wg.Add(1)
		go updateClient(&wg, client.connection, flags)
	}
	wg.Wait()
	//fmt.Println("Update done")
}

func updateClient(wg *sync.WaitGroup, client *websocket.Conn, msg []common.Flag) {
	if wg != nil {
		defer wg.Done()
	}

	//set a maximum number of flags to send within the same packet
	maximum_upload_flags := 50
	for i := 0; i*maximum_upload_flags < len(msg); i++ {
		lower_bound := i * maximum_upload_flags
		upper_bound := (i + 1) * maximum_upload_flags
		if upper_bound > len(msg) {
			upper_bound = len(msg)
		}
		err := client.WriteJSON(msg[lower_bound:upper_bound])
		if err != nil {
			fmt.Println("Error sending message:", err)
			client.Close()

			if len(clients) == 1 {
				clients = make([]WebSocketClient, 0)
			} else {
				var indexToRemove int
				for i, c := range clients {
					if c.connection == client {
						indexToRemove = i
						break
					}
				}
				if indexToRemove != len(clients)-1 {
					clients = append(clients[:indexToRemove], clients[indexToRemove+1:]...)
				} else {
					clients = clients[:indexToRemove]
				}
			}
		}
	}
}
