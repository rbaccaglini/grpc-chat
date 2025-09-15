package main

import (
	"bufio"
	"context"
	"fmt"
	"grpc-chat/chat"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := chat.NewChatServiceClient(conn)
	stream, err := client.Join(context.Background())
	if err != nil {
		log.Fatalf("Failed to join chat: %v", err)
	}

	fmt.Print("Enter your name: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	user := scanner.Text()
	go func() {
		for {
			var msg chat.Message
			if err := stream.RecvMsg(&msg); err != nil {
				log.Fatalf("Failed to receive message: %v", err)
			}
			fmt.Printf("[%s] %s: %s\n", time.Unix(msg.Timestamp, 0).Format("15:04:05"), msg.User, msg.Text)
		}
	}()

	for scanner.Scan() {
		msg := &chat.Message{
			User:      user,
			Text:      scanner.Text(),
			Timestamp: time.Now().Unix(),
		}
		if err := stream.Send(msg); err != nil {
			log.Fatalf("Failed to send message: %v", err)
		}
	}
}
