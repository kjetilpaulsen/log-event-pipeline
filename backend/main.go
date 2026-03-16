package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
)

type LogEvent struct {
	Timestamp string `json:"timestamp"`
	Level string `json:"level"`
	AppName string `json:"app_name"`
	LoggerName string `json:"logger_name"`
	FunctionName string `json:"function_name"`
	Message string `json:"message"`
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("client connected: %s", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := scanner.Bytes()

		var event LogEvent
		if err := json.Unmarshal(line, &event); err != nil {
			log.Printf("invalid json from %s %v", conn.RemoteAddr(), err)
			continue
		}
		log.Printf(
			"event recieved | ts=%s level=%s app=%s logger=%s func=%s msg=%s",
			event.Timestamp,
			event.Level,
			event.AppName,
			event.LoggerName,
			event.FunctionName,
			event.Message,
		)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("connection read error from %s: %v", conn.RemoteAddr(), err)
	}

	log.Printf("client disconnected: %s", conn.RemoteAddr())
}

func main() {
	adress := "127.0.0.1:9000"

	listener, err := net.Listen("tcp", adress)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", adress, err)
	}
	defer listener.Close()

	log.Printf("listening on %s", adress)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}
