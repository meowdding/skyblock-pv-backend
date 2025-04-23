package main

import (
	"io"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello, World!")
	})

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		panic(err)
	}
}
