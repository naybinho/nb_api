import { useState, useCallback } from "react";
import { Copy, Check, Loader2, QrCode, Send } from "lucide-react";
import { QRCodeSVG } from "qrcode.react";
import { toast } from "sonner";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { sendPix, generatePix, type PixKeyType, type SendPixParams } from "@/services/pix";

// ── Helpers ──────────────────────────────────────────────────────────────────

const KEY_TYPES: { value: PixKeyType; label: string }[] = [
  { value: "cpf", label: "CPF" },
  { value: "cnpj", label: "CNPJ" },
  { value: "phone", label: "Telefone" },
  { value: "email", label: "E-mail" },
  { value: "random", label: "Chave Aleatória" },
];

const selectClass =
  "h-9 w-full rounded-md border border-input bg-transparent px-3 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring";

// ── Component ────────────────────────────────────────────────────────────────

export const PixSender = ({ sid }: { sid: string }) => {
  const [phone, setPhone] = useState("");
  const [pixKey, setPixKey] = useState("");
  const [keyType, setKeyType] = useState<PixKeyType>("cpf");
  const [merchantName, setMerchantName] = useState("");
  const [merchantCity, setMerchantCity] = useState("");
  const [amount, setAmount] = useState("");
  const [description, setDescription] = useState("");
  const [sendQRCode, setSendQRCode] = useState(true);
  const [copied, setCopied] = useState(false);
  const [preview, setPreview] = useState<{ brcode: string; qrCode?: string } | null>(null);
  const [isPending, setIsPending] = useState(false);

  // Generate preview (client-side)
  const handlePreview = useCallback(async () => {
    if (!pixKey.trim()) {
      toast.error("Informe a chave PIX");
      return;
    }

    try {
      const result = await generatePix({
        pixKey: pixKey.trim(),
        keyType,
        merchantName: merchantName.trim() || "Pagamento",
        merchantCity: merchantCity.trim() || "Cidade",
        amount: amount ? parseFloat(amount) : undefined,
        description: description.trim(),
      });
      setPreview(result);
    } catch (e) {
      toast.error("Erro ao gerar PIX: " + (e as Error).message);
    }
  }, [pixKey, keyType, merchantName, merchantCity, amount, description]);

  // Send PIX via WhatsApp
  const handleSend = useCallback(async () => {
    if (!phone.trim()) {
      toast.error("Informe o telefone do destinatário");
      return;
    }
    if (!pixKey.trim()) {
      toast.error("Informe a chave PIX");
      return;
    }

    setIsPending(true);
    try {
      const params: SendPixParams = {
        to: phone.trim(),
        pixKey: pixKey.trim(),
        keyType,
        merchantName: merchantName.trim() || "Pagamento",
        merchantCity: merchantCity.trim() || "Cidade",
        amount: amount ? parseFloat(amount) : undefined,
        description: description.trim(),
        sendQRCode,
      };

      const result = await sendPix(sid, params);
      setPreview({ brcode: result.brcode });
      toast.success("PIX enviado com sucesso!");
      // Reset form
      setPhone("");
      setPixKey("");
      setAmount("");
      setDescription("");
      setPreview(null);
    } catch (e) {
      toast.error("Erro ao enviar PIX: " + (e as Error).message);
    } finally {
      setIsPending(false);
    }
  }, [phone, pixKey, keyType, merchantName, merchantCity, amount, description, sendQRCode, sid]);

  // Copy brcode to clipboard
  const copyBrcode = useCallback(async () => {
    if (!preview?.brcode) return;
    try {
      await navigator.clipboard.writeText(preview.brcode);
      setCopied(true);
      toast.success("Código PIX copiado!");
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error("Falha ao copiar");
    }
  }, [preview]);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <QrCode className="h-5 w-5" />
          Enviar PIX
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Destinatário */}
        <div className="space-y-1.5">
          <Label htmlFor="pix-phone">Telefone do destinatário</Label>
          <Input
            id="pix-phone"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            placeholder="+55 11 99999 9999"
            inputMode="tel"
          />
        </div>

        {/* Chave PIX */}
        <div className="space-y-1.5">
          <Label htmlFor="pix-key">Chave PIX</Label>
          <div className="flex gap-2">
            <select
              value={keyType}
              onChange={(e) => setKeyType(e.target.value as PixKeyType)}
              className={selectClass}
              style={{ width: "auto", minWidth: "120px" }}
            >
              {KEY_TYPES.map((kt) => (
                <option key={kt.value} value={kt.value}>
                  {kt.label}
                </option>
              ))}
            </select>
            <Input
              id="pix-key"
              value={pixKey}
              onChange={(e) => setPixKey(e.target.value)}
              placeholder="Ex: 123.456.789-00"
              className="flex-1"
            />
          </div>
        </div>

        {/* Beneficiário */}
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div className="space-y-1.5">
            <Label htmlFor="pix-name">Nome do beneficiário</Label>
            <Input
              id="pix-name"
              value={merchantName}
              onChange={(e) => setMerchantName(e.target.value)}
              placeholder="Nome do recebedor"
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="pix-city">Cidade</Label>
            <Input
              id="pix-city"
              value={merchantCity}
              onChange={(e) => setMerchantCity(e.target.value)}
              placeholder="Cidade"
            />
          </div>
        </div>

        {/* Valor e Descrição */}
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div className="space-y-1.5">
            <Label htmlFor="pix-amount">Valor (opcional)</Label>
            <div className="relative">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">R$</span>
              <Input
                id="pix-amount"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                placeholder="0,00"
                inputMode="decimal"
                className="pl-10"
              />
            </div>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="pix-desc">Descrição (opcional)</Label>
            <Input
              id="pix-desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Ex: Pagamento do serviço"
            />
          </div>
        </div>

        {/* QR Code toggle */}
        <div className="flex items-center gap-2">
          <input
            id="pix-send-qr"
            type="checkbox"
            checked={sendQRCode}
            onChange={(e) => setSendQRCode(e.target.checked)}
            className="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary"
          />
          <Label htmlFor="pix-send-qr">Enviar QR Code como imagem</Label>
        </div>

        {/* Ações */}
        <div className="flex flex-wrap gap-2">
          <Button
            variant="outline"
            onClick={handlePreview}
            disabled={!pixKey.trim()}
          >
            <QrCode className="h-4 w-4" />
            Visualizar QR Code
          </Button>
          <Button
            onClick={handleSend}
            disabled={isPending || !phone.trim() || !pixKey.trim()}
          >
            {isPending ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Send className="h-4 w-4" />
            )}
            {isPending ? "Enviando..." : "Enviar PIX"}
          </Button>
        </div>

        {/* Preview do QR Code */}
        {preview && (
          <div className="space-y-3 rounded-lg border bg-muted/30 p-4">
            <p className="text-sm font-medium text-muted-foreground">Prévia do PIX</p>

            <div className="flex flex-col items-center gap-3 sm:flex-row sm:items-start">
              {/* QR Code */}
              <div className="flex-shrink-0 rounded-lg bg-white p-3 shadow-sm">
                {preview.qrCode ? (
                  <img
                    src={preview.qrCode}
                    alt="PIX QR Code"
                    className="h-40 w-40"
                  />
                ) : (
                  <QRCodeSVG value={preview.brcode} size={160} level="M" />
                )}
              </div>

              {/* Brcode */}
              <div className="flex min-w-0 flex-1 flex-col gap-2">
                <Label className="text-xs text-muted-foreground">Código Pix Copia e Cola</Label>
                <div className="relative">
                  <code className="block max-h-20 overflow-y-auto break-all rounded-md border bg-background p-2 text-xs font-mono">
                    {preview.brcode}
                  </code>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="absolute right-1 top-1 h-6 w-6"
                    onClick={copyBrcode}
                    title="Copiar código"
                  >
                    {copied ? (
                      <Check className="h-3.5 w-3.5 text-primary" />
                    ) : (
                      <Copy className="h-3.5 w-3.5" />
                    )}
                  </Button>
                </div>
                {merchantName && (
                  <p className="text-xs text-muted-foreground">
                    Beneficiário: <span className="font-medium">{merchantName}</span>
                  </p>
                )}
                {amount && (
                  <p className="text-xs text-muted-foreground">
                    Valor: <span className="font-medium">R$ {parseFloat(amount).toFixed(2)}</span>
                  </p>
                )}
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
};
