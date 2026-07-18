import { apiPost } from "@/lib/api";

export type PixKeyType = "cpf" | "cnpj" | "phone" | "email" | "random";

export interface SendPixParams {
  to: string;
  pixKey: string;
  keyType: PixKeyType;
  merchantName: string;
  merchantCity?: string;
  amount?: number;
  description?: string;
  sendQRCode?: boolean;
}

export interface SendPixResponse {
  messageId: string;
  timestamp: number;
  brcode: string;
}

export interface GeneratePixResponse {
  brcode: string;
  qrCode: string;
  generatedAt: number;
}

export interface ValidatePixResponse {
  valid: boolean;
  message: string;
}

/**
 * Send a PIX payment request to a WhatsApp contact.
 */
export const sendPix = (sid: string, params: SendPixParams) =>
  apiPost<SendPixResponse>(`/api/sessions/${sid}/messages/pix`, {
    ...params,
    amount: params.amount && params.amount > 0 ? params.amount : undefined,
    sendQRCode: params.sendQRCode ?? true,
    merchantCity: params.merchantCity || "Cidade",
  });

/**
 * Generate a PIX QR Code and brcode without sending.
 */
export const generatePix = (params: Omit<SendPixParams, "to">) =>
  apiPost<GeneratePixResponse>("/api/pix/generate", {
    pixKey: params.pixKey,
    keyType: params.keyType,
    merchantName: params.merchantName,
    merchantCity: params.merchantCity || "Cidade",
    amount: params.amount && params.amount > 0 ? params.amount : undefined,
    description: params.description,
  });

/**
 * Validate a PIX key.
 */
export const validatePixKey = (pixKey: string, keyType: PixKeyType) =>
  apiPost<ValidatePixResponse>("/api/pix/validate", { pixKey, keyType });
