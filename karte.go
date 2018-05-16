package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"
)

func main() {
	session, sessionErr = mgo.Dial("localhost:27017")
	if sessionErr != nil {
		panic(sessionErr)
	}
	defer session.Close()

	fmt.Printf("Hello, world.\n")

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	msgCounter = r1.Int31()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	roomBela := newRoom("bela")
	go roomBela.run()

	http.HandleFunc("/bela", func(w http.ResponseWriter, r *http.Request) {
		log.Println("request", r)
		serveWs(roomBela, w, r)
	})

	http.HandleFunc("/room_info", func(w http.ResponseWriter, r *http.Request) {

		js := roomBela.info()

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)

	})

	if runtime.GOOS == "darwin" {
		http.ListenAndServe(":3000", nil)
	} else {
		http.ListenAndServe(":8000", nil)
	}
}
