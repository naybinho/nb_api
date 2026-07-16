package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"net/http"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// ──────────────────────────────────────────────────────────────────────────────
// PIX EMV Static QR Code generation
// ──────────────────────────────────────────────────────────────────────────────

// pixEmvField represents a TLV (Tag-Length-Value) field in the EMV format.
type pixEmvField struct {
	ID    string
	Value string
}

func (f pixEmvField) Encode() string {
	return fmt.Sprintf("%s%02d%s", f.ID, len(f.Value), f.Value)
}

// generatePixBRCode generates a PIX Static QR Code payload (copy-paste) following
// the EMV/BCB specification.
// Parameters:
//   - pixKey: a chave PIX (CPF, CNPJ, telefone, email, chave aleatória)
//   - merchantName: nome do recebedor (até 25 caracteres)
//   - merchantCity: cidade do recebedor (até 15 caracteres)
//   - amount: valor da transação (opcional, se <= 0 não é incluído)
//   - txid: identificador da transação (se vazio, usa "***")
//   - description: descrição (opcional, inserida no campo adicional)
func generatePixBRCode(pixKey, merchantName, merchantCity string, amount float64, txid, description string) string {
	const (
		gui = "BR.GOV.BCB.PIX"
	)

	// --- Merchant Account Information (ID 26) ---
	var merchantAcctInfo []pixEmvField
	merchantAcctInfo = append(merchantAcctInfo, pixEmvField{ID: "00", Value: gui})
	merchantAcctInfo = append(merchantAcctInfo, pixEmvField{ID: "01", Value: pixKey})
	if description != "" {
		// Truncate description to fit
		if len(description) > 40 {
			description = description[:40]
		}
		merchantAcctInfo = append(merchantAcctInfo, pixEmvField{ID: "02", Value: description})
	}

	var maiBuilder strings.Builder
	for _, f := range merchantAcctInfo {
		maiBuilder.WriteString(f.Encode())
	}
	merchantAccountInfo := pixEmvField{ID: "26", Value: maiBuilder.String()}

	// --- Additional Data Field Template (ID 62) ---
	if txid == "" {
		txid = "***"
	}
	if len(txid) > 25 {
		txid = txid[:25]
	}
	txidField := pixEmvField{ID: "05", Value: txid}

	var adfBuilder strings.Builder
	adfBuilder.WriteString(txidField.Encode())
	additionalData := pixEmvField{ID: "62", Value: adfBuilder.String()}

	// --- Build payload without CRC16 ---
	var payloadBuilder strings.Builder

	// 00 - Payload Format Indicator
	payloadBuilder.WriteString(pixEmvField{ID: "00", Value: "01"}.Encode())

	// 01 - Point of Initiation Method (12 = static QR Code)
	if amount > 0 {
		payloadBuilder.WriteString(pixEmvField{ID: "01", Value: "12"}.Encode())
	} else {
		payloadBuilder.WriteString(pixEmvField{ID: "01", Value: "12"}.Encode())
	}

	// 26 - Merchant Account Information
	payloadBuilder.WriteString(merchantAccountInfo.Encode())

	// 52 - Merchant Category Code
	payloadBuilder.WriteString(pixEmvField{ID: "52", Value: "0000"}.Encode())

	// 53 - Transaction Currency (986 = BRL)
	payloadBuilder.WriteString(pixEmvField{ID: "53", Value: "986"}.Encode())

	// 54 - Transaction Amount (optional)
	if amount > 0 {
		amountStr := fmt.Sprintf("%.2f", amount)
		payloadBuilder.WriteString(pixEmvField{ID: "54", Value: amountStr}.Encode())
	}

	// 58 - Country Code
	payloadBuilder.WriteString(pixEmvField{ID: "58", Value: "BR"}.Encode())

	// 59 - Merchant Name (max 25 chars)
	if len(merchantName) > 25 {
		merchantName = merchantName[:25]
	}
	payloadBuilder.WriteString(pixEmvField{ID: "59", Value: merchantName}.Encode())

	// 60 - Merchant City (max 15 chars)
	if len(merchantCity) > 15 {
		merchantCity = merchantCity[:15]
	}
	payloadBuilder.WriteString(pixEmvField{ID: "60", Value: merchantCity}.Encode())

	// 62 - Additional Data Field Template
	payloadBuilder.WriteString(additionalData.Encode())

	payloadWithoutCRC := payloadBuilder.String()

	// --- Calculate CRC16 (CRC-16-CCITT with polynomial 0x1021) ---
	crc := calculateCRC16(payloadWithoutCRC + "6304")
	crcHex := strings.ToUpper(fmt.Sprintf("%04X", crc))

	return payloadWithoutCRC + "63" + fmt.Sprintf("%02d", len(crcHex)) + crcHex
}

