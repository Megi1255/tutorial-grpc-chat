package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/riimi/tutorial-grpc-chat/clean/interface/rpc"
	"log"
)

func main() {
	port := flag.Int("port", 40040, "port")
	flag.Parse()

	gs := rpc.NewServer(fmt.Sprintf("localhost:%d", *port))
	gs.ErrorHandler = func(sess *rpc.Session, err error, msg string) {
		if sess != nil {
			log.Print(*sess)
		}
		log.Printf("%s: %v\n", msg, err)
	}
	gs.LogHandler = func(sess *rpc.Session, format string, a ...interface{}) {
		if sess != nil {
			log.Printf("SessionId(%s) ", sess.ID())
		}
		log.Print(fmt.Sprintf(format, a...))
	}
	gs.Run(context.Background())
}
