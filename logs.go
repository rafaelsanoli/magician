package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func getLastLogEntries(lines int) string {
	file := time.Now().Format("logs/chat-2006-01-02.log")
	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Sprintf("Erro ao ler log: %v", err)
	}

	entries := strings.Split(string(content), "\n")
	start := len(entries) - lines
	if start < 0 {
		start = 0
	}
	return strings.Join(entries[start:], "\n")
}
