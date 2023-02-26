package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	log.Println("server listen on :5000")
	if err := http.ListenAndServe(":5000", nil); err != nil {
		panic(err)
	}
}
