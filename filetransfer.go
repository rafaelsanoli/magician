
// Função sendFile com proteção mutex
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
peersMutex.Lock()
if targetPeer != "" {
for addr, conn := range Peers {
if strings.Contains(addr, targetPeer) {
fmt.Fprintf(conn, "[FILE_TRANSFER]%s\n", string(jsonData))
sent = true
break
}
}
} else {
for _, conn := range Peers {
fmt.Fprintf(conn, "[FILE_TRANSFER]%s\n", string(jsonData))
sent = true
}
}
peerCount := len(Peers)
peersMutex.Unlock()

if !sent && targetPeer != "" {
return fmt.Errorf("peer '%s' não encontrado", targetPeer)
} else if !sent && peerCount > 0 {
return fmt.Errorf("falha ao enviar para qualquer peer")
} else if !sent {
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

// Melhoria na função handleFileChunk para evitar potencial pânico com JSON inválido
func handleFileChunk(data string) {
var chunk FileChunk
err := json.Unmarshal([]byte(data), &chunk)
if err != nil {
logMessage(fmt.Sprintf("Erro ao processar chunk de arquivo: %v", err))
updateChatView(fmt.Sprintf("❌ Erro ao processar chunk de arquivo: %v", err))
return
}

// Validações adicionais
if chunk.FileName == "" || chunk.TotalChunks <= 0 || chunk.ChunkIndex < 0 || chunk.ChunkIndex >= chunk.TotalChunks {
logMessage("Recebido chunk de arquivo com dados inválidos")
updateChatView("❌ Recebido chunk de arquivo com dados inválidos")
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

// Verifica se o índice é válido para o array
meta := incomingMeta[fileID]
if chunk.ChunkIndex >= len(incomingFiles[fileID]) {
// Realocar o array se preciso
newArray := make([][]byte, chunk.TotalChunks)
copy(newArray, incomingFiles[fileID])
incomingFiles[fileID] = newArray
}

// Decodifica e armazena o chunk
chunkData, err := base64.StdEncoding.DecodeString(chunk.Data)
if err != nil {
logMessage(fmt.Sprintf("Erro ao decodificar chunk: %v", err))
updateChatView(fmt.Sprintf("❌ Erro ao decodificar chunk: %v", err))
return
}

incomingFiles[fileID][chunk.ChunkIndex] = chunkData

// Verifica se todos os chunks foram recebidos
complete := true
for _, c := range incomingFiles[fileID] {
if c == nil {
complete = false
break
}
}

// Evita divisão por zero se totalChunks for pequeno
updateInterval := meta.TotalChunks / 10
if updateInterval < 1 {
updateInterval = 1
}

// Atualiza o progresso a cada 10% ou quando concluído
progress := float64(countReceivedChunks(fileID)) / float64(meta.TotalChunks) * 100
if chunk.ChunkIndex%updateInterval == 0 || complete {
updateChatView(fmt.Sprintf("📥 Recebendo '%s' de %s: %.1f%% concluído",
meta.FileName, meta.Sender, progress))
}

// Se completo, salva o arquivo
if complete {
saveReceivedFile(fileID)
}
}