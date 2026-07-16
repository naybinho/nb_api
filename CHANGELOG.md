# Changelog

Todas as alterações relevantes deste projeto serão documentadas aqui.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/),
e este projeto segue [Semantic Versioning](https://semver.org/lang/pt-BR/).

---

## [v1.0.8] - 2026-07-16

### ✨ Novas Funcionalidades

- **Gravação de Chamadas:** Agora é possível gravar chamadas de voz do WhatsApp com um clique. O áudio do microfone e do peer é capturado em PCM, mixado e convertido para WAV, sendo enviado automaticamente para armazenamento **S3-compatible** (MinIO, AWS S3, etc.) ao final da chamada.
- **Histórico Persistente em PostgreSQL:** O histórico de chamadas agora é salvo permanentemente no banco de dados PostgreSQL, nunca sendo perdido mesmo após reiniciar o servidor.
- **Player de Gravação no Frontend:** O painel de histórico exibe o status da gravação (gravada/não gravada) e um botão de play para ouvir a gravação diretamente no navegador.

### 🔧 Melhorias

- **Armazenamento S3 Configurável:** Suporte a qualquer serviço compatível com S3 via variáveis de ambiente (`S3_ENDPOINT`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_BUCKET`, `S3_REGION`, `S3_SSL`).
- **Ícone de Status no Histórico:** Indicador visual de chamadas gravadas com acesso rápido à gravação.
- **Fallback Inteligente:** Se o S3 não estiver configurado, a gravação é desabilitada automaticamente sem quebrar o fluxo de chamadas.

### 🧹 Limpeza Automática de Gravações

- **`RECORDING_RETENTION_DAYS`:** Nova variável de ambiente para deletar automaticamente gravações antigas do S3 após um período configurável (10, 15, 20, 30, 60 ou 90 dias). O scheduler de limpeza roda a cada 6 horas em background. Quando uma gravação expira, o arquivo é removido do S3 e a URL é limpa do banco de dados (o histórico da chamada permanece).
- **Cleanup Seguro:** Erros de deleção no S3 não interrompem o scheduler — as gravações com falha serão tentadas novamente na próxima execução.

### 📦 Dependências

- Adicionado: `github.com/minio/minio-go/v7 v7.2.1` (cliente S3)

---

## [v1.0.7] - 2026-06-15

### ✨ Novas Funcionalidades

- Implementação completa de newsletters (canais do WhatsApp)
- Envio de PIX com QR Code via WhatsApp
- Suporte a listas interativas dropdown
- Gerenciamento de grupos: aprovação de pedidos de entrada, foto do grupo
- Configurações de privacidade (ler/escrever)

### 🔧 Melhorias

- Validação de chaves PIX (CPF/CNPJ com dígitos verificadores)
- Normalização de números BR com/sem dígito 9
- Suporte a edição e revogação de mensagens

---

## [v1.0.6] - 2026-05-20

### ✨ Novas Funcionalidades

- Envio de reações a mensagens
- Enquetes interativas
- Marcação de mensagens como lidas
- Download de mídia das mensagens

### 🔧 Melhorias

- Otimizações no processamento de áudio VoIP (MLow codec)
- Melhor gestão de sessões simultâneas

---

## [v1.0.5] - 2026-04-10

### ✨ Novas Funcionalidades

- Chamadas de voz nativas via WebRTC
- Áudio bidirecional com codificação MLow
- Suporte a SRTP para criptografia de mídia
- Interface de chamadas no frontend (Dialer, CallCard)

### 🔧 Melhorias

- Proxy do Vite para desenvolvimento
- Integração com Redis para cache e mensageria

---

## [v1.0.4] - 2026-03-01

### ✨ Novas Funcionalidades

- Multi-sessões com QR Code
- Envio de mensagens de texto, mídia, localização e contatos
- Listas interativas (quick_reply)
- Gerenciamento de grupos (criar, atualizar, participantes)
- Contatos e perfil
- Blocklist
- Eventos em tempo real via SSE
- Documentação Swagger interativa
- Autenticação Basic Auth
- Limite de chamadas simultâneas

---

[v1.0.8]: https://github.com/naybinho/nb_api/releases/tag/v1.0.8
[v1.0.7]: https://github.com/naybinho/nb_api/releases/tag/v1.0.7
[v1.0.6]: https://github.com/naybinho/nb_api/releases/tag/v1.0.6
[v1.0.5]: https://github.com/naybinho/nb_api/releases/tag/v1.0.5
[v1.0.4]: https://github.com/naybinho/nb_api/releases/tag/v1.0.4
