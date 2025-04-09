package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var peersMutex sync.Mutex

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
	tlsConfig, err := loadTLSConfig()
	if err != nil {
		updateChatView(fmt.Sprintf("Erro TLS: %v", err))
		return
	}

	tlsConfig.ServerName = "MagicianPeer" // Opcional: depende do CN do certificado do peer

	for {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // ⚠️ aceita certificado autossinado (apenas para dev)
		}
		conn, err := tls.Dial("tcp", address, tlsConfig)

		if err != nil {
			log.Println("Erro ao conectar. Tentando novamente em 5s...")
			updateChatView("Sistema: Falha ao conectar a " + address + ". Tentando novamente em 5s...")
			time.Sleep(5 * time.Second)

			peersMutex.Lock()
			_, exists := Peers[address]
			peersMutex.Unlock()

			if exists {
				return
			}
			continue
		}

		fmt.Fprintf(conn, "%s\n", Password)
		response, _ := bufio.NewReader(conn).ReadString('\n')
		if strings.TrimSpace(response) != "OK" {
			log.Println("Senha incorreta. Conexão rejeitada.")
			updateChatView("Sistema: Senha incorreta para " + address + ". Conexão rejeitada.")
			conn.Close()
			return
		}

		peersMutex.Lock()
		Peers[address] = conn
		peersMutex.Unlock()

		updateChatView("Sistema: Conectado com sucesso a " + address)
		go handleConnection(conn)
		break
	}
}

func handleConnection(conn net.Conn) {
	remote := conn.RemoteAddr().String()

	peersMutex.Lock()
	_, exists := Peers[remote]
	peersMutex.Unlock()

	if !exists {
		reader := bufio.NewReader(conn)
		clientPassword, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Erro ao ler a senha:", err)
			return
		}
		clientPassword = strings.TrimSpace(clientPassword)

		// DEBUG: Mostrar senhas recebi	da e esperada
		//fmt.Printf(">> Senha recebida: [%s]\n", strings.TrimSpace(clientPassword))
		//fmt.Printf(">> Senha esperada: [%s]\n", Password)
		if strings.TrimSpace(clientPassword) != Password {
			fmt.Fprintln(conn, "DENIED")
			fmt.Printf(">>> Senha inválida de %s: [%s] (esperada: [%s])\n", conn.RemoteAddr(), clientPassword, Password)
			conn.Close()
			return
		}
		fmt.Fprintln(conn, "OK")

		peersMutex.Lock()
		Peers[remote] = conn
		peersMutex.Unlock()

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

			peersMutex.Lock()
			delete(Peers, remote)
			peersMutex.Unlock()

			conn.Close()
			return
		}

		if strings.HasPrefix(message, "[FILE_TRANSFER]") {
			data := strings.TrimPrefix(message, "[FILE_TRANSFER]")
			handleFileChunk(strings.TrimSpace(data))
			continue
		}

		if strings.HasPrefix(message, "[PRIVADO]") {
			privateMsg := strings.TrimPrefix(message, "[PRIVADO]")
			updateChatView(fmt.Sprintf("🔒 [Mensagem privada de %s] %s", remote, privateMsg))
			logMessage(fmt.Sprintf("[PRIVADO de %s] %s", remote, privateMsg))
			continue
		}

		updateChatView("[" + remote + "] " + message)
		logMessage(fmt.Sprintf("[%s] %s", remote, strings.TrimSpace(message)))
	}
}
