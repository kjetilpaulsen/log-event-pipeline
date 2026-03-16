package main

import (
	"bufio"
	"encoding/json"
	"fmt"
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
		clients: make(map[chan LogEvent]struct{}),
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
	// close(ch)
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

func handleConnection(conn net.Conn, broker *Broker) {
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
			log.Printf("failed to accept tcp connection: %v", err)
			continue
		}

		go handleConnection(conn, broker)
	}
}

func eventsHandler(broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		clientChan := make(chan LogEvent, 16)
		broker.AddClient(clientChan)
		defer broker.RemoveClient(clientChan)

		log.Printf("sse client connected: %s", r.RemoteAddr)

		ctx := r.Context()
		for {
			select {
			case <- ctx.Done():
				log.Printf("sse client disconnected: %s", r.RemoteAddr)
				return

			case event := <- clientChan:
				data, err := json.Marshal(event)
				if err != nil {
					log.Printf("failed to marshal sse event: %v", err)
					continue
				}
				_, err = fmt.Fprintf(w, "data: %s\n\n", data)
				if err != nil {
					log.Printf("failed to write sse event: %v", err)
					return
				}

				flusher.Flush()
			}
		}
	}
}

// func startHTTPServer(adress string, broker *Broker) error {
// 	mux := http.NewServeMux()
//
// 	mux.HandleFunc("/events", eventsHandler(broker))
// 	fileServer := http.FileServer(http.Dir("../frontend"))
// 	mux.Handle("/", fileServer)
// 	// mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 	// 	http.ServeFile(w, r, "../frontend/index.html")
// 	// }) 
// 	//
// 	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
// 		w.Write([]byte("ok"))
// 	})
// 	log.Printf("http server listening on %s", adress)
// 	return http.ListenAndServe(adress, mux)
// }


func startHTTPServer(address string, broker *Broker) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/events", eventsHandler(broker))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "../frontend/index.html")
	})

	mux.HandleFunc("/main.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		http.ServeFile(w, r, "../frontend/main.js")
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	log.Printf("http server listening on %s", address)
	return http.ListenAndServe(address, mux)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	broker := NewBroker()
	adressIn := "127.0.0.1:9000"
	adressOut := "127.0.0.1:9001"
	
	go func() {
		if err := startTCPServer(adressIn, broker); err != nil {
			log.Printf("tcp server error: %v", err)
			os.Exit(1)
		}
	}()

	if err := startHTTPServer(adressOut, broker); err != nil {
		log.Printf("http server error: %v", err)
		os.Exit(1)
	}
}

