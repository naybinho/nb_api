import { useEffect, useState, useCallback } from "react";
import { Plus, Trash2, Loader2, Send, Check, WebhookIcon } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import {
  listWebhooks,
  createWebhook,
  updateWebhook,
  deleteWebhook,
  testWebhook,
  type Webhook,
} from "@/services/webhooks";

const EVENTS_PRESETS = [
  { label: "Todos", value: "*" },
  { label: "Mensagens", value: "message,message-receipt" },
  { label: "Chamadas", value: "incoming,call-status,call-ended" },
  { label: "Presença", value: "presence" },
];

export const WebhookManager = ({ sid }: { sid: string }) => {
  const [webhooks, setWebhooks] = useState<Webhook[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);

  // Form state
  const [url, setUrl] = useState("");
  const [events, setEvents] = useState("*");
  const [secret, setSecret] = useState("");
  const [enabled, setEnabled] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testingId, setTestingId] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      setLoading(true);
      const data = await listWebhooks(sid);
      setWebhooks(data);
    } catch (e) {
      toast.error("Erro ao carregar webhooks");
    } finally {
      setLoading(false);
    }
  }, [sid]);

  useEffect(() => {
    load();
  }, [load]);

  const resetForm = () => {
    setUrl("");
    setEvents("*");
    setSecret("");
    setEnabled(true);
    setEditingId(null);
    setShowForm(false);
  };

  const startEdit = (wh: Webhook) => {
    setUrl(wh.url);
    setEvents(wh.events);
    setSecret(wh.secret);
    setEnabled(wh.enabled);
    setEditingId(wh.id);
    setShowForm(true);
  };

  const handleSave = async () => {
    if (!url.trim()) {
      toast.error("URL é obrigatória");
      return;
    }
    setSaving(true);
    try {
      if (editingId) {
        await updateWebhook(sid, editingId, {
          url: url.trim(),
          events: events || "*",
          enabled,
          secret,
        });
        toast.success("Webhook atualizado!");
      } else {
        await createWebhook(sid, {
          url: url.trim(),
          events: events || "*",
          enabled,
          secret,
        });
        toast.success("Webhook criado!");
      }
      resetForm();
      load();
    } catch (e) {
      toast.error((e as Error).message);
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (wid: string) => {
    try {
      await deleteWebhook(sid, wid);
      toast.success("Webhook removido!");
      load();
    } catch (e) {
      toast.error((e as Error).message);
    }
  };

  const handleTest = async (wid: string) => {
    setTestingId(wid);
    try {
      const result = await testWebhook(sid, wid);
      toast.success(result.message || "Teste enviado com sucesso!");
    } catch (e) {
      toast.error("Falha no teste: " + (e as Error).message);
    } finally {
      setTestingId(null);
    }
  };

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <WebhookIcon className="h-4 w-4" />
          <span className="text-sm font-medium">Webhooks</span>
          {!loading && (
            <Badge variant="secondary" className="text-xs">
              {webhooks.length}
            </Badge>
          )}
        </div>
        {!showForm && (
          <Button size="sm" variant="outline" onClick={() => setShowForm(true)}>
            <Plus className="mr-1 h-3.5 w-3.5" />
            Adicionar
          </Button>
        )}
      </div>

      {showForm && (
        <div className="rounded-lg border p-3 space-y-3">
          <div className="space-y-1.5">
            <Label className="text-xs">URL do Webhook</Label>
            <Input
              placeholder="https://meusistema.com/webhook"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              className="h-8 text-xs"
            />
          </div>

          <div className="space-y-1.5">
            <Label className="text-xs">Eventos</Label>
            <div className="flex flex-wrap gap-1.5">
              {EVENTS_PRESETS.map((preset) => (
                <button
                  key={preset.value}
                  type="button"
                  onClick={() => setEvents(events === preset.value ? "" : preset.value)}
                  className={`rounded-md px-2 py-0.5 text-xs font-medium transition-colors ${
                    events === preset.value
                      ? "bg-primary text-primary-foreground"
                      : "bg-muted text-muted-foreground hover:bg-muted/80"
                  }`}
                >
                  {preset.label}
                </button>
              ))}
            </div>
            <Input
              placeholder="Ou digite manualmente: message,presence,call-status"
              value={events}
              onChange={(e) => setEvents(e.target.value)}
              className="h-7 text-xs"
            />
          </div>

          <div className="space-y-1.5">
            <Label className="text-xs">
              Segredo <span className="text-muted-foreground">(para assinatura HMAC, opcional)</span>
            </Label>
            <Input
              placeholder="MeuSegredo123"
              value={secret}
              onChange={(e) => setSecret(e.target.value)}
              className="h-8 text-xs"
            />
          </div>

          <div className="flex items-center justify-between">
            <Label className="text-xs cursor-pointer" onClick={() => setEnabled(!enabled)}>
              Ativo
            </Label>
            <Switch checked={enabled} onCheckedChange={setEnabled} />
          </div>

          <div className="flex gap-2">
            <Button size="sm" className="flex-1" onClick={handleSave} disabled={saving}>
              {saving && <Loader2 className="mr-1 h-3.5 w-3.5 animate-spin" />}
              {editingId ? "Salvar" : "Criar"}
            </Button>
            <Button size="sm" variant="ghost" onClick={resetForm}>
              Cancelar
            </Button>
          </div>
        </div>
      )}

      {loading ? (
        <div className="flex justify-center py-4">
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        </div>
      ) : webhooks.length === 0 && !showForm ? (
        <p className="text-xs text-muted-foreground">Nenhum webhook configurado.</p>
      ) : (
        <div className="space-y-2">
          {webhooks.map((wh) => (
            <div key={wh.id} className="rounded-lg border p-3">
              <div className="flex items-start justify-between gap-2">
                <div className="min-w-0 flex-1 space-y-1">
                  <div className="flex items-center gap-2">
                    <span className="truncate text-sm font-medium">{wh.url}</span>
                    {wh.enabled ? (
                      <Badge variant="success" className="text-[10px]">
                        Ativo
                      </Badge>
                    ) : (
                      <Badge variant="secondary" className="text-[10px]">
                        Inativo
                      </Badge>
                    )}
                  </div>
                  <div className="flex flex-wrap gap-1">
                    {wh.events === "*" ? (
                      <Badge variant="outline" className="text-[10px]">
                        Todos os eventos
                      </Badge>
                    ) : (
                      wh.events.split(",").map((evt) => (
                        <Badge key={evt} variant="outline" className="text-[10px]">
                          {evt.trim()}
                        </Badge>
                      ))
                    )}
                  </div>
                </div>
                <div className="flex shrink-0 gap-1">
                  <Button
                    size="icon"
                    variant="ghost"
                    className="h-7 w-7"
                    onClick={() => handleTest(wh.id)}
                    disabled={testingId === wh.id}
                    title="Testar"
                  >
                    {testingId === wh.id ? (
                      <Loader2 className="h-3.5 w-3.5 animate-spin" />
                    ) : (
                      <Send className="h-3.5 w-3.5" />
                    )}
                  </Button>
                  <Button
                    size="icon"
                    variant="ghost"
                    className="h-7 w-7"
                    onClick={() => startEdit(wh)}
                    title="Editar"
                  >
                    <Check className="h-3.5 w-3.5" />
                  </Button>
                  <Button
                    size="icon"
                    variant="ghost"
                    className="h-7 w-7 text-destructive hover:text-destructive"
                    onClick={() => handleDelete(wh.id)}
                    title="Remover"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};