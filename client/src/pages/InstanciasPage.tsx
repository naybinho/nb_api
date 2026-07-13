import { useCallback, useState } from "react";
import { Check, Copy, Eye, EyeOff, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useSessions } from "@/stores/sessions";
import { updateSessionAPIKey, updateSessionName } from "@/services/sessions";
import type { SessionInfo, SessionState } from "@/types/session";

const statusLabel: Record<SessionState, string> = {
  open: "Connected",
  qr: "Scan QR",
  connecting: "Connecting…",
  logged_out: "Disconnected",
};

const statusVariant: Record<SessionState, "success" | "secondary" | "muted" | "destructive"> = {
  open: "success",
  qr: "secondary",
  connecting: "muted",
  logged_out: "destructive",
};

export const InstanciasPage = () => {
  const sessions = useSessions((s) => s.sessions);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [visibleKeys, setVisibleKeys] = useState<Set<string>>(new Set());
  const [editingName, setEditingName] = useState<string | null>(null);
  const [editValue, setEditValue] = useState("");

  const copyApiKey = useCallback(async (apiKey: string, sessionId: string) => {
    try {
      await navigator.clipboard.writeText(apiKey);
      setCopiedId(sessionId);
      toast.success("API Key copiada!");
      setTimeout(() => setCopiedId(null), 2000);
    } catch {
      toast.error("Falha ao copiar");
    }
  }, []);

  const toggleVisible = (id: string) => {
    setVisibleKeys((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const regenerateKey = async (session: SessionInfo) => {
    try {
      const bytes = new Uint8Array(20);
      crypto.getRandomValues(bytes);
      const hex = Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
      const newKey = `wac_${hex}`;
      await updateSessionAPIKey(session.id, newKey);
      toast.success("API Key regenerada!");
    } catch (e) {
      toast.error((e as Error).message);
    }
  };

  const startEditName = (session: SessionInfo) => {
    setEditingName(session.id);
    setEditValue(session.name);
  };

  const saveName = async (id: string) => {
    try {
      await updateSessionName(id, editValue.trim());
      setEditingName(null);
      toast.success("Nome atualizado!");
    } catch (e) {
      toast.error((e as Error).message);
    }
  };

  if (sessions.length === 0) {
    return (
      <div className="mx-auto max-w-3xl space-y-6">
        <h2 className="text-lg font-semibold">Instancias</h2>
        <p className="text-sm text-muted-foreground">Nenhuma instancia criada. Crie uma pelo menu ao lado.</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <h2 className="text-lg font-semibold">Instancias</h2>
      <div className="grid grid-cols-1 gap-4">
        {sessions.map((s) => (
          <Card key={s.id}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <div className="flex items-center gap-2">
                {editingName === s.id ? (
                  <div className="flex items-center gap-2">
                    <input
                      type="text"
                      value={editValue}
                      onChange={(e) => setEditValue(e.target.value)}
                      className="rounded-md border border-input bg-transparent px-2 py-1 text-sm font-semibold"
                      autoFocus
                      onKeyDown={(e) => {
                        if (e.key === "Enter") saveName(s.id);
                        if (e.key === "Escape") setEditingName(null);
                      }}
                    />
                    <Button size="sm" variant="ghost" onClick={() => saveName(s.id)}>
                      Salvar
                    </Button>
                  </div>
                ) : (
                  <CardTitle
                    className="cursor-pointer text-base hover:underline"
                    onClick={() => startEditName(s)}
                  >
                    {s.name}
                  </CardTitle>
                )}
                <Badge variant={statusVariant[s.state]}>{statusLabel[s.state]}</Badge>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-2 text-sm">
                {s.jid && (
                  <div className="flex items-center gap-2">
                    <span className="text-muted-foreground">JID:</span>
                    <span className="font-mono text-xs">{s.jid}</span>
                  </div>
                )}
                <div className="flex items-center gap-2">
                  <span className="text-muted-foreground">API Key:</span>
                  <code className="flex-1 rounded bg-muted px-2 py-0.5 font-mono text-xs">
                    {visibleKeys.has(s.id) ? s.apiKey || "-" : `${(s.apiKey || "").slice(0, 12) || "---"}...`}
                  </code>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={() => toggleVisible(s.id)}
                    title={visibleKeys.has(s.id) ? "Ocultar" : "Mostrar"}
                  >
                    {visibleKeys.has(s.id) ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={() => copyApiKey(s.apiKey, s.id)}
                    title="Copiar"
                  >
                    {copiedId === s.id ? <Check className="h-3.5 w-3.5 text-primary" /> : <Copy className="h-3.5 w-3.5" />}
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={() => regenerateKey(s)}
                    title="Regenerar"
                  >
                    <RefreshCw className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
};
