package main

import (
	"fmt"
	"grpc-chat/chat"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type chatServer struct {
	chat.UnimplementedChatServiceServer
	mu      sync.Mutex
	clients map[chat.ChatService_JoinServer]bool
	message chan *chat.Message
}

func newServer() *chatServer {
	return &chatServer{
		clients: make(map[chat.ChatService_JoinServer]bool),
		message: make(chan *chat.Message),
	}
}

func (s *chatServer) Join(stream chat.ChatService_JoinServer) error {
	s.mu.Lock()
	s.clients[stream] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, stream)
		s.mu.Unlock()
	}()

	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				return
			}
			s.message <- msg
		}
	}()

	// Broadcast messages to all clients
	for msg := range s.message {
		s.mu.Lock()
		for client := range s.clients {
			if err := client.Send(msg); err != nil {
				fmt.Printf("Error sending message: %v", err)
			}
		}
		s.mu.Unlock()
	}

	return nil
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	grpcServer := grpc.NewServer()
	chat.RegisterChatServiceServer(grpcServer, newServer())

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to initialize gRPC server: %v", err)
	}
}
