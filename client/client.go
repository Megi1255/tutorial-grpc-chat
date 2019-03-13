package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/riimi/tutorial-grpc-chat/pb"
	"github.com/zserge/lorca"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"time"
)

type ChatClient struct {
	addr string
	rpc  pb.ChatServiceClient
	ui   UI
	conn net.Conn

	Connected bool
	Id        string
}

func NewGophersClient(w, h int, addr string) *ChatClient {
	ui, err := lorca.New("", "", w, h)
	if err != nil {
		log.Fatal(err)
	}

	return &ChatClient{
		ui:   UI{ui},
		addr: addr,
	}
}

func (c *ChatClient) Close() {
	c.conn.Close()
}

func (c *ChatClient) Connect() error {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(c.addr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	client := pb.NewChatServiceClient(conn)
	c.rpc = client

	return nil
}

func (c *ChatClient) Subscribe() {
	ctx := context.Background()
	stream, err := c.rpc.Subscribe(ctx)
	if err != nil {
		log.Printf("[Hello] failed to connect: %v", err)
		c.ui.PushMessage(err.Error())
		return
	}

	go c.readPump(stream)
}

func (c *ChatClient) readPump(stream pb.ChatService_SubscribeClient) {
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			log.Printf("[readpump] recv got EOF: %v", err)
			c.Close()
			return
		} else if err != nil {
			log.Printf("[readpump] failed to recv: %v", err)
			return
		}

		//c.PushMessage(fmt.Sprintf("id: %s, text: %s", in.Id, in.Text))
	}
}

func (c *ChatClient) Login() {
	if err := c.Connect(); err != nil {
		log.Printf("[chat] failed to connect: %v", err)
		c.ui.PushMessage(fmt.Sprintf("[chat] failed to connect: %v", err))
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	res, err := c.rpc.Login(ctx, &pb.LoginRequest{
		Name: "test",
	})
	if err != nil {
		log.Printf("[chat] failed to send message: %v", err)
		return
	}

	c.ui.PushMessage(fmt.Sprintf("[Login] token: %s", res.Token))
	c.ui.Eval(`window.app.connected = true`)
}

func (c *ChatClient) Logout() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	_, err := c.rpc.Logout(ctx, &pb.LogoutRequest{
		Token: c.Id,
	})
	if err != nil {
		log.Printf("[chat] failed to send message: %v", err)
		return
	}

	c.Close()
	c.ui.PushMessage("[Logout]")
	c.ui.Eval(`window.app.connected = false`)
}

func (c *ChatClient) Run() {

	if err := c.ui.Bind("connect", c.Login); err != nil {
		log.Fatal(err)
	}
	if err := c.ui.Bind("disconnect", c.Logout); err != nil {
		log.Fatal(err)
	}
	if err := c.ui.Bind("send", c.Subscribe); err != nil {
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

type UI struct {
	lorca.UI
}

func (u *UI) PushMessage(msg string) {
	if err := u.Eval(fmt.Sprintf(`
        window.app.pushMessage('%s');
	`, msg)).Err(); err != nil {
		log.Printf("[PushMessage] %v, %s", err, msg)
	}
}

func main() {
	width := flag.Int("width", 800, "window size width")
	height := flag.Int("height", 450, "window size height")
	serverAddr := flag.String("addr", "localhost:40040", "grpc server address")
	flag.Parse()

	gophers := NewGophersClient(*width, *height, *serverAddr)
	gophers.Run()
}
