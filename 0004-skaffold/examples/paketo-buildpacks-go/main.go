package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "Hello World go!")
	})

	log.Fatal(http.ListenAndServe(":3000", nil))
}
