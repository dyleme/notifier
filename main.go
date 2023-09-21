package main

import (
	"log"
	"time"
)

func main() {
	_, err := time.LoadLocation("UTC+3")
	if err != nil {
		log.Fatal(err)
	}
}
