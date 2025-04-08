package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func listenForPeers(port string) {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Fatal("Erro ao carregar TLS:", err)
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	ln, err := tls.Listen("tcp", ":"+port, config)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	log.Println("Escutando em :", port)
	updateChatView("Sistema: Escutando na porta " + port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Erro na conexão:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func connectToPeer(address string) {
	// Evita conexões redundantes
	if _, exists := Peers[address]; exists {
		log.Println("Já conectado a", address)
		return
	}

	updateChatView("Sistema: Tentando conectar a " + address)

	for {
		conn, err := tls.Dial("tcp", address, &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			log.Println("Erro ao conectar. Tentando novamente em 5s...")
			updateChatView("Sistema: Falha ao conectar a " + address + ". Tentando novamente em 5s...")
			time.Sleep(5 * time.Second)

			// Verifica se já nos conectamos enquanto esperávamos
			if _, exists := Peers[address]; exists {
				return
			}
			continue
		}

		fmt.Fprintf(conn, Password+"\n")
		response, _ := bufio.NewReader(conn).ReadString('\n')
		if strings.TrimSpace(response) != "OK" {
			log.Println("Senha incorreta. Conexão rejeitada.")
			updateChatView("Sistema: Senha incorreta para " + address + ". Conexão rejeitada.")
			conn.Close()
			return
		}

		Peers[address] = conn
		updateChatView("Sistema: Conectado com sucesso a " + address)
		go handleConnection(conn)
		break
	}
}

func handleConnection(conn net.Conn) {
	remote := conn.RemoteAddr().String()

	if _, exists := Peers[remote]; !exists {
		reader := bufio.NewReader(conn)
		clientPassword, _ := reader.ReadString('\n')
		if strings.TrimSpace(clientPassword) != Password {
			fmt.Fprintln(conn, "DENIED")
			updateChatView("Sistema: Conexão rejeitada de " + remote + " (senha incorreta)")
			conn.Close()
			return
		}
		fmt.Fprintln(conn, "OK")
		Peers[remote] = conn
		updateChatView("Sistema: Novo peer conectado de " + remote)
		logMessage("Novo peer conectado: " + remote)
	}

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Peer desconectado:", remote)
			updateChatView("Sistema: Peer desconectado: " + remote)
			logMessage("Peer desconectado: " + remote)
			delete(Peers, remote)
			conn.Close()
			return
		}

		// Verifica se é uma transferência de arquivo
		if strings.HasPrefix(message, "[FILE_TRANSFER]") {
			data := strings.TrimPrefix(message, "[FILE_TRANSFER]")
			handleFileChunk(strings.TrimSpace(data))
			continue
		}

		// Verifica se é uma mensagem privada
		if strings.HasPrefix(message, "[PRIVADO]") {
			privateMsg := strings.TrimPrefix(message, "[PRIVADO]")
			updateChatView(fmt.Sprintf("🔒 [Mensagem privada de %s] %s", remote, privateMsg))
			logMessage(fmt.Sprintf("[PRIVADO de %s] %s", remote, privateMsg))
			continue
		}

		// Mensagem normal
		updateChatView("[" + remote + "] " + message)
		logMessage(fmt.Sprintf("[%s] %s", remote, strings.TrimSpace(message)))
	}
}
