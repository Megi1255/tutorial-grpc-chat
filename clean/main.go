package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/riimi/tutorial-grpc-chat/clean/infrastructure/chatserver"
	"log"
)

func main() {
	port := flag.Int("port", 40040, "port")
	flag.Parse()

	gs := chatserver.NewServer(fmt.Sprintf("localhost:%d", *port))
	gs.ErrorHandler = func(sess *chatserver.Session, err error, msg string) {
		if sess != nil {
			log.Print(*sess)
		}
		log.Printf("%s: %v\n", msg, err)
	}
	gs.LogHandler = func(sess *chatserver.Session, format string, a ...interface{}) {
		if sess != nil {
			log.Printf("SessionId(%s) ", sess.ID())
		}
		log.Print(fmt.Sprintf(format, a...))
	}
	gs.Run(context.Background())
}
