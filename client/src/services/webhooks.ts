import { apiGet, apiPost, apiPut, apiDelete } from "@/lib/api";

export interface Webhook {
  id: string;
  sessionId: string;
  url: string;
  events: string;
  enabled: boolean;
  secret: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateWebhookParams {
  url: string;
  events?: string;
  enabled?: boolean;
  secret?: string;
}

export interface UpdateWebhookParams {
  url?: string;
  events?: string;
  enabled?: boolean;
  secret?: string;
}

export const listWebhooks = (sid: string) =>
  apiGet<{ webhooks: Webhook[] }>(`/api/sessions/${sid}/webhooks`)
    .then((r) => r.webhooks ?? []);

export const createWebhook = (sid: string, params: CreateWebhookParams) =>
  apiPost<Webhook>(`/api/sessions/${sid}/webhooks`, params);

export const updateWebhook = (sid: string, wid: string, params: UpdateWebhookParams) =>
  apiPut<Webhook>(`/api/sessions/${sid}/webhooks/${wid}`, params);

export const deleteWebhook = (sid: string, wid: string) =>
  apiDelete(`/api/sessions/${sid}/webhooks/${wid}`);

export const testWebhook = (sid: string, wid: string) =>
  apiPost<{ status: string; message: string }>(`/api/sessions/${sid}/webhooks/${wid}/test`, {});