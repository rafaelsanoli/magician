package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"log"
	"strings"
	"sync"
	"time"
)

var uiMutex sync.Mutex

func initUI() {
	var err error
	g, err = gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	// Mensagem de boas-vindas
	updateChatView("🧙 Bem-vindo ao Magician-Chat!")
	updateChatView("Digite /ajuda para ver os comandos disponíveis")

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("chat", 0, 0, maxX-1, maxY-3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "🧙 Magician-Chat"
		v.Wrap = true
		v.Autoscroll = true
	}
	if v, err := g.SetView("input", 0, maxY-3, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "🪄 " + nickname
		v.Editable = true
		g.SetCurrentView("input")
	}
	return nil
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, sendMessage(g)); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	return nil
}

func sendMessage(g *gocui.Gui) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		msg := strings.TrimSpace(v.Buffer())
		if msg == "" {
			return nil
		}

		// Limpa o buffer de entrada
		v.Clear()
		v.SetCursor(0, 0)

		// Verifica se é um comando
		if strings.HasPrefix(msg, "/") {
			isCommand, response := processMessage(msg)
			if isCommand {
				// Verifica se é para limpar a tela
				if response == "[LIMPAR]" {
					chatView, _ := g.View("chat")
					chatView.Clear()
					return nil
				}

				// Verifica se é para sair
				if response == "Saindo..." {
					updateChatView("👋 " + response)
					time.Sleep(500 * time.Millisecond)
					return gocui.ErrQuit
				}

				// Verifica se é para enviar arquivo
				if strings.HasPrefix(msg, "/arquivo") {
					parts := strings.Fields(msg)
					if len(parts) < 2 {
						updateChatView("❌ Uso: /arquivo <caminho> [peer]")
						return nil
					}

					filePath := parts[1]
					var targetPeer string
					if len(parts) > 2 {
						targetPeer = parts[2]
					}

					go func() {
						if err := sendFile(filePath, targetPeer); err != nil {
							updateChatView("❌ " + err.Error())
						}
					}()

					return nil
				}

				// Comando normal, exibe resposta
				updateChatView(response)
				return nil
			}
		}

		// Mensagem normal
		chatView, _ := g.View("chat")
		fmt.Fprintf(chatView, "[%s] %s\n", nickname, msg)

		for _, peer := range peers {
			fmt.Fprintf(peer, "%s\n", msg)
		}

		return nil
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	logMessage("--- Sessão encerrada por " + nickname + " ---")
	return gocui.ErrQuit
}

// Função para atualizar a view de chat em tempo real
func updateChatView(message string) {
	uiMutex.Lock()
	defer uiMutex.Unlock()

	// Verifica se a UI já foi inicializada
	if g == nil {
		log.Println(message)
		return
	}

	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("chat")
		if err != nil {
			return err
		}

		fmt.Fprintln(v, message)
		return nil
	})
}

// Adiciona mensagem de sistema ao chat
func addSystemMessage(message string) {
	updateChatView("📢 " + message)
}
