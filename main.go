package main

import (
	"github.com/jeremymichel/sse-server/sse"
	"log"
	"net/http"
)

func main() {
	broker := sse.NewSSEBroker()

	http.Handle("/", broker)

	log.Fatalln(http.ListenAndServe(":3000", nil))
}
