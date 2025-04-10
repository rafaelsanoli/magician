package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
)

var chatView *gocui.View

// initUI inicializa a interface do usuário baseada em terminal usando gocui
func initUI() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Fatalf("Falha ao iniciar interface: %v", err)
	}
	defer g.Close()
	G = g

	g.Cursor = true
	g.Mouse = true
	g.SetManagerFunc(layout)

	// Configurando atalhos de teclado
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Fatalf("Erro ao configurar tecla: %v", err)
	}

	if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, sendMessage); err != nil {
		log.Fatalf("Erro ao configurar tecla: %v", err)
	}

	// Foca na área de input
	g.SetCurrentView("input")

	// Exibe mensagem de boas-vindas
	updateChatView(fmt.Sprintf("--- Bem-vindo ao Magician Chat, %s! ---", Nickname))
	updateChatView("Use /ajuda para ver os comandos disponíveis.")

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatalf("Erro na interface: %v", err)
	}
}

// layout define o layout da interface
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// View do chat (ocupa maior parte da tela)
	if v, err := g.SetView("chat", 0, 0, maxX-1, maxY-4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "📤 Magician Chat"
		v.Wrap = true
		v.Autoscroll = true
		v.SetCursor(0, 0)
		chatView = v
	}

	// View do input (parte inferior)
	if v, err := g.SetView("input", 0, maxY-3, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "📝 Mensagem"
		v.Editable = true
		v.Wrap = true
		if _, err := g.SetCurrentView("input"); err != nil {
			return err
		}
	}

	return nil
}

// updateChatView atualiza a view do chat com uma nova mensagem
func updateChatView(message string) {
	if G == nil || chatView == nil {
		return
	}

	G.Update(func(g *gocui.Gui) error {
		timestamp := time.Now().Format("15:04:05")
		fmt.Fprintf(chatView, "[%s] %s\n", timestamp, message)
		return nil
	})
}

// processMessage processa uma mensagem para verificar se é um comando
func processMessage(message string) (bool, string) {
	if !strings.HasPrefix(message, "/") {
		return false, ""
	}

	parts := strings.Fields(message)
	command := parts[0]
	args := parts[1:]

	switch command {
	case "/ajuda", "/help":
		return true, `
📋 Comandos disponíveis:
/ajuda              - Mostra esta ajuda
/usuarios           - Lista os peers conectados
/privado <peer> <msg> - Envia mensagem privada
/limpar             - Limpa a tela
/logs [n]           - Mostra últimas n mensagens do log
/arquivo <path> [peer] - Envia arquivo
/info               - Mostra as informações da Rede Tor
/sair               - Fecha o programa
`
	case "/usuarios", "/users":
		return true, cmdListUsers(args)
	case "/privado", "/private":
		return true, cmdPrivateMsg(args)
	case "/limpar", "/clear":
		return true, "[LIMPAR]"
	case "/logs":
		return true, cmdShowLogs(args)
	case "/arquivo", "/file":
		return true, cmdSendFile(args)
	case "/info":
		return true, cmdInfo(args)
	case "/sair", "/exit":
		return true, "Saindo..."
	default:
		return true, fmt.Sprintf("Comando desconhecido: %s. Use /ajuda.", command)
	}
}

// cmdShowLogs mostra as últimas mensagens do log
func cmdShowLogs(args []string) string {
	var lines int = 10 // Padrão: 10 linhas
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &lines)
	}
	return getLastLogEntries(lines)
}

// cmdSendFile inicia a transferência de um arquivo
func cmdSendFile(args []string) string {
	if len(args) < 1 {
		return "Uso: /arquivo <caminho_do_arquivo> [peer_destino]"
	}

	filePath := args[0]
	targetPeer := ""
	if len(args) > 1 {
		targetPeer = args[1]
	}

	go func() {
		if err := sendFile(filePath, targetPeer); err != nil {
			updateChatView(fmt.Sprintf("❌ Erro ao enviar arquivo: %v", err))
		}
	}()

	return fmt.Sprintf("Iniciando envio do arquivo: %s", filePath)
}

// quit encerra a aplicação
func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
