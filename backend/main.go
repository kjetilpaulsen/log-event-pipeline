package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
)

type LogEvent struct {
	Timestamp string `json: "timestamp"`
}
