package main

import (
	"bufio"
	"context"
	"flag"
	"github.com/golang/protobuf/ptypes"
	"github.com/riimi/tutorial-grpc-chat/clean/domain"
	"github.com/riimi/tutorial-grpc-chat/clean/interface/rpc/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
	"os"
	"time"
)

type Client struct {
	protocol.ChatServiceClient
	addr      string
	user      domain.User
	connected bool

	waitc      chan struct{}
	send       chan string
	LogHandler func(string, ...interface{})
}

func NewClient(addr, name string) *Client {
	return &Client{
		addr: addr,
		user: domain.User{
			Name: name,
		},
		waitc:      make(chan struct{}),
		send:       make(chan string, 100),
		LogHandler: func(string, ...interface{}) {},
	}
}

func (c *Client) Run(ctx context.Context) error {
	connCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	conn, err := grpc.DialContext(connCtx, c.addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()
	c.ChatServiceClient = protocol.NewChatServiceClient(conn)

	resLogin, err := c.Login(ctx, &protocol.LoginRequest{
		Name: c.user.Name,
	})
	if err != nil {
		return err
	}
	c.user.Token = resLogin.Token
	c.LogHandler("login")

	startTime := time.Now()
	if err = c.Chat(ctx); err != nil {
		return err
	}
	c.LogHandler("chat exit: %f", time.Since(startTime).Seconds())

	if _, err := c.Logout(ctx, &protocol.LogoutRequest{
		Token: c.user.Token,
	}); err != nil {
		return err
	}
	c.LogHandler("logout")
	return nil
}

func (c *Client) Chat(ctx context.Context) error {
	md := metadata.New(map[string]string{"myheader-token": c.user.Token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stream, err := c.Subscribe(ctx)
	if err != nil {
		return err
	}
	c.LogHandler("subscribe")

	go c.readPump(stream)
	c.writePump(stream)
	<-c.waitc
	return nil
}

func (c *Client) readPump(stream protocol.ChatService_SubscribeClient) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			c.LogHandler("[readpump] recv got EOF: %v", err)
			close(c.waitc)
			return nil
		} else if err != nil {
			c.LogHandler("[readpump] failed to recv: %v", err)
			return err
		}

		switch in.Event.(type) {
		case *protocol.Message_Login_:
			//c.LogHandler("[SYSTEM] New Gopher(%s)", in.GetLogin().Name)
		case *protocol.Message_Logout_:
			//c.LogHandler("[STSTEM] Bye Gopher(%s)", in.GetLogout().Name)
		case *protocol.Message_Message_:
			//c.LogHandler("[%s] %s", in.GetMessage().Name, in.GetMessage().Message)
		case *protocol.Message_Shutdown_:
			c.LogHandler("[SYSTEM] Sever shutdown")
			return nil
		}

	}
}

func (c *Client) writePump(stream protocol.ChatService_SubscribeClient) {
	defer stream.CloseSend()
	for {
		select {
		case <-stream.Context().Done():
			c.LogHandler("write end")
		case text, more := <-c.send:
			if !more {
				c.LogHandler("quit")
				return
			}
			msg := &protocol.Message{
				Timestamp: ptypes.TimestampNow(),
				Event: &protocol.Message_Message_{
					Message: &protocol.Message_Message{
						Name:    c.user.Name,
						Message: text,
					},
				},
			}
			if err := stream.Send(msg); err != nil {
				c.LogHandler("write error: %v", err)
				return
			}
		}
	}
}

func main() {
	serverAddr := flag.String("addr", "localhost:40040", "grpc server address")
	name := flag.String("name", "riimi", "name")
	//conn := flag.Int("conn", 50, "max connection")
	flag.Parse()

	done := make(chan struct{})
	gophers := NewClient(*serverAddr, *name)
	go func() {
		gophers.Run(context.Background())
		done <- struct{}{}
	}()
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		text := sc.Text()
		if text == "!q" {
			close(gophers.send)
			break
		}
		gophers.send <- text
	}
	<-done
}