// calculateCRC16 computes CRC-16-CCITT with polynomial 0x1021.
func calculateCRC16(data string) uint16 {
	var crc uint16 = 0xFFFF
	for _, c := range []byte(data) {
		crc ^= uint16(c) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc & 0xFFFF
}

// generatePixQRCode generates a PNG QR Code image from a PIX brcode string.
// Returns the PNG bytes and any error.
func generatePixQRCode(brcode string, size int) ([]byte, error) {
	if size <= 0 {
		size = 512
	}
	qr, err := qrcode.New(brcode, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}
	qr.DisableBorder = true

	img := qr.Image(size)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode QR code PNG: %w", err)
	}
	return buf.Bytes(), nil
}

// ──────────────────────────────────────────────────────────────────────────────
// PIX Helper - format PIX key type
// ──────────────────────────────────────────────────────────────────────────────

func pixKeyTypeLabel(t string) string {
	switch strings.ToLower(t) {
	case "cpf":
		return "CPF"
	case "cnpj":
		return "CNPJ"
	case "phone", "telefone":
		return "Telefone"
	case "email":
		return "E-mail"
	case "random", "aleatoria":
		return "Chave Aleatória"
	default:
		return "Chave PIX"
	}
}

func formatPIXMessageText(pixKey, keyType, merchantName string, amount float64, description string, brcode string) string {
	var b strings.Builder

	b.WriteString("💳 *PAGAMENTO VIA PIX*\n\n")

	if merchantName != "" {
		b.WriteString(fmt.Sprintf("📌 *Beneficiário:* %s\n", merchantName))
	}

	if amount > 0 {
		b.WriteString(fmt.Sprintf("💰 *Valor:* R$ %.2f\n", amount))
	}

	b.WriteString(fmt.Sprintf("🔑 *%s:* `%s`\n\n", pixKeyTypeLabel(keyType), pixKey))

	if description != "" {
		b.WriteString(fmt.Sprintf("📝 *Descrição:* %s\n\n", description))
	}

	b.WriteString("📋 *Código Pix Copia e Cola:*\n")
	b.WriteString("```\n")
	b.WriteString(brcode)
	b.WriteString("\n```\n\n")
	b.WriteString("✅ Use o código acima ou escaneie o QR Code para realizar o pagamento.")

	return b.String()
}

// ──────────────────────────────────────────────────────────────────────────────
// Handler: Send PIX
// ──────────────────────────────────────────────────────────────────────────────

