package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/jroimartin/gocui"
)

// Variáveis globais compartilhadas
var Nickname, Password string
var Peers = make(map[string]net.Conn)
var G *gocui.Gui

func initLogSystem() error {
	os.MkdirAll("logs", 0755)
	return nil
}

func initFileTransferSystem() error {
	os.MkdirAll("recebidos", 0755)
	return nil
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Seu nome: ")
	Nickname, _ = reader.ReadString('\n')
	Nickname = strings.TrimSpace(Nickname)

	fmt.Print("Porta para escutar (ex: 9000): ")
	port, _ := reader.ReadString('\n')
	port = strings.TrimSpace(port)

	fmt.Print("Senha (obrigatória): ")
	Password, _ = reader.ReadString('\n')
	Password = strings.TrimSpace(Password)

	fmt.Print("Habilitar descoberta automática de peers? (s/n): ")
	enableDiscovery, _ := reader.ReadString('\n')
	enableDiscovery = strings.TrimSpace(strings.ToLower(enableDiscovery))

	// Inicializa os subsistemas
	if err := initLogSystem(); err != nil {
		log.Fatalf("Erro ao inicializar sistema de logs: %v", err)
	}

	if err := initFileTransferSystem(); err != nil {
		log.Fatalf("Erro ao inicializar sistema de transferência: %v", err)
	}

	// Adiciona entrada inicial ao log
	logMessage(fmt.Sprintf("--- Sessão iniciada por %s na porta %s ---", Nickname, port))

	go listenForPeers(port)

	if enableDiscovery == "s" || enableDiscovery == "sim" {
		go startDiscovery(port)
	}

	fmt.Print("Conectar a peer (IP:porta) ou Enter para pular: ")
	addr, _ := reader.ReadString('\n')
	addr = strings.TrimSpace(addr)

	if addr != "" {
		go connectToPeer(addr)
	}

	// Inicia a interface
	initUI()
}
