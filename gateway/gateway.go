package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	gw "github.com/riimi/tutorial-grpc-chat/pb"
	"google.golang.org/grpc"
	"log"
	"net/http"
)

var (
	EndPoint = flag.String("endpoint", "localhost:40040", "endpoint of chatserver")
	port     = flag.Int("port", 8081, "gateway port")
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	if err := gw.RegisterChatServiceHandlerFromEndpoint(ctx, mux, *EndPoint, opts); err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *port), mux))
}
