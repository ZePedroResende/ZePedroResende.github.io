package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	generated := "./generated"

	if _, err := os.Stat(generated); os.IsNotExist(err) {
		log.Fatalf("Directory %v does not exist", generated)
		return
	}

	fs := http.FileServer(http.Dir(generated))
	http.Handle("/", fs)

	log.Print("Listening on http://localhost:3000")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
