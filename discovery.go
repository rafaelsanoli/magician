package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// Função para iniciar a descoberta automática de peers na rede local
func startDiscovery(port string) {
	// Cria um socket UDP para broadcast
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 0,
	})
	if err != nil {
		log.Printf("Erro ao criar socket de descoberta: %v", err)
		updateChatView("Sistema: Erro ao iniciar descoberta automática")
		return
	}
	defer conn.Close()

	// Configura para permitir broadcast
	conn.SetReadBuffer(1024)

	// Inicia a escuta por broadcasts de peers
	go listenForDiscovery(port)

	// Anuncia nossa presença periodicamente
	broadcastAddr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:"+port)
	if err != nil {
		log.Printf("Erro ao resolver endereço de broadcast: %v", err)
		return
	}

	logMessage("Descoberta automática iniciada na porta " + port)
	updateChatView("Sistema: Descoberta automática iniciada")

	// A cada 5 segundos, envia um anúncio
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		message := fmt.Sprintf("MAGICIAN_DISCOVERY_%s", port)
		_, err := conn.WriteToUDP([]byte(message), broadcastAddr)
		if err != nil {
			log.Printf("Erro ao enviar broadcast: %v", err)
		}
	}
}

// Escuta por anúncios de outros peers na rede
func listenForDiscovery(port string) {
	addr := net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 9999, // Porta dedicada para descoberta
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Printf("Erro ao escutar anúncios: %v", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Erro ao ler UDP: %v", err)
			continue
		}

		message := string(buffer[:n])
		if strings.HasPrefix(message, "MAGICIAN_DISCOVERY_") {
			peerPort := strings.TrimPrefix(message, "MAGICIAN_DISCOVERY_")
			peerAddr := fmt.Sprintf("%s:%s", remoteAddr.IP.String(), peerPort)

			// Não conecta a si mesmo
			myIPs, err := getLocalIPs()
			if err != nil {
				log.Printf("Erro ao obter IPs locais: %v", err)
				continue
			}

			isSelf := false
			for _, ip := range myIPs {
				if remoteAddr.IP.String() == ip && peerPort == port {
					isSelf = true
					break
				}
			}

			if !isSelf {
				// Verifica se já estamos conectados a este peer
				if _, exists := Peers[peerAddr]; !exists {
					log.Printf("Descoberto novo peer: %s", peerAddr)
					updateChatView(fmt.Sprintf("Sistema: Descoberto novo peer: %s", peerAddr))
					go connectToPeer(peerAddr)
				}
			}
		}
	}
}

// Obtém todos os IPs locais da máquina
func getLocalIPs() ([]string, error) {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	return ips, nil
}
