import { getClientId } from "./client-id";
import { useSessions } from "@/stores/sessions";

const baseHeaders = (): Record<string, string> => {
  const headers: Record<string, string> = {
    "X-Client-Id": getClientId(),
    "Content-Type": "application/json",
  };
  const state = useSessions.getState();
  const active = state.sessions.find((s) => s.id === state.activeId);
  if (active?.apiKey) {
    headers["X-Api-Key"] = active.apiKey;
  }
  return headers;
};

export const apiGet = async <T>(path: string): Promise<T> => {
  const r = await fetch(path, { headers: baseHeaders() });
  if (!r.ok) throw new Error(`${path} ${r.status}`);
  return r.json() as Promise<T>;
};

export const apiPost = async <T>(path: string, body: unknown, apiKey?: string): Promise<T> => {
  const headers = baseHeaders();
  if (apiKey) headers["X-Api-Key"] = apiKey;
  const r = await fetch(path, { method: "POST", headers, body: JSON.stringify(body) });
  if (!r.ok) {
    const text = await r.text().catch(() => "");
    throw new Error(`${path} ${r.status} ${text}`);
  }
  const text = await r.text().catch(() => "");
  if (!text) return undefined as T;
  return JSON.parse(text) as T;
};

export const apiPut = async <T>(path: string, body: unknown, apiKey?: string): Promise<T> => {
  const headers = baseHeaders();
  if (apiKey) headers["X-Api-Key"] = apiKey;
  const r = await fetch(path, { method: "PUT", headers, body: JSON.stringify(body) });
  if (!r.ok) {
    const text = await r.text().catch(() => "");
    throw new Error(`${path} ${r.status} ${text}`);
  }
  const text = await r.text().catch(() => "");
  if (!text) return undefined as T;
  return JSON.parse(text) as T;
};

export const apiDelete = async (path: string, apiKey?: string): Promise<void> => {
  const headers = baseHeaders();
  if (apiKey) headers["X-Api-Key"] = apiKey;
  const r = await fetch(path, { method: "DELETE", headers });
  if (!r.ok) throw new Error(`${path} ${r.status}`);
};
