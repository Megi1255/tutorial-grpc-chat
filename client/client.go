package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/riimi/tutorial-grpc-chat/pb"
	"github.com/zserge/lorca"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
)

type ChatClient struct {
	conn *grpc.ClientConn
	rpc  pb.ChatServiceClient
	ui   lorca.UI

	Connected bool
	Id        string
}

func NewGophersClient(w, h int) *ChatClient {
	ui, err := lorca.New("", "", w, h)
	if err != nil {
		log.Fatal(err)
	}

	return &ChatClient{
		ui: ui,
	}
}

func (c *ChatClient) Connect(addr string) error {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	client := pb.NewChatServiceClient(conn)
	c.conn = conn
	c.rpc = client

	return nil
}

func (c *ChatClient) Close() {
	c.conn.Close()
}

func (c *ChatClient) Subscribe() {
	ctx := context.Background()
	stream, err := c.rpc.Subscribe(ctx, &empty.Empty{})
	if err != nil {
		log.Printf("[Hello] failed to connect: %v", err)
		c.PushMessage(err.Error())
		return
	}

	go c.readPump(stream)
}

func (c *ChatClient) readPump(stream pb.ChatService_SubscribeClient) {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			log.Printf("[readpump] recv got EOF: %v", err)
			c.Close()
			return
		} else if err != nil {
			log.Printf("[readpump] failed to recv: %v", err)
			return
		}
		if c.Id == "" {
			c.Id = in.Id
		}

		c.PushMessage(fmt.Sprintf("id: %s, text: %s", in.Id, in.Text))
	}
}

func (c *ChatClient) Send(msg string) {
	if _, err := c.rpc.Send(context.Background(), &pb.Message{
		Id:   c.Id,
		Text: msg,
	}); err != nil {
		log.Printf("[chat] failed to send message: %v", err)
	}
}

func (c *ChatClient) PushMessage(msg string) {
	if err := c.ui.Eval(fmt.Sprintf(`
        window.app.pushMessage('%s');
	`, msg)).Err(); err != nil {
		log.Printf("[PushMessage] %v, %s", err, msg)
	}
}

func (c *ChatClient) Run() {
	if c.conn == nil {
		return
	}

	if err := c.ui.Bind("subscribe", c.Subscribe); err != nil {
		log.Fatal(err)
	}
	if err := c.ui.Bind("send", c.Send); err != nil {
		log.Fatal(err)
	}

	fp, err := os.Open("ui.html")
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()

	html, err := ioutil.ReadAll(fp)
	if err != nil {
		log.Fatal(err)
	}

	c.ui.Load("data:text/html," + url.PathEscape(string(html)))
	<-c.ui.Done()
}

func main() {
	width := flag.Int("width", 800, "window size width")
	height := flag.Int("height", 450, "window size height")
	serverAddr := flag.String("addr", "localhost:40040", "grpc server address")
	flag.Parse()

	gophers := NewGophersClient(*width, *height)
	if err := gophers.Connect(*serverAddr); err != nil {
		log.Fatal(err)
	}
	gophers.Run()
}
