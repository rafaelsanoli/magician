package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CommandHandler define uma função que processa um comando
type CommandHandler func(args []string) string

// Mapa de comandos disponíveis
var commands = map[string]CommandHandler{
	"ajuda":    cmdHelp,
	"usuarios": cmdListUsers,
	"sair":     cmdExit,
	"limpar":   cmdClear,
	"privado":  cmdPrivateMsg,
	"logs":     cmdShowLogs,
	"arquivo":  cmdSendFile,
}

// Define a pasta de logs
const logsDir = "logs"

// Processa uma mensagem, verificando se é um comando
func processMessage(message string) (bool, string) {
	// Se não começar com /, não é um comando
	if !strings.HasPrefix(message, "/") {
		// Loga mensagem normal
		logMessage(fmt.Sprintf("[%s] %s", Nickname, message))
		return false, ""
	}

	// Remove a barra e divide em argumentos
	parts := strings.Fields(strings.TrimPrefix(message, "/"))
	if len(parts) == 0 {
		return true, "Comando vazio. Use /ajuda para ver os comandos disponíveis."
	}

	// Obtém o nome do comando e os argumentos
	cmdName := strings.ToLower(parts[0])
	args := parts[1:]

	// Verifica se o comando existe
	handler, exists := commands[cmdName]
	if !exists {
		return true, fmt.Sprintf("Comando desconhecido: /%s. Use /ajuda para ver os comandos disponíveis.", cmdName)
	}

	// Executa o comando
	result := handler(args)

	// Loga o comando executado
	logMessage(fmt.Sprintf("[COMANDO] /%s %s", cmdName, strings.Join(args, " ")))

	return true, result
}

// Comando /ajuda - Lista todos os comandos disponíveis
func cmdHelp(args []string) string {
	help := "📚 Comandos disponíveis:\n" +
		"/ajuda - Mostra esta ajuda\n" +
		"/usuarios - Lista todos os peers conectados\n" +
		"/privado <peer> <mensagem> - Envia mensagem privada para um peer específico\n" +
		"/limpar - Limpa a tela de chat\n" +
		"/logs [n] - Mostra as últimas n mensagens do log (padrão: 10)\n" +
		"/arquivo <caminho> [peer] - Envia um arquivo para todos ou para um peer específico\n" +
		"/sair - Fecha o chat"

	return help
}

// Comando /usuarios - Lista todos os peers conectados
func cmdListUsers(args []string) string {
	if len(Peers) == 0 {
		return "Nenhum peer conectado."
	}

	result := fmt.Sprintf("🔌 Peers conectados (%d):\n", len(Peers))
	i := 1
	for addr := range Peers {
		result += fmt.Sprintf("%d. %s\n", i, addr)
		i++
	}

	return result
}

// Comando /sair - Encerra o programa
func cmdExit(args []string) string {
	// O fechamento real é tratado em sendMessage
	return "Saindo..."
}

// Comando /limpar - Limpa o chat
func cmdClear(args []string) string {
	// A limpeza real é tratada em sendMessage
	return "[LIMPAR]"
}

// Comando /privado - Envia mensagem privada para um peer específico
func cmdPrivateMsg(args []string) string {
	if len(args) < 2 {
		return "Uso: /privado <peer> <mensagem>"
	}

	target := args[0]
	message := strings.Join(args[1:], " ")

	// Procura o peer pelo endereço (completo ou parcial)
	var targetConn net.Conn
	var targetAddr string

	for addr, conn := range Peers {
		if strings.Contains(addr, target) {
			targetConn = conn
			targetAddr = addr
			break
		}
	}

	if targetConn == nil {
		return fmt.Sprintf("Peer '%s' não encontrado.", target)
	}

	// Envia mensagem privada
	fmt.Fprintf(targetConn, "[PRIVADO] %s\n", message)

	// Loga a mensagem privada
	logMessage(fmt.Sprintf("[PRIVADO para %s] %s", targetAddr, message))

	return fmt.Sprintf("Mensagem privada enviada para %s", targetAddr)
}

// Comando /logs - Mostra as últimas mensagens do log
func cmdShowLogs(args []string) string {
	// Número de linhas padrão
	n := 10

	// Se fornecido um argumento, usa como número de linhas
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &n)
		if n <= 0 {
			n = 10
		}
	}

	// Lê as últimas n linhas do log
	lines, err := readLastNLines(getLogFilePath(), n)
	if err != nil {
		return fmt.Sprintf("Erro ao ler logs: %s", err)
	}

	if len(lines) == 0 {
		return "Nenhuma mensagem no log."
	}

	result := fmt.Sprintf("📜 Últimas %d mensagens:\n", len(lines))
	for _, line := range lines {
		result += line + "\n"
	}

	return result
}

// Comando /arquivo - Envia um arquivo para um peer específico ou todos
func cmdSendFile(args []string) string {
	if len(args) < 1 {
		return "Uso: /arquivo <caminho> [peer]"
	}

	filePath := args[0]
	var targetPeer string
	if len(args) > 1 {
		targetPeer = args[1]
	}

	// Inicia envio em goroutine separada
	go func() {
		if err := sendFile(filePath, targetPeer); err != nil {
			updateChatView("❌ " + err.Error())
		}
	}()

	return fmt.Sprintf("Iniciando envio do arquivo: %s", filePath)
}

// Inicializa o sistema de logs
func initLogSystem() error {
	// Cria diretório de logs se não existir
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	// Testa se podemos escrever no arquivo de log
	logFile := getLogFilePath()
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Escreve entrada inicial no log
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, err = fmt.Fprintf(file, "[%s] --- Sessão iniciada por %s ---\n", timestamp, Nickname)

	return err
}

// Adiciona uma mensagem ao log
func logMessage(message string) {
	logFile := getLogFilePath()
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Erro ao abrir arquivo de log: %v", err)
		return
	}
	defer file.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, err = fmt.Fprintf(file, "[%s] %s\n", timestamp, message)
	if err != nil {
		log.Printf("Erro ao escrever no log: %v", err)
	}
}

// Retorna o caminho para o arquivo de log do dia atual
func getLogFilePath() string {
	today := time.Now().Format("2006-01-02")
	return filepath.Join(logsDir, fmt.Sprintf("chat-%s.log", today))
}

// Lê as últimas n linhas de um arquivo
func readLastNLines(filePath string, n int) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		// Se o arquivo não existe, retorna lista vazia sem erro
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	// Método simples: lê todas as linhas e pega as últimas n
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Retorna as últimas n linhas ou todas se houver menos que n
	if len(lines) <= n {
		return lines, nil
	}
	return lines[len(lines)-n:], nil
}
