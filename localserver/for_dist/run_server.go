package main

// for distributions

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	port     = "8080"
	hostport = ":" + port
)

func main() {
	var execDir, err = os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	var publishDir = filepath.Join(filepath.Dir(execDir), "html")

	fs := http.FileServer(http.Dir(publishDir))
	http.Handle("/", fs)

	fmt.Printf("\nhttp://localhost:%s/colortress.html をブラウザで開いてください\n", port)
	fmt.Printf("終了時は Ctrl + C を押してください\n")
	if err := http.ListenAndServe(hostport, nil); err != nil {
		log.Fatalf("すでに同じアプリケーションが開かれている可能性があります。一旦閉じてから起動してください:\n%v", err)
	}
}
