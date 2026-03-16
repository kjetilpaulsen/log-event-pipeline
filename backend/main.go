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

type Broker struct {
	mu sync.Mutex
	clients map[chan LogEvent]struct{}
}

func NewBroker() *Broker {
	return &Broker{
		clients: make(map[chan LogEvent]struct{})
	}
}

func (b *Broker) AddClient(ch chan LogEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[ch] = struct{}{}
}

func (b *Broker) RemoveClient (ch chan LogEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.clients, ch)
	close(ch)
}

func (b *Broker) Broadcast(event LogEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for ch := range b.clients {
		select {
		case ch <- event:
		default:
		}
	}
}

func isHighSeverity(level string) bool {
	upper := strings.ToUpper(level)
	return upper == "ERROR" || upper == "CRITICAL"
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

		if isHighSeverity(event.Level) {
			broker.Broadcast(event)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("connection read error from %s: %v", conn.RemoteAddr(), err)
	}

	log.Printf("client disconnected: %s", conn.RemoteAddr())
}

func startTCPServer(adress string, broker *Broker) error {
	listener, err := net.Listen("tcp", adress)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", adress, err)
	}
	defer listener.Close()

	log.Printf("tcp server listening on %s", adress)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Prinf("failed to accept tcp connection: %v", err)
			continue
		}

		go handleConnection(conn, broker)
	}
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
