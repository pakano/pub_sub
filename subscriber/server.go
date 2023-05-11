package subscriber

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	"pub_sub/proto"

	"google.golang.org/grpc"
)

//topic subtopic

// servers
type Server struct {
	proto.UnimplementedSubscribeServiceServer
	publishers map[string]*Publisher
}

func NewServer(publishers ...*Publisher) *Server {
	server := &Server{
		publishers: make(map[string]*Publisher),
	}
	for _, publisher := range publishers {
		server.publishers[publisher.topic] = publisher
	}
	return server
}

func (s *Server) Run(addr string) error {
	grpcServer := grpc.NewServer()
	proto.RegisterSubscribeServiceServer(grpcServer, s)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("listen err:", err)
	} else {
		fmt.Println("service start.")
	}
	return grpcServer.Serve(listener)
}

func (s *Server) Subscribe(stream proto.SubscribeService_SubscribeServer) error {
	topicsubs := make(map[string]Subscriber)
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			log.Println("err:", err)
			break
		}
		if err != nil {
			log.Println("err:", err)
			break
		}
		topic := in.GetTopic()
		log.Println(topic)
		if publisher, ok := s.publishers[topic]; ok {
			suber, ok := topicsubs[topic]
			if !ok {
				suber = publisher.Subscribe()
				topicsubs[topic] = suber
			} else {
				log.Println("info:", "topic has been found")
			}
			publisher.Update(suber, in.GetRule())
			go func() {
				topic := topic
				for v := range suber {
					var data string
					if data, ok = v.(string); !ok {
						d, err := json.Marshal(v)
						if err != nil {
							log.Println("err:", err)
							continue
						}
						data = string(d)
					}
					log.Print(v)
					if err := stream.Send(&proto.Message{Topic: topic, Data: string(data)}); err != nil {
						log.Println("err:", err)
						break
					}
				}
			}()
		} else {
			log.Println("err:", "invalid topic", topic)
		}
	}
	for topic, subscriber := range topicsubs {
		if publisher, ok := s.publishers[topic]; ok {
			publisher.Evict(subscriber)
		}
	}
	return nil
}
