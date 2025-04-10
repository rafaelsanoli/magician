package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"net"
	"os"
	"strings"
)

func cmdPrivateMsg(args []string) string {
	if len(args) < 2 {
		return "Uso: /privado <peer> <mensagem>"
	}

	target := args[0]
	message := strings.Join(args[1:], " ")

	// Procura o peer pelo endereço (completo ou parcial)
	var targetConn net.Conn
	var targetAddr string

	peersMutex.Lock()
	for addr, conn := range Peers {
		if strings.Contains(addr, target) {
			targetConn = conn
			targetAddr = addr
			break
		}
	}
	peersMutex.Unlock()

	if targetConn == nil {
		return fmt.Sprintf("Peer '%s' não encontrado.", target)
	}

	// Envia mensagem privada
	fmt.Fprintf(targetConn, "[PRIVADO] %s\n", message)

	// Loga a mensagem privada
	logMessage(fmt.Sprintf("[PRIVADO para %s] %s", targetAddr, message))

	return fmt.Sprintf("Mensagem privada enviada para %s", targetAddr)
}

func cmdListUsers(args []string) string {
	peersMutex.Lock()
	peerCount := len(Peers)

	if peerCount == 0 {
		peersMutex.Unlock()
		return "Nenhum peer conectado."
	}

	result := fmt.Sprintf("🔌 Peers conectados (%d):\n", peerCount)
	i := 1
	for addr := range Peers {
		result += fmt.Sprintf("%d. %s\n", i, addr)
		i++
	}
	peersMutex.Unlock()

	return result
}

func sendMessage(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		message := strings.TrimSpace(v.Buffer())
		if message == "" {
			v.Clear()
			v.SetCursor(0, 0)
			return nil
		}

		// Processa comandos
		isCommand, response := processMessage(message)
		if isCommand {
			if message == "/sair" || message == "/exit" {
				return gocui.ErrQuit
			}
			if response == "[LIMPAR]" {
				g.Update(func(g *gocui.Gui) error {
					chatView.Clear()
					return nil
				})
			} else {
				updateChatView(response)
			}
		} else {
			// Formata a mensagem para envio (IMPORTANTE: sem prefixo AUTH para não confundir)
			formattedMsg := fmt.Sprintf("[%s] %s", Nickname, message)

			// Envia mensagem para todos os peers
			peersMutex.Lock()
			for _, conn := range Peers {
				fmt.Fprintf(conn, "%s\n", formattedMsg)
			}
			peersMutex.Unlock()

			updateChatView(fmt.Sprintf("[Você] %s", message))
		}

		v.Clear()
		v.SetCursor(0, 0)
	}
	return nil
}

func cmdInfo(args []string) string {
	onion, err := os.ReadFile("/var/lib/tor/magician_chat/hostname")
	if err != nil {
		return "Erro ao ler endereço .onion"
	}

	ip := getLocalIP()

	resp := "🔍 Informações do Peer:\n"
	resp += fmt.Sprintf("🧅 Endereço .onion: %s", strings.TrimSpace(string(onion)))
	resp += fmt.Sprintf("\n📡 IP local: %s", ip)
	resp += "\n🛡️  Modo: Tor ativado"

	return resp
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "desconhecido"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
