package main

import (
	"log"
	"net/http"
)

func main() {
	gr := NewGomibakoRepository()
	go gr.RunBroker()

	handler := NewServerHandler(gr)
	log.Fatal(http.ListenAndServe(":8000", handler))
}
