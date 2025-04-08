package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func logMessage(message string) {
	os.MkdirAll("logs", 0755)
	filename := time.Now().Format("logs/chat-2006-01-02.log")
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Erro ao gravar no log: %v", err)
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("15:04:05")
	fmt.Fprintf(f, "[%s] %s\n", timestamp, message)
}
