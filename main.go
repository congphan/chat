package main

import (
	"context"
	"log"
	"net"
	"os"
	"sync"

	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"

	"github.com/congphan/chat/schema"
)

var grpcLog glog.LoggerV2

func init() {
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

type Connection struct {
	stream schema.Broadcast_CreateStreamServer
	id     string //user_id
	active bool
	error  chan error
}

type Server struct {
	Connections []*Connection
}

func (s *Server) CreateStream(c *schema.Connect, stream schema.Broadcast_CreateStreamServer) error {
	conn := &Connection{
		stream,
		c.GetUser().GetId(),
		true,
		make(chan error),
	}

	s.Connections = append(s.Connections, conn)
	return <-conn.error
}

func (s *Server) BroadcastMessage(ctx context.Context, msg *schema.Message) (*schema.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)

	for _, conn := range s.Connections {
		wait.Add(1)

		go func(msg *schema.Message, conn *Connection) {
			defer wait.Done()

			if conn.active {
				grpcLog.Info("Sending messge to : ", conn.stream)

				err := conn.stream.Send(msg)
				if err != nil {
					grpcLog.Errorf("Error with Stream: %v - Error: %v", conn.stream, err)
					conn.active = false
					conn.error <- err
				}
			}
		}(msg, conn)
	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
	return &schema.Close{}, nil
}

func main() {
	var connections []*Connection
	server := &Server{connections}

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("error creating the server %v", err)
	}

	grpcLog.Info("Starting server at port :50051")

	schema.RegisterBroadcastServer(grpcServer, server)
	grpcServer.Serve(listener)
}
