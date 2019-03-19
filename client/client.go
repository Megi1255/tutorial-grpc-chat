package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/riimi/tutorial-grpc-chat/clean/interface/rpc/protocol"
	"github.com/zserge/lorca"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"time"
)

type ChatClient struct {
	addr   string
	rpc    protocol.ChatServiceClient
	stream protocol.ChatService_SubscribeClient
	ui     UI
	conn   *grpc.ClientConn

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

func (c *ChatClient) connect() error {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(c.addr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	client := protocol.NewChatServiceClient(conn)
	c.rpc = client
	c.conn = conn

	return nil
}

func (c *ChatClient) Subscribe(ctx context.Context) {
	md := metadata.New(map[string]string{"myheader-token": c.Id})
	ctx = metadata.NewOutgoingContext(ctx, md)

	stream, err := c.rpc.Subscribe(ctx)
	if err != nil {
		log.Printf("[Hello] failed to connect: %v", err)
		c.ui.PushMessage(err.Error())
		return
	}

	c.stream = stream
	go c.readPump()
}

func (c *ChatClient) Send(text string) {
	if err := c.stream.Send(&protocol.Message{
		Timestamp: ptypes.TimestampNow(),
		Event: &protocol.Message_Message_{
			Message: &protocol.Message_Message{
				Name:    "test",
				Message: text,
			},
		},
	}); err != nil {
		log.Printf("[Hello] failed to send: %v", err)
		c.ui.PushMessage(err.Error())
	}

	c.ui.Eval(`window.app.text1 = ''`)
}

func (c *ChatClient) readPump() {
	defer c.stream.CloseSend()
	defer c.Close()

	for {
		in, err := c.stream.Recv()
		if err == io.EOF {
			log.Printf("[readpump] recv got EOF: %v", err)
			return
		} else if err != nil {
			log.Printf("[readpump] failed to recv: %v", err)
			return
		}

		switch in.Event.(type) {
		case *protocol.Message_Login_:
			c.ui.PushMessage(fmt.Sprintf("[SYSTEM] New Gopher(%s)", in.GetLogin().Name))
		case *protocol.Message_Logout_:
			c.ui.PushMessage(fmt.Sprintf("[STSTEM] Bye Gopher(%s)", in.GetLogout().Name))
		case *protocol.Message_Message_:
			c.ui.PushMessage(fmt.Sprintf("[%s] %s", in.GetMessage().Name, in.GetMessage().Message))
		case *protocol.Message_Shutdown_:
			c.ui.PushMessage("[SYSTEM] Sever shutdown")
			return
		}

	}
}

func (c *ChatClient) Login() {
	if err := c.connect(); err != nil {
		log.Printf("[chat] failed to connect: %v", err)
		c.ui.PushMessage(fmt.Sprintf("[chat] failed to connect: %v", err))
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	res, err := c.rpc.Login(ctx, &protocol.LoginRequest{
		Name: "test",
	})
	if err != nil {
		log.Printf("[chat] failed to send message: %v", err)
		return
	}
	c.Id = res.Token

	c.Subscribe(context.Background())

	c.ui.PushMessage(fmt.Sprintf("[Login] token: %s", res.Token))
	c.ui.Eval(`window.app.connected = true`)
}

func (c *ChatClient) Logout() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	_, err := c.rpc.Logout(ctx, &protocol.LogoutRequest{
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
