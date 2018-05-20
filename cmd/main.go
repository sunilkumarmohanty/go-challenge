package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/sunilkumarmohanty/go-challenge/handlers"
)

func main() {
	listenAddr := flag.String("http.addr", ":8080", "http listen address")
	flag.Parse()
	http.HandleFunc("/numbers", handlers.NumberHandler)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
