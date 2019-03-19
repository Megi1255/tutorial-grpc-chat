package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/riimi/tutorial-grpc-chat/server/domain"
	"log"
)

func main() {
	port := flag.Int("port", 40040, "port")
	flag.Parse()

	gs := NewServer(fmt.Sprintf("localhost:%d", *port))
	gs.ErrorHandler = func(user *domain.User, err error, msg string) {
		if user != nil {
			log.Print(*user)
		}
		log.Printf("%s: %v\n", msg, err)
	}
	gs.LogHandler = func(user *domain.User, format string, a ...interface{}) {
		if user != nil {
			log.Printf("SessionId(%s) ", user.Token)
		}
		log.Print(fmt.Sprintf(format, a...))
	}
	log.Fatal(gs.Run(context.Background()))
}