func (s *server) handleSendPix(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}
	if sess.client == nil || !sess.client.IsConnected() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "not connected"})
		return
	}

	var req struct {
		To              string  `json:"to"`              // destination phone or JID
		PixKey          string  `json:"pixKey"`          // chave PIX (CPF, CNPJ, phone, email, random)
		KeyType         string  `json:"keyType"`         // cpf, cnpj, phone, email, random
		MerchantName    string  `json:"merchantName"`    // nome do recebedor
		MerchantCity    string  `json:"merchantCity"`    // cidade do recebedor
		Amount          float64 `json:"amount"`          // valor (opcional)
		Description     string  `json:"description"`     // descrição (opcional)
		SendQRCode      bool    `json:"sendQRCode"`      // se true, envia QR Code como imagem
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Validation
	if strings.TrimSpace(req.To) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "to is required"})
		return
	}
	if strings.TrimSpace(req.PixKey) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "pixKey is required"})
		return
	}
	if strings.TrimSpace(req.MerchantName) == "" {
		req.MerchantName = "Pagamento"
	}
	if strings.TrimSpace(req.MerchantCity) == "" {
		req.MerchantCity = "Cidade"
	}

	// Generate PIX brcode
	brcode := generatePixBRCode(
		strings.TrimSpace(req.PixKey),
		strings.TrimSpace(req.MerchantName),
		strings.TrimSpace(req.MerchantCity),
		req.Amount,
		"***",
		strings.TrimSpace(req.Description),
	)

	toJID, err := s.resolvePhoneJID(r.Context(), sess, req.To)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JID"})
		return
	}

	if req.SendQRCode {
		// Generate QR Code image
		qrPNG, err := generatePixQRCode(brcode, 512)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate QR code: " + err.Error()})
			return
		}

		// Build caption
		caption := formatPIXMessageText(req.PixKey, req.KeyType, req.MerchantName, req.Amount, req.Description, brcode)

		// Upload image to WhatsApp
		uploadResp, err := sess.client.Upload(r.Context(), qrPNG, whatsmeow.MediaImage)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "upload failed: " + err.Error()})
			return
		}

		waMsg := &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				URL:           proto.String(uploadResp.URL),
				DirectPath:    proto.String(uploadResp.DirectPath),
				MediaKey:      uploadResp.MediaKey,
				FileEncSHA256: uploadResp.FileEncSHA256,
				FileSHA256:    uploadResp.FileSHA256,
				FileLength:    proto.Uint64(uploadResp.FileLength),
				Mimetype:      proto.String("image/png"),
				Caption:       proto.String(caption),
			},
		}

		resp, err := sess.client.SendMessage(r.Context(), toJID, waMsg)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"messageId": resp.ID,
			"timestamp": resp.Timestamp,
			"brcode":    brcode,
		})
	} else {
		// Send as text message only
		text := formatPIXMessageText(req.PixKey, req.KeyType, req.MerchantName, req.Amount, req.Description, brcode)
		msg := &waE2E.Message{Conversation: proto.String(text)}
		resp, err := sess.client.SendMessage(r.Context(), toJID, msg)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"messageId": resp.ID,
			"timestamp": resp.Timestamp,
			"brcode":    brcode,
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Handler: Generate PIX QR Code only (no send)
// ──────────────────────────────────────────────────────────────────────────────

func (s *server) handleGeneratePix(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PixKey       string  `json:"pixKey"`
		KeyType      string  `json:"keyType"`
		MerchantName string  `json:"merchantName"`
		MerchantCity string  `json:"merchantCity"`
		Amount       float64 `json:"amount"`
		Description  string  `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if strings.TrimSpace(req.PixKey) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "pixKey is required"})
		return
	}
	if strings.TrimSpace(req.MerchantName) == "" {
		req.MerchantName = "Pagamento"
	}
	if strings.TrimSpace(req.MerchantCity) == "" {
		req.MerchantCity = "Cidade"
	}

	brcode := generatePixBRCode(
		strings.TrimSpace(req.PixKey),
		strings.TrimSpace(req.MerchantName),
		strings.TrimSpace(req.MerchantCity),
		req.Amount,
		"***",
		strings.TrimSpace(req.Description),
	)

	qrPNG, err := generatePixQRCode(brcode, 512)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate QR code: " + err.Error()})
		return
	}

	qrB64 := base64.StdEncoding.EncodeToString(qrPNG)

	writeJSON(w, http.StatusOK, map[string]any{
		"brcode":    brcode,
		"qrCode":    "data:image/png;base64," + qrB64,
		"generatedAt": time.Now().UnixMilli(),
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Validate PIX Key endpoint
// ──────────────────────────────────────────────────────────────────────────────

func (s *server) handleValidatePixKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PixKey  string `json:"pixKey"`
		KeyType string `json:"keyType"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	valid := true
	var message string

	switch strings.ToLower(req.KeyType) {
	case "cpf":
		if !isValidCPF(req.PixKey) {
			valid = false
			message = "CPF inválido"
		}
	case "cnpj":
		if !isValidCNPJ(req.PixKey) {
			valid = false
			message = "CNPJ inválido"
		}
	case "phone", "telefone":
		if len(strings.TrimSpace(req.PixKey)) < 10 {
			valid = false
			message = "Telefone inválido"
		}
	case "email":
		if !strings.Contains(req.PixKey, "@") || !strings.Contains(req.PixKey, ".") {
			valid = false
			message = "E-mail inválido"
		}
	case "random", "aleatoria":
		if len(strings.TrimSpace(req.PixKey)) < 10 {
			valid = false
			message = "Chave aleatória inválida"
		}
	default:
		// Se não especificou o tipo, tenta validar como CPF ou CNPJ
		if !isValidCPF(req.PixKey) && !isValidCNPJ(req.PixKey) {
			valid = false
			message = "Chave PIX inválida"
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"valid":   valid,
		"message": message,
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// CPF/CNPJ validation
// ──────────────────────────────────────────────────────────────────────────────

func isValidCPF(cpf string) bool {
	// Remove non-numeric characters
	digits := onlyDigits(cpf)
	if len(digits) != 11 {
		return false
	}

	// Check all same digits
	allSame := true
	for i := 1; i < len(digits); i++ {
		if digits[i] != digits[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}

	// Validate first check digit
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(digits[i]-'0') * (10 - i)
	}
	rem := sum % 11
	digit1 := 0
	if rem >= 2 {
		digit1 = 11 - rem
	}
	if int(digits[9]-'0') != digit1 {
		return false
	}

	// Validate second check digit
	sum = 0
	for i := 0; i < 10; i++ {
		sum += int(digits[i]-'0') * (11 - i)
	}
	rem = sum % 11
	digit2 := 0
	if rem >= 2 {
		digit2 = 11 - rem
	}
	return int(digits[10]-'0') == digit2
}

func isValidCNPJ(cnpj string) bool {
	digits := onlyDigits(cnpj)
	if len(digits) != 14 {
		return false
	}

	// Check all same digits
	allSame := true
	for i := 1; i < len(digits); i++ {
		if digits[i] != digits[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}

	// First check digit
	weights1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := 0; i < 12; i++ {
		sum += int(digits[i]-'0') * weights1[i]
	}
	rem := sum % 11
	digit1 := 0
	if rem >= 2 {
		digit1 = 11 - rem
	}
	if int(digits[12]-'0') != digit1 {
		return false
	}

	// Second check digit
	weights2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum = 0
	for i := 0; i < 13; i++ {
		sum += int(digits[i]-'0') * weights2[i]
	}
	rem = sum % 11
	digit2 := 0
	if rem >= 2 {
		digit2 = 11 - rem
	}
	return int(digits[13]-'0') == digit2
}

func onlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Helper to truncate string with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
