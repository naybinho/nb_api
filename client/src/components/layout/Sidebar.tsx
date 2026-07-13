import { useCallback, useEffect, useState } from "react";
import { Check, Copy, Loader2, Plus, RefreshCw, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { setActiveSession, useSessions } from "@/stores/sessions";
import { createSession, deleteSession } from "@/services/sessions";
import type { SessionInfo, SessionState } from "@/types/session";

const dotClass: Record<SessionState, string> = {
  open: "bg-primary",
  qr: "bg-amber-500",
  connecting: "bg-muted-foreground/50",
  logged_out: "bg-destructive",
};

function generateAPIKey(): string {
  const bytes = new Uint8Array(20);
  crypto.getRandomValues(bytes);
  const hex = Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
  return `wac_${hex}`;
}

export const Sidebar = ({ onNavigate }: { onNavigate?: () => void }) => {
  const sessions = useSessions((s) => s.sessions);
  const activeId = useSessions((s) => s.activeId);
  const [creating, setCreating] = useState(false);
  const [toDelete, setToDelete] = useState<SessionInfo | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [newName, setNewName] = useState("WhatsApp");
  const [newApiKey, setNewApiKey] = useState("");
  const [copiedId, setCopiedId] = useState<string | null>(null);

  useEffect(() => {
    if (dialogOpen) {
      setNewName("WhatsApp");
      setNewApiKey(generateAPIKey());
    }
  }, [dialogOpen]);

  const onNew = async () => {
    setCreating(true);
    try {
      const { id } = await createSession(newName, newApiKey);
      setActiveSession(id);
      setDialogOpen(false);
      onNavigate?.();
    } catch (e) {
      toast.error((e as Error).message);
    } finally {
      setCreating(false);
    }
  };

  const remove = async (id: string) => {
    try {
      await deleteSession(id);
    } catch (e) {
      toast.error((e as Error).message);
    }
  };

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

  return (
    <TooltipProvider delayDuration={300}>
      <div className="flex h-full flex-col gap-2 p-3">
        <p className="px-2 pt-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">Accounts</p>
        <div className="flex-1 space-y-1 overflow-y-auto">
          {sessions.map((s) => (
            <div
              key={s.id}
              role="button"
              tabIndex={0}
              onClick={() => {
                setActiveSession(s.id);
                onNavigate?.();
              }}
              className={cn(
                "group flex cursor-pointer items-center gap-2 rounded-md px-2 py-2 text-sm",
                s.id === activeId ? "bg-accent text-accent-foreground" : "hover:bg-muted",
              )}
            >
              <span className={cn("h-2 w-2 shrink-0 rounded-full", dotClass[s.state])} />
              <div className="min-w-0 flex-1">
                <p className="truncate font-medium">{s.name}</p>
                {s.jid && <p className="truncate text-xs text-muted-foreground">{s.jid.split("@")[0]}</p>}
              </div>
              {s.apiKey && (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        copyApiKey(s.apiKey, s.id);
                      }}
                      className="text-muted-foreground opacity-0 transition-opacity hover:text-foreground group-hover:opacity-100"
                      aria-label={`Copy API Key de ${s.name}`}
                    >
                      {copiedId === s.id ? <Check className="h-3.5 w-3.5 text-primary" /> : <Copy className="h-3.5 w-3.5" />}
                    </button>
                  </TooltipTrigger>
                  <TooltipContent side="right">
                    <p className="font-mono text-xs">{s.apiKey.slice(0, 16)}...</p>
                  </TooltipContent>
                </Tooltip>
              )}
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  setToDelete(s);
                }}
                className="text-muted-foreground opacity-0 transition-opacity hover:text-destructive group-hover:opacity-100"
                aria-label={`Delete ${s.name}`}
              >
                <Trash2 className="h-4 w-4" />
              </button>
            </div>
          ))}
          {sessions.length === 0 && <p className="px-2 text-sm text-muted-foreground">No accounts yet.</p>}
        </div>
        <Button variant="outline" className="w-full" onClick={() => setDialogOpen(true)} disabled={creating}>
          <Plus className="h-4 w-4" />
          New session
        </Button>

        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogContent className="sm:max-w-md">
            <DialogHeader>
              <DialogTitle>Nova Instancia</DialogTitle>
              <DialogDescription>Escolha um nome e configure a API Key para esta instancia.</DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-2">
              <div className="space-y-2">
                <Label htmlFor="session-name">Nome</Label>
                <Input
                  id="session-name"
                  placeholder="Ex: Atendimento Principal"
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="session-apikey">API Key</Label>
                <div className="flex gap-2">
                  <Input
                    id="session-apikey"
                    className="font-mono text-xs"
                    value={newApiKey}
                    onChange={(e) => setNewApiKey(e.target.value)}
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    className="shrink-0"
                    onClick={() => setNewApiKey(generateAPIKey())}
                    title="Gerar nova key"
                  >
                    <RefreshCw className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setDialogOpen(false)}>
                Cancelar
              </Button>
              <Button onClick={onNew} disabled={creating || !newName.trim()}>
                {creating ? <Loader2 className="h-4 w-4 animate-spin" /> : <Plus className="h-4 w-4" />}
                Criar
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <ConfirmDialog
          open={!!toDelete}
          onOpenChange={(o) => !o && setToDelete(null)}
          title="Delete account?"
          description={toDelete ? `${toDelete.name} will be logged out and removed.` : undefined}
          confirmLabel="Delete"
          destructive
          onConfirm={() => {
            if (toDelete) void remove(toDelete.id);
          }}
        />
      </div>
    </TooltipProvider>
  );
};
