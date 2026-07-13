<div align="center">

# 📞 NB_Api

**API WhatsApp com suporte a mensagens interativas, chamadas de voz e muito mais.**

Este projeto oferece uma solução robusta de API conectada diretamente ao WhatsApp, permitindo envio de mensagens, gerenciamento de grupos, visualização de histórico e realização de chamadas de voz (VoIP) diretamente do navegador.

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=black)](https://react.dev)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Ready-336791?logo=postgresql&logoColor=white)](https://postgresql.org)
[![Redis](https://img.shields.io/badge/Redis-Ready-DC382D?logo=redis&logoColor=white)](https://redis.io)

</div>

---

## 🚀 Principais Funcionalidades

- **Multi-Sessões:** Conecte múltiplas contas de WhatsApp lendo o QR Code diretamente pela interface.
- **Chamadas de Voz Nativas:** Faça e receba chamadas de voz diretamente pelo navegador. O áudio do microfone trafega via WebRTC para o servidor Go, que encoda nativamente (MLow) e repassa para a rede do WhatsApp (SRTP).
- **Gestão Completa (REST API):**
  - **Mensagens:** Envio de texto, mídia (fotos, áudios, vídeos, documentos), localização, listas interativas e reações.
  - **Grupos:** Criação, visualização de informações, gestão de participantes e links de convite.
  - **Contatos e Perfil:** Leitura da agenda, alteração de foto de perfil, status e configurações de privacidade.
  - **Normalização de números BR:** Tenta automaticamente com e sem o dígito `9` para números brasileiros.
- **Armazenamento Seguro:** As sessões de conexão e configurações ficam salvas no **PostgreSQL**, com o **Redis** servindo como camada de cache e mensageria em tempo real.
- **Interface Swagger:** Documentação interativa da API disponível em `/swagger`.

---

## 🏗️ Arquitetura do Sistema

O ecossistema é dividido em duas partes principais:

1. **Backend (Go - `cmd/server`):** Responsável por gerenciar a conexão com o WhatsApp usando a biblioteca `whatsmeow`, processar os pacotes de áudio (WebRTC/RTP/SRTP), expor a REST API e gerenciar as conexões WebSocket (SSE) com os clientes.
2. **Frontend (React - `client/`):** Interface moderna feita em React 19, Vite e TailwindCSS, para interagir com o WhatsApp, realizar chamadas e gerenciar as contas conectadas.

---

## 🛠️ Requisitos e Configuração

- **Go 1.26+**
- **Node.js 22+** (para o Frontend)
- **PostgreSQL** (Banco de dados relacional para persistência de sessões)
- **Redis** (Gerenciamento de cache e comunicação interna)

### 1. Clonar e Instalar Dependências

```bash
# Dependências do Backend (Go)
go mod download

# Dependências do Frontend (React)
cd client
npm install
cd ..
```

### 2. Configurar Variáveis de Ambiente (.env)

Crie ou edite o arquivo `.env` na raiz do projeto:

```env
# Banco de Dados (PostgreSQL)
DB_HOST=192.168.50.2
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=sua_senha
DB_NAME=nb_api
DB_SSLMODE=disable

# Redis
REDIS_HOST=192.168.50.2:6379
REDIS_PASSWORD=
REDIS_DB=0

# Autenticação da API (Basic Auth)
AUTH_USERNAME=admin
AUTH_PASSWORD=admin123
```

> **Nota:** Se `AUTH_USERNAME` estiver vazio, a autenticação é desabilitada.

---

## ⚙️ Como Executar

Para facilitar a inicialização no Windows, foram criados arquivos `.bat` na raiz do projeto:

### Opção A: Ambiente de Desenvolvimento (Recomendado para programar)
Dê um duplo clique no arquivo **`start.bat`**.
Ele abrirá duas janelas do terminal automaticamente:
- Uma rodando a API do backend na porta `:8080`
- Outra rodando o servidor de desenvolvimento do React (Vite) na porta `:5173`
- *Qualquer alteração no código do frontend será refletida instantaneamente no navegador.*

### Opção B: Ambiente de Produção (Uso diário)
Se você não vai alterar o código e quer apenas **usar** o sistema de forma limpa e unificada na porta `:8080`:

1. Compile o frontend primeiro (se houve alguma alteração recente):
   ```bash
   cd client && npm run build && cd ..
   ```
2. Dê um duplo clique no arquivo **`start-prod.bat`**.
   Ele vai rodar o servidor Go, que servirá tanto a API quanto a interface web compilada estática. Acesse: 👉 `http://localhost:8080`

---

## 📚 Documentação da API

A API possui **57 endpoints** documentados interativamente no Swagger UI:

👉 **http://localhost:8080/swagger** (acesso livre, sem autenticação)

### Estrutura

Todas as rotas baseadas em ação ocorrem dentro de uma `{sid}` (ID da Sessão).

| Categoria | Exemplo de Rota | Funcionalidade |
|---|---|---|
| **Sessões** | `POST /api/sessions` | Cria nova sessão (gera QR Code). O `sid` é o nome slugificado |
| **Chamadas** | `POST /api/sessions/{sid}/calls` | Inicia uma chamada de voz (VoIP) |
| **Mensagens**| `POST /api/sessions/{sid}/messages/text` | Envia uma mensagem de texto |
| **Lista (botões)** | `POST /api/sessions/{sid}/messages/list` | Envia botões interativos (quick_reply). Substitui o ListMessage, depreciado pelo WhatsApp. |
| **Lista Interativa** | `POST /api/sessions/{sid}/messages/list-interactive` | Envia lista dropdown com seções e linhas para menus com muitas opções. |
| **Enquete** | `POST /api/sessions/{sid}/messages/poll` | Envia enquete interativa com múltiplas opções. |

| **Grupos** | `GET /api/sessions/{sid}/groups` | Lista grupos |
| **Contatos** | `GET /api/sessions/{sid}/contacts` | Busca todos os contatos salvos |
| **Perfil** | `PUT /api/sessions/{sid}/profile/photo`| Atualiza a foto de perfil |
| **Eventos** | `GET /api/events` | Stream em tempo real (SSE) de status |

### SID (Session ID)

O `sid` é gerado automaticamente a partir do **nome da instância** no momento da criação:

| Nome informado | sid gerado |
|---|---|
| `Minha Sessao` | `Minha-Sessao` |
| `Instância 1` | `Instância-1` |
| `Teste` | `Teste` |

Se o nome for vazio, um ID aleatório é gerado. Se houver conflito, um sufixo numérico é adicionado (`Instancia-2`, `Instancia-3`, ...).

### Autenticação

Todas as requisições à API (exceto Swagger) exigem **HTTP Basic Auth**:

```bash
curl -u admin:admin123 http://localhost:8080/api/sessions
```

As credenciais são definidas no `.env` via `AUTH_USERNAME` e `AUTH_PASSWORD`.

### Mensagens com Botões Interativos (Lista / Quick Reply)

> **Nota:** O WhatsApp **depreciou** o formato `ListMessage` (lista com seções). Desde meados de 2023, mensagens nesse formato são silenciosamente descartadas pelo cliente do destinatário.

O endpoint `POST /api/sessions/{sid}/messages/list` agora envia **botões interativos** (`quick_reply`) usando o formato moderno `InteractiveMessage` + `NativeFlowMessage`, que funciona em Android, iOS e Web.

O corpo da requisição mantém o mesmo formato de `sections`/`rows` para compatibilidade, mas cada `row` vira um botão clicável:

```json
{
  "to": "5531992604940",
  "title": "Menu de Atendimento",
  "description": "Escolha uma opção abaixo:",
  "footerText": "Toque em um botão para responder",
  "sections": [
    {
      "title": "Seção (ignorada no quick_reply)",
      "rows": [
        { "title": "Falar com Vendas",   "rowId": "vendas" },
        { "title": "Falar com Suporte",  "rowId": "suporte" },
        { "title": "Falar com Financeiro", "rowId": "financeiro" }
      ]
    }
  ]
}
```

Quando o destinatário toca em um botão, o `rowId` é enviado de volta como uma mensagem de resposta (evento `InteractiveResponseMessage`), que pode ser capturado via WebSocket/SSE.

### Normalização de Números Brasileiros

Ao enviar mensagens ou fazer chamadas, o sistema tenta automaticamente ambas as versões do número (com e sem o dígito `9`) via `IsOnWhatsApp`, usando a que existir:

| Número digitado | Testa |
|---|---|
| `553197302067` (12 dígitos) | Original + com 9 (`553197302067` / `5531973020679`) |
| `5531973070267` (13 dígitos) | Original + sem 9 (`5531973070267` / `553197302067`) |

---

## 🔐 Segurança e Boas Práticas

- **Credenciais:** Mantenha o seu arquivo `.env` seguro. Ele contém as chaves e dados de acesso à sua infraestrutura de sessões do WhatsApp.
- **Basic Auth:** A API é protegida por HTTP Basic Auth. Configure `AUTH_USERNAME`/`AUTH_PASSWORD` no `.env`.
- **Rede Local:** Recomenda-se rodar o sistema em rede local confiável (LAN) ou protegida por proxy reverso com HTTPS.

---

## 🙌 Créditos

Este projeto foi construído com base no trabalho de:

- [jotadev66](https://github.com/jotadev66)
- [jobasfernandes](https://github.com/jobasfernandes)
- [edgardmessias](https://github.com/edgardmessias)
- [w3nder](https://github.com/w3nder)

## Agradecimentos

- [whatsmeow](https://github.com/tulir/whatsmeow) — Biblioteca Go para o protocolo WhatsApp Web
- [pion/webrtc](https://github.com/pion/webrtc) — Stack WebRTC em Go puro (ICE + DTLS + SCTP)
- [whatsapp-rust](https://github.com/WhatsApp/whatsapp-rust) — Implementação de referência do codec MLow (portada para Go puro em internal/voip/media/mlow)
- [zapo](https://github.com/mattermost/zapo) — Referência de stack de mídia VoIP
