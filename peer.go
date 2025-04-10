package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"magician/tor"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var peersMutex sync.Mutex

// Mapa para controlar quais peers já passaram pela autenticação
var peerAuthenticated = make(map[string]bool)

func loadTLSConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		return nil, err
	}

	caCert, err := os.ReadFile("cert.pem")
	if err != nil {
		return nil, fmt.Errorf("erro ao ler certificado: %v", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("erro ao adicionar CA ao pool")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.NoClientCert,
		ClientCAs:    caPool,
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS12,
	}

	return tlsConfig, nil
}

func listenForPeers(port string) {
	tlsConfig, err := loadTLSConfig()
	if err != nil {
		log.Fatal("Erro ao configurar TLS:", err)
	}

	ln, err := tls.Listen("tcp", ":"+port, tlsConfig)
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
	peersMutex.Lock()
	_, exists := Peers[address]
	peersMutex.Unlock()

	if exists {
		log.Println("Já conectado a", address)
		return
	}

	updateChatView("Sistema: Tentando conectar a " + address)

	// Carrega configuração segura
	//tlsConfig, err := loadTLSConfig()
	//if err != nil {
	//	updateChatView(fmt.Sprintf("Erro TLS: %v", err))
	//	return
	//}

	// Usar corretamente o tor.DialOrDirect e salvar o resultado
	rawConn, err := tor.DialOrDirect(address)
	if err != nil {
		log.Printf("Erro ao conectar (via Tor): %v", err)
		return
	}

	// Configuração TLS para cliente
	insecureTlsConfig := &tls.Config{
		InsecureSkipVerify: true, // ⚠️ aceita certificado autossinado (apenas para dev)
	}

	// Configurar conexão TLS
	conn := tls.Client(rawConn, insecureTlsConfig)

	// Verificação de erro após conectar
	if err := conn.Handshake(); err != nil {
		log.Println("Erro ao conectar. Tentando novamente em 5s...")
		updateChatView("Sistema: Falha ao conectar a " + address + ". Tentando novamente em 5s...")
		conn.Close()
		time.Sleep(5 * time.Second)
		return
	}

	// AQUI É A PARTE CRÍTICA: Envia a senha como um comando especial para identificação
	fmt.Fprintf(conn, "AUTH %s\n", Password)

	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Erro ao ler resposta de autenticação: %v", err)
		conn.Close()
		return
	}

	if strings.TrimSpace(response) != "OK" {
		log.Println("Senha incorreta. Conexão rejeitada.")
		updateChatView("Sistema: Senha incorreta para " + address + ". Conexão rejeitada.")
		conn.Close()
		return
	}

	peersMutex.Lock()
	Peers[address] = conn
	peerAuthenticated[address] = true // Marcar como autenticado
	peersMutex.Unlock()

	updateChatView("Sistema: Conectado com sucesso a " + address)

	// Depois da autenticação, continuar com a rotina normal de tratamento
	go handlePeerMessages(conn)
}

func handleConnection(conn net.Conn) {
	remote := conn.RemoteAddr().String()

	reader := bufio.NewReader(conn)
	authenticated := false

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Peer desconectado:", remote)
			updateChatView("Sistema: Peer desconectado: " + remote)
			logMessage("Peer desconectado: " + remote)

			peersMutex.Lock()
			delete(Peers, remote)
			delete(peerAuthenticated, remote)
			peersMutex.Unlock()

			conn.Close()
			return
		}

		trimmedMsg := strings.TrimSpace(message)

		// Verificar se é um comando de autenticação
		if !authenticated && strings.HasPrefix(trimmedMsg, "AUTH ") {
			// Extrair a senha da mensagem de autenticação
			receivedPassword := strings.TrimPrefix(trimmedMsg, "AUTH ")
			expectedPassword := Password

			//fmt.Printf("DEBUG - Senha recebida: [%s] (len: %d)\n", receivedPassword, len(receivedPassword))
			//fmt.Printf("DEBUG - Senha esperada: [%s] (len: %d)\n", expectedPassword, len(expectedPassword))

			if receivedPassword != expectedPassword {
				fmt.Fprintln(conn, "DENIED")
				fmt.Printf(">>> Senha inválida de %s: [%s] (esperada: [%s])\n",
					conn.RemoteAddr(), receivedPassword, expectedPassword)
				conn.Close()
				return
			}

			// Senha correta
			fmt.Fprintln(conn, "OK")
			authenticated = true

			peersMutex.Lock()
			Peers[remote] = conn
			peerAuthenticated[remote] = true
			peersMutex.Unlock()

			updateChatView("Sistema: Novo peer conectado de " + remote)
			logMessage("Novo peer conectado: " + remote)
			continue
		}

		// Se não estiver autenticado e não for um comando AUTH, rejeita
		if !authenticated {
			fmt.Fprintln(conn, "DENIED")
			fmt.Printf(">>> Tentativa de comunicação sem autenticação: %s\n", remote)
			conn.Close()
			return
		}

		// Processamento normal de mensagens após autenticação
		if strings.HasPrefix(trimmedMsg, "[FILE_TRANSFER]") {
			data := strings.TrimPrefix(trimmedMsg, "[FILE_TRANSFER]")
			handleFileChunk(strings.TrimSpace(data))
		} else if strings.HasPrefix(trimmedMsg, "[PRIVADO]") {
			privateMsg := strings.TrimPrefix(trimmedMsg, "[PRIVADO]")
			updateChatView(fmt.Sprintf("🔒 [Mensagem privada de %s] %s", remote, privateMsg))
			logMessage(fmt.Sprintf("[PRIVADO de %s] %s", remote, privateMsg))
		} else {
			// Mensagem normal
			updateChatView(fmt.Sprintf("[%s] %s", remote, trimmedMsg))
			logMessage(fmt.Sprintf("[%s] %s", remote, trimmedMsg))
		}
	}
}

// Nova função para lidar com as mensagens de peers já autenticados
func handlePeerMessages(conn net.Conn) {
	remote := conn.RemoteAddr().String()
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Peer desconectado:", remote)
			updateChatView("Sistema: Peer desconectado: " + remote)
			logMessage("Peer desconectado: " + remote)

			peersMutex.Lock()
			delete(Peers, remote)
			delete(peerAuthenticated, remote)
			peersMutex.Unlock()

			conn.Close()
			return
		}

		trimmedMsg := strings.TrimSpace(message)

		// Processamento normal de mensagens
		if strings.HasPrefix(trimmedMsg, "[FILE_TRANSFER]") {
			data := strings.TrimPrefix(trimmedMsg, "[FILE_TRANSFER]")
			handleFileChunk(strings.TrimSpace(data))
		} else if strings.HasPrefix(trimmedMsg, "[PRIVADO]") {
			privateMsg := strings.TrimPrefix(trimmedMsg, "[PRIVADO]")
			updateChatView(fmt.Sprintf("🔒 [Mensagem privada de %s] %s", remote, privateMsg))
			logMessage(fmt.Sprintf("[PRIVADO de %s] %s", remote, privateMsg))
		} else {
			// Mensagem normal
			updateChatView(trimmedMsg)
			logMessage(fmt.Sprintf("[%s] %s", remote, trimmedMsg))
		}
	}
}
