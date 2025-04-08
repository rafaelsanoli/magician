package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

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
	broadcastAddr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:9999") // Porta fixa para discovery
	if err != nil {
		log.Printf("Erro ao resolver endereço de broadcast: %v", err)
		updateChatView("Sistema: Erro ao configurar descoberta - " + err.Error())
		return
	}

	logMessage("Descoberta automática iniciada na porta " + port)
	updateChatView("Sistema: Descoberta automática iniciada")

	// A cada 5 segundos, envia um anúncio
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop() // Importante: evita vazamento de goroutines

	for range ticker.C {
		message := fmt.Sprintf("MAGICIAN_DISCOVERY_%s", port)
		_, err := conn.WriteToUDP([]byte(message), broadcastAddr)
		if err != nil {
			log.Printf("Erro ao enviar broadcast: %v", err)
		}
	}
}

// Melhoria na função listenForDiscovery para tratar melhor erros de conectividade
func listenForDiscovery(port string) {
	addr := net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 9999, // Porta dedicada para descoberta
	}

	// Tentativa com tratamento de erro aprimorado
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		// Tenta portas alternativas se a padrão estiver em uso
		for altPort := 10000; altPort < 10010; altPort++ {
			addr.Port = altPort
			conn, err = net.ListenUDP("udp", &addr)
			if err == nil {
				log.Printf("Usando porta alternativa para descoberta: %d", altPort)
				updateChatView(fmt.Sprintf("Sistema: Usando porta alternativa para descoberta: %d", altPort))
				break
			}
		}

		if err != nil {
			log.Printf("Erro ao escutar anúncios: %v", err)
			updateChatView("Sistema: Falha ao iniciar serviço de descoberta")
			return
		}
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
				peersMutex.Lock()
				_, exists := Peers[peerAddr]
				peersMutex.Unlock()

				if !exists {
					log.Printf("Descoberto novo peer: %s", peerAddr)
					updateChatView(fmt.Sprintf("Sistema: Descoberto novo peer: %s", peerAddr))
					go connectToPeer(peerAddr)
				}
			}
		}
	}
}

func getLocalIPs() ([]string, error) {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && !ip.IsLoopback() && ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
	}
	return ips, nil
}
