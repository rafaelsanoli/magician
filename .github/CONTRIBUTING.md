# 🤝 Contribuindo com o Magician Chat

Obrigado por considerar contribuir com o Magician Chat! Este projeto depende da colaboração da comunidade para crescer.

---

## 🧰 Pré-requisitos

Antes de começar, certifique-se de ter instalado:

- [Go](https://golang.org/dl/) 1.20 ou superior
- [OpenSSL](https://www.openssl.org/) para gerar certificados TLS
- Git

Clone o projeto:

```bash
git clone https://github.com/seuusuario/magician-chat.git
cd magician-chat
go mod tidy
```

---

## 🛠️ Como contribuir

### 1. Crie uma branch para sua feature ou fix

```bash
git checkout -b feat/nome-da-feature
```

### 2. Faça as alterações

- Siga o estilo do código existente
- Comente funções importantes
- Prefira modularização se possível

### 3. Teste localmente

```bash
go run .
```

### 4. Commit e push

```bash
git add .
git commit -m "✨ feat: descrição clara da mudança"
git push origin feat/nome-da-feature
```

### 5. Abra um Pull Request

- Use o template de PR
- Descreva claramente o que foi feito
- Marque se há dependência de outro PR

---

## 🧪 Boas práticas

- Mensagens de commit no formato: `tipo: descrição`  
  Exemplos: `fix: corrige reconexão TLS`, `feat: adiciona suporte a proxy Tor`
- Use comandos como `/ajuda` para testar interações
- Adicione logs úteis com `logMessage` se necessário

---

## 💬 Precisa de ajuda?

Abra uma [issue](https://github.com/seuusuario/magician-chat/issues) com sua dúvida ou sugestão.  
Ficaremos felizes em colaborar!

---

Feito com 🖤 por quem acredita na liberdade digital.