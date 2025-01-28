package main

import (
	"log"
	"net/http"
)

const (
	hostport   = ":8080"
	publishDir = "./"
)

func main() {
	fs := http.FileServer(http.Dir(publishDir))
	http.Handle("/", fs)

	if err := http.ListenAndServe(hostport, nil); err != nil {
		log.Fatal(err)
	}
}
