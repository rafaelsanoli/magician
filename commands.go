
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

// Correção para o arquivo commands.go - Função cmdListUsers com mutex
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

// Correção para o arquivo commands.go - Função sendMessage com mutex (usada em ui.go)
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
// Envia mensagem para todos os peers
peersMutex.Lock()
for _, conn := range Peers {
fmt.Fprintf(conn, "[%s] %s\n", Nickname, message)
}
peersMutex.Unlock()

updateChatView(fmt.Sprintf("[Você] %s", message))
}

v.Clear()
v.SetCursor(0, 0)
}
return nil
}