package main

import (
	"log"
	"net/http"
)

const (
	port       = "8080"
	hostport   = ":" + port
	publishDir = "./"
)

func main() {
	fs := http.FileServer(http.Dir(publishDir))
	http.Handle("/", fs)

	log.Printf("Running HTTP Server on http://localhost:%s/colortress.html", port)
	if err := http.ListenAndServe(hostport, nil); err != nil {
		log.Fatal(err)
	}
}
