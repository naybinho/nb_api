<div align="center">

# 📞 NB_Api

**API WhatsApp com suporte a mensagens interativas, chamadas de voz e muito mais.**

Este projeto oferece uma solução robusta de API conectada diretamente ao WhatsApp, permitindo envio de mensagens, gerenciamento de grupos, visualização de histórico e realização de chamadas de voz (VoIP) diretamente do navegador.

[![Version](https://img.shields.io/badge/Version-1.0.7-blue)](https://github.com/naybinho/nb_api)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=black)](https://react.dev)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Ready-336791?logo=postgresql&logoColor=white)](https://postgresql.org)
[![Redis](https://img.shields.io/badge/Redis-Ready-DC382D?logo=redis&logoColor=white)](https://redis.io)
[![Docker](https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=white)](https://hub.docker.com/r/naybinho/nb_api)

</div>

---

## 🚀 Principais Funcionalidades

- **Multi-Sessões:** Conecte múltiplas contas de WhatsApp lendo o QR Code diretamente pela interface. As sessões são restauradas automaticamente ao reiniciar o servidor.
- **Chamadas de Voz Nativas:** Faça e receba chamadas de voz diretamente pelo navegador. O áudio do microfone trafega via WebRTC para o servidor Go, que encoda nativamente (MLow) e repassa para a rede do WhatsApp (SRTP).
- **Gestão Completa de Sessões:** Criação, logout, re-pareamento (QR Code), atualização de nome e chave de API (API Key) por sessão.
- **Pagamentos PIX:** Geração de QR Code PIX estático (EMV), envio do código copia e cola e/ou imagem do QR Code via WhatsApp, suporte a todos os tipos de chave (CPF, CNPJ, telefone, e-mail, aleatória), validação de chaves (CPF/CNPJ com dígitos verificadores) e valor opcional.
- **Mensagens Avançadas:** Envio de texto, mídia (fotos, áudios, vídeos, documentos, stickers), localização, contatos, listas interativas (quick_reply), listas dropdown, enquetes, reações, edição e revogação de mensagens, marcação de leitura e download de mídia.
- **Grupos:** Criação, atualização, gestão de participantes (adicionar/remover/promover/rebaixar), links de convite (criar/revogar), pedidos de entrada (listar/aprovar) e foto do grupo.
- **Contatos e Perfil:** Leitura da agenda, informações de contato, avatar, subscrição de presença, alteração de nome, status, foto de perfil e configurações de privacidade (ler/escrever).
- **Blocklist:** Gerenciamento de blocagem/desbloqueio de contatos.
- **Newsletters/Canais:** Criação, visualização e deixar de seguir canais/ newsletters do WhatsApp.
- **Normalização de números BR:** Tenta automaticamente com e sem o dígito `9` para números brasileiros.
- **Eventos em Tempo Real (SSE):** Stream de eventos de sessão, mensagens, presença, chamadas e mais via `GET /api/events`.
- **Armazenamento Seguro:** As sessões de conexão e configurações ficam salvas no **PostgreSQL**, com o **Redis** servindo como camada de cache e mensageria em tempo real.
- **Interface Swagger:** Documentação interativa da API disponível em `/swagger`.
- **Limite de Chamadas Simultâneas:** Configure o número máximo de chamadas simultâneas por sessão via flag `-max-calls-per-session`.

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

### Flags do Servidor

O binário do servidor (`cmd/server`) aceita as seguintes flags na linha de comando:

| Flag | Padrão | Descrição |
|---|---|---|
| `-addr` | `:8080` | Endereço HTTP de escuta |
| `-static` | `client/dist` | Diretório com os arquivos estáticos do frontend |
| `-debug` | `false` | Habilita logs verbosos para depuração |
| `-max-calls-per-session` | `8` | Limite de chamadas simultâneas por sessão (`0` = ilimitado) |
| `-swagger-url` | `""` | URL pública do servidor para o Swagger (ex: `http://192.168.1.100:8080`). Se vazio, tenta detectar automaticamente. Também lê da variável de ambiente `SWAGGER_URL`. |

---

## ⚙️ Como Executar

Para facilitar a inicialização no Windows, foram criados arquivos `.bat` na raiz do projeto:

### Opção A: Ambiente de Desenvolvimento (Recomendado para programar)
Dê um duplo clique no arquivo **`start-dev.bat`**.
Ele abrirá duas janelas do terminal automaticamente:
- Uma rodando a API do backend na porta `:8081`
- Outra rodando o servidor de desenvolvimento do React (Vite) na porta `:5173` com HMR
- *Qualquer alteração no código do frontend será refletida instantaneamente no navegador.*

### Opção B: Ambiente de Produção (Uso diário)
Se você não vai alterar o código e quer apenas **usar** o sistema de forma limpa e unificada na porta `:8080`:

Dê um duplo clique no arquivo **`start.bat`**.
Ele vai compilar o frontend automaticamente e rodar o servidor Go, que servirá tanto a API quanto a interface web compilada estática. Acesse: 👉 `http://localhost:8080`

> **Nota:** Se você já compilou o frontend manualmente, pode usar **`start-prod.bat`** para pular a etapa de compilação.

### Opção C: Docker (Recomendado para implantação)

Certifique-se de que o `docker-compose.yml` está configurado com as credenciais corretas do PostgreSQL e Redis, e execute:

```bash
docker compose up -d
```

Acesse: 👉 `http://SEU_IP:8080`

> ⚠️ **WebRTC em Docker — Linux (recomendado):**
> O `docker-compose.yml` já utiliza `network_mode: host`, que faz o container compartilhar a rede do host diretamente. Isso resolve os problemas de WebRTC sem precisar configurar IP externo.
>
> ```yaml
> services:
>   api:
>     network_mode: host
> ```
>
> ⚠️ **WebRTC em Docker Desktop (Windows/Mac):**
> O `network_mode: host` não funciona no Docker Desktop. Nesse caso, remova o `network_mode: host` e descomente a variável `EXTERNAL_IP` no `docker-compose.yml` com o IP da máquina host:
>
> ```yaml
> environment:
>   EXTERNAL_IP: "192.168.1.100"   # <- substitua pelo IP da sua máquina
> ```

---

## 📚 Documentação da API

A API possui **mais de 60 endpoints** documentados interativamente no Swagger UI:

👉 **http://localhost:8080/swagger** (acesso livre, sem autenticação)

### Estrutura

Todas as rotas baseadas em ação (exceto `GET /api/events` e `/swagger`) ocorrem dentro de uma `{sid}` (ID da Sessão).

| Categoria | Método | Rota | Funcionalidade |
|---|---|---|---|
| **Sessões** | `GET` | `/api/sessions` | Lista todas as sessões |
| | `POST` | `/api/sessions` | Cria nova sessão (gera QR Code) |
| | `DELETE` | `/api/sessions/{sid}` | Remove uma sessão |
| | `POST` | `/api/sessions/{sid}/logout` | Desconecta a sessão do WhatsApp |
| | `POST` | `/api/sessions/{sid}/pair` | Reinicia o pareamento (novo QR Code) |
| | `PUT` | `/api/sessions/{sid}/apikey` | Atualiza a chave de API da sessão |
| | `PUT` | `/api/sessions/{sid}/name` | Renomeia a sessão |
| **Chamadas** | `POST` | `/api/sessions/{sid}/calls` | Inicia uma chamada de voz (VoIP) |
| | `POST` | `/api/sessions/{sid}/calls/{id}/webrtc` | Inicia o fluxo WebRTC para uma chamada |
| | `POST` | `/api/sessions/{sid}/calls/{id}/accept` | Aceita uma chamada recebida |
| | `POST` | `/api/sessions/{sid}/calls/{id}/reject` | Rejeita uma chamada recebida |
| | `DELETE` | `/api/sessions/{sid}/calls/{id}` | Encerra uma chamada |
| | `GET` | `/api/sessions/{sid}/history` | Histórico de chamadas da sessão |
| **Mensagens** | `POST` | `/api/sessions/{sid}/messages` | Dispatcher genérico (texto/mídia/localização/contato/lista) |
| | `POST` | `/api/sessions/{sid}/messages/text` | Envia mensagem de texto |
| | `POST` | `/api/sessions/{sid}/messages/media` | Envia mídia (imagem/vídeo/áudio/documento/sticker) |
| | `POST` | `/api/sessions/{sid}/messages/location` | Envia localização |
| | `POST` | `/api/sessions/{sid}/messages/contact` | Envia contato |
| | `POST` | `/api/sessions/{sid}/messages/list` | Envia botões interativos (quick_reply) |
| | `POST` | `/api/sessions/{sid}/messages/list-interactive` | Envia lista dropdown com seções e linhas |
| | `POST` | `/api/sessions/{sid}/messages/poll` | Envia enquete interativa |
| | `POST` | `/api/sessions/{sid}/messages/react` | Envia reação a uma mensagem |
| | `POST` | `/api/sessions/{sid}/messages/edit` | Edita mensagem enviada |
| | `DELETE` | `/api/sessions/{sid}/messages/{id}` | Revoga/exclui uma mensagem |
| | `POST` | `/api/sessions/{sid}/messages/{id}/read` | Marca mensagem como lida |
| | `GET` | `/api/sessions/{sid}/media/{id}` | Baixa mídia de uma mensagem |
| **PIX** | `POST` | `/api/sessions/{sid}/messages/pix` | Envia PIX via WhatsApp (QR Code + texto) |
| | `POST` | `/api/pix/generate` | Gera QR Code PIX (pré-visualização, sem enviar) |
| | `POST` | `/api/pix/validate` | Valida chave PIX (CPF/CNPJ com dígitos verificadores) |
| **Grupos** | `GET` | `/api/sessions/{sid}/groups` | Lista grupos da sessão |
| | `POST` | `/api/sessions/{sid}/groups` | Cria um grupo |
| | `GET` | `/api/sessions/{sid}/groups/{gid}` | Obtém informações do grupo |
| | `PUT` | `/api/sessions/{sid}/groups/{gid}` | Atualiza nome, tópico, locked e announce |
| | `DELETE` | `/api/sessions/{sid}/groups/{gid}` | Sai do grupo |
| | `GET` | `/api/sessions/{sid}/groups/{gid}/invite` | Obtém link de convite do grupo |
| | `POST` | `/api/sessions/{sid}/groups/{gid}/invite/revoke` | Revoga o link de convite |
| | `POST` | `/api/sessions/{sid}/groups/join` | Entra em grupo via link de convite |
| | `POST` | `/api/sessions/{sid}/groups/{gid}/participants` | Adiciona/remove participantes |
| | `PUT` | `/api/sessions/{sid}/groups/{gid}/participants/{jid}` | Promove/demove um participante |
| | `GET` | `/api/sessions/{sid}/groups/{gid}/requests` | Lista pedidos de entrada no grupo |
| | `POST` | `/api/sessions/{sid}/groups/{gid}/requests` | Aprova pedidos de entrada |
| | `PUT` | `/api/sessions/{sid}/groups/{gid}/photo` | Define foto do grupo |
| | `DELETE` | `/api/sessions/{sid}/groups/{gid}/photo` | Remove a foto do grupo |
| **Usuários/Contatos** | `GET` | `/api/sessions/{sid}/users` | Busca informações de um ou mais usuários (JIDs) |
| | `GET` | `/api/sessions/{sid}/users/{jid}/presence` | Assina presença de um usuário |
| | `POST` | `/api/sessions/{sid}/users/check` | Verifica se números estão no WhatsApp |
| | `GET` | `/api/sessions/{sid}/contacts` | Lista contatos salvos |
| | `GET` | `/api/sessions/{sid}/contacts/{jid}` | Busca detalhes de um contato |
| | `GET` | `/api/sessions/{sid}/contacts/{jid}/avatar` | Busca avatar/foto do contato |
| **Perfil** | `PUT` | `/api/sessions/{sid}/profile/name` | Atualiza nome de perfil |
| | `PUT` | `/api/sessions/{sid}/profile/status` | Atualiza status (recado) |
| | `PUT` | `/api/sessions/{sid}/profile/photo` | Atualiza foto de perfil |
| | `GET` | `/api/sessions/{sid}/profile/privacy` | Lê configurações de privacidade |
| | `PUT` | `/api/sessions/{sid}/profile/privacy` | Atualiza configurações de privacidade |
| **Blocklist** | `GET` | `/api/sessions/{sid}/blocklist` | Lista contatos bloqueados |
| | `POST` | `/api/sessions/{sid}/blocklist` | Adiciona/remove da blocklist |
| **Newsletters** | `POST` | `/api/sessions/{sid}/newsletters` | Cria newsletter/canal |
| | `GET` | `/api/sessions/{sid}/newsletters/{jid}` | Obtém informações da newsletter |
| | `DELETE` | `/api/sessions/{sid}/newsletters/{jid}` | Deixa de seguir newsletter |
| | `POST` | `/api/sessions/{sid}/newsletters/{jid}/mute` | Alterna mute da newsletter |
| **Eventos** | `GET` | `/api/events` | Stream SSE em tempo real |

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

## � Eventos em Tempo Real (SSE)

O endpoint `GET /api/events` fornece um stream **Server-Sent Events (SSE)** com todos os eventos do sistema em tempo real. É o mecanismo principal para manter a interface atualizada sem polling.

```bash
curl -u admin:admin123 http://localhost:8080/api/events
```

### Tipos de Eventos

| Evento | Descrição |
|---|---|
| `session:update` | Mudança no estado de uma sessão (conectado, desconectado, QR Code) |
| `message:received` | Nova mensagem recebida |
| `message:ack` | Confirmação de entrega/leitura de mensagem |
| `call:offer` | Chamada recebida (offer) |
| `call:accept` | Chamada aceita |
| `call:terminate` | Chamada encerrada |
| `presence:update` | Mudança de presença de um contato |
| `contact:update` | Atualização de contato |

O formato do evento é `data: { "type": "event_name", "payload": { ... } }\n\n`.

---

## 📱 Gestão de Sessões

### Criação

```bash
curl -X POST http://localhost:8080/api/sessions \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"name": "Minha Instância"}'
```

Retorna o `sid` (slug do nome), `id` (UUID) e a `apiKey` gerada automaticamente. O QR Code para pareamento é emitido via SSE.

### Logout e Re-pareamento

```bash
# Desconecta do WhatsApp (mantém a sessão salva)
curl -X POST http://localhost:8080/api/sessions/{sid}/logout -u admin:admin123

# Gera novo QR Code para pareamento
curl -X POST http://localhost:8080/api/sessions/{sid}/pair -u admin:admin123
```

### Atualização

```bash
# Renomear sessão
curl -X PUT http://localhost:8080/api/sessions/{sid}/name \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"name": "Novo Nome"}'

# Atualizar chave de API
curl -X PUT http://localhost:8080/api/sessions/{sid}/apikey \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"apiKey": "nova-chave"}'
```

### Exclusão

```bash
curl -X DELETE http://localhost:8080/api/sessions/{sid} -u admin:admin123
```

> **Nota:** As sessões são restauradas automaticamente ao reiniciar o servidor, desde que o banco PostgreSQL esteja acessível.

---

## 💬 Mensagens

### Dispatcher Genérico

O endpoint `POST /api/sessions/{sid}/messages` aceita um campo `type` para escolher automaticamente o formato:

```json
{
  "to": "5531999999999",
  "type": "text",
  "content": { "text": "Olá!" }
}
```

Tipos suportados: `text`, `media`, `location`, `contact`, `list`, `list-interactive`, `poll`.

### Mídia

Envio de imagens, vídeos, áudios, documentos e stickers:

```bash
curl -X POST http://localhost:8080/api/sessions/{sid}/messages/media \
  -u admin:admin123 \
  -F "to=5531999999999" \
  -F "file=@imagem.jpg" \
  -F "caption=Legenda opcional"
```

### Localização e Contato

```json
// Localização
{
  "to": "5531999999999",
  "latitude": -19.9167,
  "longitude": -43.9345,
  "title": "Praça Sete"
}

// Contato
{
  "to": "5531999999999",
  "contactJid": "5531888888888@s.whatsapp.net",
  "contactName": "Maria"
}
```

### Reações, Edição e Revogação

```bash
# Reagir a uma mensagem
curl -X POST http://localhost:8080/api/sessions/{sid}/messages/react \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"to": "5531999999999", "messageId": "ID_DA_MSG", "emoji": "👍"}'

# Editar mensagem
curl -X POST http://localhost:8080/api/sessions/{sid}/messages/edit \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"to": "5531999999999", "messageId": "ID_DA_MSG", "text": "Novo texto"}'

# Revogar/excluir mensagem
curl -X DELETE "http://localhost:8080/api/sessions/{sid}/messages/ID_DA_MSG?to=5531999999999" \
  -u admin:admin123
```

### Marcação de Leitura e Download de Mídia

```bash
# Marcar como lida
curl -X POST http://localhost:8080/api/sessions/{sid}/messages/{id}/read \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"to": "5531999999999"}'

# Baixar mídia de uma mensagem
curl -u admin:admin123 \
  http://localhost:8080/api/sessions/{sid}/media/ID_DA_MSG \
  -o arquivo_baixado.bin
```

---

## 👥 Grupos

### Criação e Atualização

```bash
# Criar grupo
curl -X POST http://localhost:8080/api/sessions/{sid}/groups \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"subject": "Meu Grupo", "participants": ["5531999999999@s.whatsapp.net"]}'

# Atualizar nome, tópico, locked (só admins enviam msg) e announce (só admins falam)
curl -X PUT http://localhost:8080/api/sessions/{sid}/groups/{gid} \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"subject": "Novo Nome", "topic": "Descrição do grupo", "locked": false, "announce": false}'
```

### Participantes

```bash
# Adicionar/remover participantes
curl -X POST http://localhost:8080/api/sessions/{sid}/groups/{gid}/participants \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"action": "add", "participants": ["5531999999999@s.whatsapp.net"]}'
# action: "add" | "remove"

# Promover/demover (admin/normal)
curl -X PUT http://localhost:8080/api/sessions/{sid}/groups/{gid}/participants/{jid} \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"action": "promote"}'  # "promote" | "demote"
```

### Links de Convite

```bash
# Obter link de convite
curl -u admin:admin123 \
  http://localhost:8080/api/sessions/{sid}/groups/{gid}/invite

# Revogar link de convite
curl -X POST http://localhost:8080/api/sessions/{sid}/groups/{gid}/invite/revoke \
  -u admin:admin123

# Entrar em grupo via link
curl -X POST http://localhost:8080/api/sessions/{sid}/groups/join \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"inviteCode": "CODIGO_DO_CONVITE"}'
```

### Pedidos de Entrada

```bash
# Listar pedidos pendentes
curl -u admin:admin123 \
  http://localhost:8080/api/sessions/{sid}/groups/{gid}/requests

# Aprovar/recusar pedidos
curl -X POST http://localhost:8080/api/sessions/{sid}/groups/{gid}/requests \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"action": "approve", "participants": ["5531999999999@s.whatsapp.net"]}'
# action: "approve" | "reject"
```

### Foto do Grupo

```bash
# Definir foto do grupo
curl -X PUT http://localhost:8080/api/sessions/{sid}/groups/{gid}/photo \
  -u admin:admin123 \
  -F "file=@foto.jpg"

# Remover foto do grupo
curl -X DELETE http://localhost:8080/api/sessions/{sid}/groups/{gid}/photo \
  -u admin:admin123
```

---

## 👤 Usuários, Contatos e Presença

```bash
# Buscar informações de um ou mais usuários
curl -u admin:admin123 \
  "http://localhost:8080/api/sessions/{sid}/users?jid[]=5531999999999@s.whatsapp.net"

# Verificar se números estão no WhatsApp
curl -X POST http://localhost:8080/api/sessions/{sid}/users/check \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"phones": ["5531999999999", "5531888888888"]}'

# Assinar presença de um contato (recebe updates via SSE)
curl -u admin:admin123 \
  http://localhost:8080/api/sessions/{sid}/users/{jid}/presence

# Buscar avatar do contato
curl -u admin:admin123 \
  http://localhost:8080/api/sessions/{sid}/contacts/{jid}/avatar
```

---

## 🖼️ Perfil

### Nome, Status e Foto

```bash
# Atualizar nome de perfil
curl -X PUT http://localhost:8080/api/sessions/{sid}/profile/name \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"name": "Meu Nome"}'

# Atualizar status (recado)
curl -X PUT http://localhost:8080/api/sessions/{sid}/profile/status \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"status": "Disponível"}'

# Atualizar foto de perfil
curl -X PUT http://localhost:8080/api/sessions/{sid}/profile/photo \
  -u admin:admin123 \
  -F "file=@foto_perfil.jpg"
```

### Privacidade

```bash
# Ler configurações de privacidade
curl -u admin:admin123 \
  http://localhost:8080/api/sessions/{sid}/profile/privacy

# Atualizar configurações de privacidade
curl -X PUT http://localhost:8080/api/sessions/{sid}/profile/privacy \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{
    "profilePhoto": "contacts",   # "all" | "contacts" | "none"
    "status": "contacts",
    "readReceipts": "all",
    "groupAdd": "contacts",
    "lastSeen": "all"
  }'
```

---

## 🚫 Blocklist

```bash
# Listar contatos bloqueados
curl -u admin:admin123 \
  http://localhost:8080/api/sessions/{sid}/blocklist

# Adicionar/remover da blocklist
curl -X POST http://localhost:8080/api/sessions/{sid}/blocklist \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"action": "block", "jid": "5531999999999@s.whatsapp.net"}'
# action: "block" | "unblock"
```

---

## 📢 Newsletters (Canais)

```bash
# Criar newsletter/canal
curl -X POST http://localhost:8080/api/sessions/{sid}/newsletters \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"name": "Meu Canal", "description": "Descrição do canal"}'

# Obter informações da newsletter
curl -u admin:admin123 \
  http://localhost:8080/api/sessions/{sid}/newsletters/{jid}

# Deixar de seguir newsletter
curl -X DELETE http://localhost:8080/api/sessions/{sid}/newsletters/{jid} \
  -u admin:admin123

# Alternar mute (notificações silenciosas)
curl -X POST http://localhost:8080/api/sessions/{sid}/newsletters/{jid}/mute \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"muted": true}'
```

---

## 📞 Chamadas de Voz (WebRTC)

### Fluxo Completo

1. **Iniciar chamada** — `POST /api/sessions/{sid}/calls` com `{"to": "5531999999999"}` retorna um `callId`.
2. **Iniciar WebRTC** — `POST /api/sessions/{sid}/calls/{callId}/webrtc` inicia a negociação ICE/DTLS com o servidor.
3. **Troca de SDP** — O servidor negocia o áudio (MLow → PCM) e estabelece o fluxo SRTP com o WhatsApp.
4. **Encerrar** — `DELETE /api/sessions/{sid}/calls/{callId}` finaliza a chamada.

### Chamadas Recebidas

Quando uma chamada é recebida, um evento `call:offer` é emitido via SSE. Para responder:

```bash
# Aceitar chamada
curl -X POST http://localhost:8080/api/sessions/{sid}/calls/{callId}/accept \
  -u admin:admin123

# Rejeitar chamada
curl -X POST http://localhost:8080/api/sessions/{sid}/calls/{callId}/reject \
  -u admin:admin123
```

### Histórico

```bash
curl -u admin:admin123 \
  http://localhost:8080/api/sessions/{sid}/history
```

### Limite de Chamadas Simultâneas

O servidor limita o número de chamadas simultâneas por sessão conforme a flag `-max-calls-per-session` (padrão: 8). Quando o limite é atingido, novas chamadas são recusadas com erro `429 Too Many Requests`.

---

## �🔐 Segurança e Boas Práticas

- **Credenciais:** Mantenha o seu arquivo `.env` seguro. Ele contém as chaves e dados de acesso à sua infraestrutura de sessões do WhatsApp.
- **Basic Auth:** A API é protegida por HTTP Basic Auth. Configure `AUTH_USERNAME`/`AUTH_PASSWORD` no `.env`.
- **Rede Local:** Recomenda-se rodar o sistema em rede local confiável (LAN) ou protegida por proxy reverso com HTTPS.

---

## � Changelog
### v1.0.7 (2026-07-16)
- 🌐 **Webhooks**: Sistema completo de webhooks para integração externa
  - Novo endpoint `GET /api/sessions/{sid}/webhooks` — Listar webhooks de uma sessão
  - Novo endpoint `POST /api/sessions/{sid}/webhooks` — Criar novo webhook
  - Novo endpoint `PUT /api/sessions/{sid}/webhooks/{wid}` — Atualizar webhook
  - Novo endpoint `DELETE /api/sessions/{sid}/webhooks/{wid}` — Remover webhook
  - Novo endpoint `POST /api/sessions/{sid}/webhooks/{wid}/test` — Testar disparo do webhook
  - Disparo assíncrono com HMAC-SHA256, retry com backoff (3 tentativas)
  - Eventos filtráveis: message, message-receipt, presence, call-status, incoming, call-ended
  - Interface de gerenciamento no frontend (adicionar, editar, excluir, testar)
  - Documentação completa no Swagger
- 💳 **Menu PIX**: Separado do menu Chamadas, agora com aba dedicada no frontend
  - Nova rota PIX separada das chamadas no backend
  - Menu "PIX" independente na interface
- 📖 **Swagger**: Tags organizadas com ordem definida para melhor navegação

### v1.0.6 (2026-07-15)
- 🧩 **Fallback para WhatsApp Web**: Adicionado parâmetro `asText` nos endpoints de listas interativas
  - Novo campo `asText` (booleano) em `POST /api/sessions/{sid}/messages/list` e `POST /api/sessions/{sid}/messages/list-interactive`
  - Quando `true`, envia as opções como texto plano formatado com numeração, evitando o erro "Could not load message" no WhatsApp Web
  - Ideal para destinatários que não renderizam botões interativos corretamente
- 📖 **Swagger**: Documentado o parâmetro `asText` nos schemas e exemplos dos endpoints de lista

### v1.0.5 (2026-07-15)
- 💳 **Pagamentos PIX**: Implementação completa de geração de QR Code PIX estático (EMV 2024)
  - Novo endpoint `POST /api/sessions/{sid}/messages/pix` — Envia PIX via WhatsApp (texto + QR Code)
  - Novo endpoint `POST /api/pix/generate` — Gera QR Code PIX para pré-visualização
  - Novo endpoint `POST /api/pix/validate` — Valida chave PIX com dígitos verificadores (CPF/CNPJ)
  - Suporte a todos os tipos de chave: CPF, CNPJ, telefone, e-mail e chave aleatória
  - Geração de QR Code em PNG usando `github.com/skip2/go-qrcode`
  - Componente React `PixSender` com pré-visualização do QR Code e cópia do código copia e cola
- 📖 **Swagger**: Documentados todos os endpoints PIX com schemas (`PixResponse`, `PixGenerateResponse`, `PixValidateResponse`)
- 🔧 **Swagger UI**: Adicionado servidor `http://localhost:8081` para ambiente de desenvolvimento (além do `:8080` para Docker)
- 🐳 **Imagem Docker**: Otimizado o processo de build (binário Go pré-compilado para evitar timeout de CGO no Docker Desktop)
- 🚢 **Docker Hub**: Imagem `v1.0.5` publicada com tags `v1.0.5` e `latest`

### v1.0.4 (2026-07-13)
- 🐳 **Docker**: Substituído `EXTERNAL_IP` por `network_mode: host` para resolução definitiva do WebRTC em Linux
- ⚡ **Scripts de inicialização**: Unificado `start.bat` (compila frontend + servidor Go na porta 8080); novo `start-dev.bat` para desenvolvimento com dois terminais (backend :8081 + frontend :5173 com HMR)
- 🔗 **AppShell**: Link do Swagger alterado para caminho relativo (`/swagger`)
- 🗄️ **Configuração**: DB e Redis apontando para servidor externo (`192.168.50.2`)
- 🚢 **Imagem Docker**: Atualizada para `v1.0.4` e publicada no Docker Hub

### v1.0.3 (2026-07-13)
- 🔧 **Correções de API**: Adicionado `context.Context` em chamadas da biblioteca `whatsmeow` (`GetJoinedGroups`, `MarkRead`, `GetBlocklist`, `CreateNewsletter` etc.)
- 🐳 **Imagem Docker**: Atualizada para `v1.0.3` e publicada no Docker Hub

### v1.0.2 (2026-07-13)
- 📖 **Documentação completa da API**: Adicionadas seções detalhadas para todos os endpoints (sessões, mensagens, grupos, contatos, perfil, blocklist, newsletters, chamadas WebRTC, eventos SSE)
- 📋 **Tabela de endpoints**: Substituída por tabela completa com método HTTP, rota e descrição de todos os 60+ endpoints
- 🚩 **Flags do servidor**: Documentadas todas as flags de linha de comando (`-addr`, `-static`, `-debug`, `-max-calls-per-session`)

### v1.0.1 (2026-07-13)
- 🐳 **Correção Docker**: Adicionado suporte a `EXTERNAL_IP` para ICE candidates do WebRTC, resolvendo problemas de microfone/áudio em containers Docker (Windows/Mac)
- 🎤 **Correção de eco**: Removida linha `captureNode.connect(ctx.destination)` que roteava o microfone para os alto-falantes
- 📖 **Documentação**: Adicionada seção Docker e variável `EXTERNAL_IP`

### v1.0.0
- 🎉 Lançamento inicial

---

## �🙌 Créditos

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
