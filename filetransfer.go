package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Tamanho do chunk para transferência de arquivos
const chunkSize = 4096

// Estrutura para um chunk de arquivo
type FileChunk struct {
	FileName    string `json:"filename"`
	TotalChunks int    `json:"total_chunks"`
	ChunkIndex  int    `json:"chunk_index"`
	Data        string `json:"data"` // Base64 encoded
	Sender      string `json:"sender"`
}

// Mapas para rastrear transferências em andamento
var (
	incomingFiles = make(map[string][][]byte)
	incomingMeta  = make(map[string]FileChunk)
	transferMutex sync.Mutex
)

// Diretório para salvar os arquivos recebidos
const filesDir = "received_files"

// Inicializa o sistema de transferência de arquivos
func initFileTransferSystem() error {
	// Cria diretório para arquivos recebidos se não existir
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		return err
	}
	return nil
}

// Envia um arquivo para um peer específico ou todos os peers
func sendFile(filePath string, targetPeer string) error {
	// Abre o arquivo
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo: %v", err)
	}
	defer file.Close()

	// Obtém informações do arquivo
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("erro ao obter informações do arquivo: %v", err)
	}

	if fileInfo.Size() > 100*1024*1024 {
		return fmt.Errorf("arquivo muito grande (limite: 100MB)")
	}

	// Calcula o número total de chunks
	totalChunks := int((fileInfo.Size() + chunkSize - 1) / chunkSize)
	fileName := filepath.Base(filePath)

	// Adiciona entrada no log
	logMessage(fmt.Sprintf("Iniciando envio do arquivo '%s' (%d bytes, %d chunks)",
		fileName, fileInfo.Size(), totalChunks))

	updateChatView(fmt.Sprintf("📤 Enviando arquivo '%s' (%d bytes)...", fileName, fileInfo.Size()))

	// Envia os chunks
	buffer := make([]byte, chunkSize)
	for chunkIndex := 0; chunkIndex < totalChunks; chunkIndex++ {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("erro ao ler arquivo: %v", err)
		}

		// Codifica o chunk em base64
		chunk := FileChunk{
			FileName:    fileName,
			TotalChunks: totalChunks,
			ChunkIndex:  chunkIndex,
			Data:        base64.StdEncoding.EncodeToString(buffer[:n]),
			Sender:      Nickname,
		}

		// Serializa o chunk em JSON
		jsonData, err := json.Marshal(chunk)
		if err != nil {
			return fmt.Errorf("erro ao serializar chunk: %v", err)
		}

		// Envia para o peer específico ou todos
		sent := false
		if targetPeer != "" {
			for addr, conn := range Peers {
				if strings.Contains(addr, targetPeer) {
					fmt.Fprintf(conn, "[FILE_TRANSFER]%s\n", string(jsonData))
					sent = true
					break
				}
			}
			if !sent {
				return fmt.Errorf("peer '%s' não encontrado", targetPeer)
			}
		} else {
			for _, conn := range Peers {
				fmt.Fprintf(conn, "[FILE_TRANSFER]%s\n", string(jsonData))
				sent = true
			}
		}

		if !sent {
			return fmt.Errorf("nenhum peer conectado para envio do arquivo")
		}

		// Atualiza status a cada 10% do progresso
		if chunkIndex%(totalChunks/10+1) == 0 || chunkIndex == totalChunks-1 {
			progress := float64(chunkIndex+1) / float64(totalChunks) * 100
			updateChatView(fmt.Sprintf("📤 Enviando '%s': %.1f%% concluído", fileName, progress))
		}
	}

	updateChatView(fmt.Sprintf("✅ Arquivo '%s' enviado com sucesso!", fileName))
	logMessage(fmt.Sprintf("Arquivo '%s' enviado com sucesso", fileName))
	return nil
}

// Processa um chunk de arquivo recebido
func handleFileChunk(data string) {
	var chunk FileChunk
	err := json.Unmarshal([]byte(data), &chunk)
	if err != nil {
		logMessage(fmt.Sprintf("Erro ao processar chunk de arquivo: %v", err))
		return
	}

	transferMutex.Lock()
	defer transferMutex.Unlock()

	// Gera um ID para o arquivo (nome + remetente)
	fileID := fmt.Sprintf("%s_%s", chunk.Sender, chunk.FileName)

	// Inicializa estruturas se necessário
	if _, exists := incomingFiles[fileID]; !exists {
		incomingFiles[fileID] = make([][]byte, chunk.TotalChunks)
		incomingMeta[fileID] = chunk
		updateChatView(fmt.Sprintf("📥 Recebendo arquivo '%s' de %s...", chunk.FileName, chunk.Sender))
	}

	// Decodifica e armazena o chunk
	data, err = base64.StdEncoding.DecodeString(chunk.Data)
	if err != nil {
		logMessage(fmt.Sprintf("Erro ao decodificar chunk: %v", err))
		return
	}

	incomingFiles[fileID][chunk.ChunkIndex] = data

	// Verifica se todos os chunks foram recebidos
	meta := incomingMeta[fileID]
	complete := true
	for _, c := range incomingFiles[fileID] {
		if c == nil {
			complete = false
			break
		}
	}

	// Atualiza o progresso a cada 10% ou quando concluído
	progress := float64(countReceivedChunks(fileID)) / float64(meta.TotalChunks) * 100
	if chunk.ChunkIndex%(meta.TotalChunks/10+1) == 0 || complete {
		updateChatView(fmt.Sprintf("📥 Recebendo '%s' de %s: %.1f%% concluído",
			meta.FileName, meta.Sender, progress))
	}

	// Se completo, salva o arquivo
	if complete {
		saveReceivedFile(fileID)
	}
}

// Conta quantos chunks foram recebidos para um arquivo
func countReceivedChunks(fileID string) int {
	count := 0
	for _, chunk := range incomingFiles[fileID] {
		if chunk != nil {
			count++
		}
	}
	return count
}

// Salva um arquivo recebido quando todos os chunks estiverem completos
func saveReceivedFile(fileID string) {
	meta := incomingMeta[fileID]
	chunks := incomingFiles[fileID]

	// Cria o caminho do arquivo
	filePath := filepath.Join(filesDir, meta.FileName)

	// Adiciona um sufixo se o arquivo já existir
	if _, err := os.Stat(filePath); err == nil {
		ext := filepath.Ext(filePath)
		baseName := strings.TrimSuffix(filePath, ext)
		filePath = fmt.Sprintf("%s_%s%s", baseName, meta.Sender, ext)
	}

	// Cria o arquivo
	file, err := os.Create(filePath)
	if err != nil {
		logMessage(fmt.Sprintf("Erro ao criar arquivo recebido: %v", err))
		updateChatView(fmt.Sprintf("❌ Erro ao salvar arquivo '%s': %v", meta.FileName, err))
		return
	}
	defer file.Close()

	// Escreve todos os chunks
	for _, chunk := range chunks {
		_, err := file.Write(chunk)
		if err != nil {
			logMessage(fmt.Sprintf("Erro ao escrever chunk: %v", err))
			updateChatView(fmt.Sprintf("❌ Erro ao salvar arquivo '%s': %v", meta.FileName, err))
			return
		}
	}

	// Limpa a memória
	delete(incomingFiles, fileID)
	delete(incomingMeta, fileID)

	updateChatView(fmt.Sprintf("✅ Arquivo '%s' recebido e salvo como '%s'",
		meta.FileName, filepath.Base(filePath)))
	logMessage(fmt.Sprintf("Arquivo '%s' recebido de %s e salvo como '%s'",
		meta.FileName, meta.Sender, filePath))
}
