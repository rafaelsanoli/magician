# 🧙 Magician — Chat Anônimo P2P com TLS

Um chat anônimo de terminal, descentralizado e criptografado, com reconexão automática e interface no terminal.  
Inspirado pelo [AnonChat](https://github.com/l50/anonchat), mas com foco em segurança, usabilidade e liberdade P2P.

---

## ✨ Funcionalidades

| Recurso                           | Descrição                                                                 |
|-----------------------------------|---------------------------------------------------------------------------|
| ✅ **Chat entre peers (P2P)**     | Comunicação direta entre usuários, sem servidor central                   |
| 🔐 **Criptografia TLS**           | Todas as conexões são criptografadas com certificados TLS (cert.pem / key.pem) |
| 🔁 **Reconexão automática**       | Conexões perdidas são restabelecidas automaticamente                     |
| 🧑‍💻 **Nickname personalizado**     | Cada usuário escolhe seu nome ao entrar                                   |
| 🔒 **Autenticação obrigatória**   | Todos os peers exigem senha ao se conectar (definida na inicialização)   |
| 💬 **Interface terminal (gocui)** | Interface moderna no terminal, com separação de input e rolagem          |
| 🧱 **Modularidade**               | Código dividido por responsabilidades: interface, peers, segurança, etc. |
| 🧭 **Descoberta automática**      | Descoberta de peers via UDP broadcast na rede local                      |
| 📜 **Comandos de terminal**       | Comandos como `/ajuda`, `/usuarios`, `/privado`, `/limpar`, `/logs`      |
| 📝 **Logs locais**                | Histórico das mensagens e eventos salvo em arquivos de log diários       |
| 📁 **Envio de arquivos** (Beta)   | Transferência de arquivos entre peers (implementação parcial)            |

---

## 🚀 Como usar

### 1. Clone o projeto

```bash
git clone https://github.com/seuusuario/magician-chat.git
cd magician-chat
```

### 2. Gere os certificados TLS

```bash
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365 -nodes
```

> Isso cria `cert.pem` e `key.pem` na raiz do projeto. Eles serão usados para criptografar as conexões.

### 3. Execute o chat

```go run .
```

Você será solicitado a informar:
- ✅ Seu **nickname**
- 🔒 Uma **senha obrigatória** (todos os peers devem usar a mesma)
- 📡 A **porta local** de escuta
- 🧭 Se deseja **ativar a descoberta automática** de peers

### 4. Conecte a outros peers

Você tem duas opções:

**Opção 1**: Deixe a descoberta automática encontrar peers na rede local (respondendo "s" à pergunta).

**Opção 2**: Informe manualmente o IP:porta de um peer existente quando solicitado.

---

## 📋 Comandos Disponíveis

| Comando                      | Descrição                                           |
|------------------------------|-----------------------------------------------------|
| `/ajuda`                     | Mostra a lista de comandos disponíveis              |
| `/usuarios`                  | Lista todos os peers conectados                     |
| `/privado <peer> <mensagem>` | Envia mensagem privada para um peer específico      |
| `/limpar`                    | Limpa a tela de chat                                |
| `/logs [n]`                  | Mostra as últimas n mensagens do log (padrão: 10)   |
| `/sair`                      | Fecha o chat                                        |
| `/arquivo <caminho> [peer]`  | Envia um arquivo para todos ou para um peer específico (beta) |

---

## 📦 Estrutura do projeto

```
magician/
├── main.go         # Ponto de entrada do programa
├── peer.go         # Lógica P2P: conexão, TLS, reconexão e autenticação
├── ui.go           # Interface de usuário com gocui
├── commands.go     # Implementação dos comandos de terminal
├── discovery.go    # Descoberta automática de peers via UDP broadcast
├── filetransfer.go # Sistema de transferência de arquivos (parcial)
├── cert.pem        # Certificado público TLS (gerado com OpenSSL)
├── key.pem         # Chave privada TLS (gerado com OpenSSL)
├── logs/           # Diretório onde são armazenados os logs diários
└── README.md       # Documentação do projeto
```

---

## 📝 Sistema de Logs

O chat mantém um registro de todas as mensagens e eventos em arquivos de log diários. Os logs são armazenados no diretório `logs/` com o formato `chat-YYYY-MM-DD.log`.

Para visualizar os logs dentro do chat, use o comando `/logs [n]`, onde `n` é o número de linhas que deseja ver (o padrão é 10).

---

## 🔮 Roadmap

Funcionalidades em desenvolvimento:

    ✅ 📁 Finalização do sistema de envio de arquivos com reconstrução e progresso

    🌐 Modo híbrido: P2P + servidor relay para conexões remotas

    🧠 Criptografia de ponta a ponta opcional (além de TLS)

    🧅 Integração com a rede Tor (via proxy SOCKS5 e .onion)

    🔔 Sistema de notificações para eventos importantes

    🗂️ Histórico de mensagens persistente com rolagem na interface

---

🧅 **Integração com a Rede Tor (planejada)**

A futura versão do Magician Chat incluirá suporte a comunicação via rede Tor para anonimato extremo.  
Será possível rodar o chat com:

    Proxy SOCKS5 apontando para o Tor local (127.0.0.1:9050)

    Conexões entre peers através de .onion addresses

    Relay via hidden service opcional

    Comunicação criptografada ponta a ponta sobre o Tor

Essa funcionalidade ainda está em desenvolvimento, mas é parte do roadmap oficial para garantir privacidade e descentralização totais.

---

## 📋 Resolução de Problemas

### Conexão Recusada

Se você receber um erro "Conexão recusada", verifique:

1. Se a porta no peer de destino está aberta e livre
2. Se o firewall está permitindo conexões naquela porta
3. Se ambos os peers estão usando a mesma senha

### Certificados TLS

Se ocorrerem erros relacionados aos certificados:

1. Verifique se os arquivos `cert.pem` e `key.pem` estão na raiz do projeto
2. Regenere os certificados usando o comando OpenSSL fornecido acima

---

## ✅ Requisitos

- [Go](https://golang.org/dl/) 1.20+
- [gocui](https://github.com/jroimartin/gocui) (Para a interface)
- [OpenSSL](https://www.openssl.org/) (para gerar certificados TLS)

Para instalar as dependências:

```bash
go get github.com/jroimartin/gocui
```

---

## 📜 Licença

MIT License © 2025  
Feito com 🖤 por quem acredita na liberdade digital.
