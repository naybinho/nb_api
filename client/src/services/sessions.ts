import { apiGet, apiPost, apiPut, apiDelete } from "@/lib/api";
import { useSessions } from "@/stores/sessions";
import type { SessionInfo } from "@/types/session";

const sessionAPIKey = (id: string): string | undefined =>
  useSessions.getState().sessions.find((s) => s.id === id)?.apiKey;

export const listSessions = () =>
  apiGet<{ sessions: SessionInfo[] }>("/api/sessions").then((r) => r.sessions ?? []);

export const createSession = (name: string, apiKey?: string) =>
  apiPost<{ id: string; apiKey: string }>("/api/sessions", { name, apiKey });

export const deleteSession = (id: string) =>
  apiDelete(`/api/sessions/${id}`, sessionAPIKey(id));

export const updateSessionAPIKey = (id: string, apiKey: string) =>
  apiPut<{ apiKey: string }>(`/api/sessions/${id}/apikey`, { apiKey }, sessionAPIKey(id));

export const updateSessionName = (id: string, name: string) =>
  apiPut<void>(`/api/sessions/${id}/name`, { name }, sessionAPIKey(id));

export const logoutSession = (id: string) =>
  apiPost<void>(`/api/sessions/${id}/logout`, {}, sessionAPIKey(id));

export const pairSession = (id: string) =>
  apiPost<void>(`/api/sessions/${id}/pair`, {}, sessionAPIKey(id));
