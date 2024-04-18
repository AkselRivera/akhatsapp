package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	hub := newHub()

	go hub.run()

	serverMux := http.NewServeMux()

	serverMux.Handle("/", http.FileServer(http.Dir("public")))

	serverMux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHandler(hub, w, r)
	})

	fmt.Println("    ___    __ __ __          __                        ")
	fmt.Println("   /   |  / //_// /_  ____ _/ /__________ _____  ____  ")
	fmt.Println(`  / /| | / ,<  / __ \/ __  / __/ ___/ __  / __ \/ __ \ `)
	fmt.Println(" / ___ |/ /| |/ / / / /_/ / /_(__  ) /_/ / /_/ / /_/ / ")
	fmt.Println(`/_/  |_/_/ |_/_/ /_/\__,_/\__/____/\__,_/ .___/ .___/  `)
	fmt.Println("                                       /_/   /_/       \n\n\n")

	log.Println("Starting server on :8080")
	http.ListenAndServe(":8080", serverMux)

}
