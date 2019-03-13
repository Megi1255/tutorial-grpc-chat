package main

import (
	"context"
	"flag"
	"fmt"
	"log"
)

func main() {
	port := flag.Int("port", 40040, "port")
	flag.Parse()

	gs := NewServer(fmt.Sprintf("localhost:%d", *port))
	gs.ErrorHandler = func(sess *Session, err error, msg string) {
		if sess != nil {
			log.Print(*sess)
		}
		log.Printf("%s: %v\n", msg, err)
	}
	gs.LogHandler = func(sess *Session, format string, a ...interface{}) {
		if sess != nil {
			log.Printf("SessionId(%s) ", sess.Id)
		}
		log.Print(fmt.Sprintf(format, a...))
	}
	gs.Run(context.Background())
}
