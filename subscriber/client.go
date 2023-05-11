package subscriber

import (
	"context"
	"io"
	"log"

	"pub_sub/proto"

	"google.golang.org/grpc"
)

type ContentFunc func(msg *proto.Message)
type Client struct {
	Topics      chan *proto.Topic
	ContentFunc ContentFunc
}

func NewClient(contentFunc ContentFunc) *Client {
	return &Client{
		Topics:      make(chan *proto.Topic, 1000),
		ContentFunc: contentFunc,
	}
}

func (c *Client) Subscribe(addr string) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	client := proto.NewSubscribeServiceClient(conn)
	stream, err := client.Subscribe(context.Background())
	if err != nil {
		log.Fatalf("%v", err)
	}

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Println("err:", err)
				break
			}
			if in != nil {
				c.ContentFunc(in)
			}
		}
	}()

	go func() {
		for topic := range c.Topics {
			log.Println(topic)
			stream.Send(topic)
		}
	}()
}
