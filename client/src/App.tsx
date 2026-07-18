import { useEffect, useState } from "react";
import { PlusCircle } from "lucide-react";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Toaster } from "@/components/ui/sonner";
import { AppShell, type ViewType, type ApiSubView } from "@/components/layout/AppShell";
import { CallsPage } from "@/pages/CallsPage";
import { InstanciasPage } from "@/pages/InstanciasPage";
import { PixSender } from "@/components/domain/pix/PixSender";
import { SessionPairing } from "@/components/domain/session/SessionPairing";
import { SessionHeader } from "@/components/domain/session/SessionHeader";
import { IncomingCallModal } from "@/components/domain/call/IncomingCallModal";
import { EmptyState } from "@/components/shared/EmptyState";
import { ensureSessionsWired, useSessions } from "@/stores/sessions";
import { ensureCallsWired } from "@/stores/calls";
import { useTheme } from "@/stores/theme";

export const App = () => {
  const sessions = useSessions((s) => s.sessions);
  const activeId = useSessions((s) => s.activeId);
  const theme = useTheme((s) => s.theme);
  const [view, setView] = useState<ViewType>("api");
  const [subView, setSubView] = useState<ApiSubView>("calls");

  useEffect(() => {
    ensureSessionsWired();
    ensureCallsWired();
  }, []);

  const active = sessions.find((s) => s.id === activeId) ?? null;

  return (
    <TooltipProvider delayDuration={200}>
      <AppShell view={view} onViewChange={setView} subView={subView} onSubViewChange={setSubView}>
        {view === "api" && subView === "instancias" ? (
          <InstanciasPage />
        ) : sessions.length === 0 ? (
          <EmptyState
            icon={<PlusCircle className="h-6 w-6" />}
            title="No accounts yet"
            description="Create your first WhatsApp account from the sidebar to start calling."
          />
        ) : active ? (
          <div className="space-y-6">
            <SessionHeader session={active} />
            {subView === "pix" ? (
              <PixSender sid={active.id} />
            ) : active.paired ? (
              <CallsPage sid={active.id} />
            ) : (
              <SessionPairing session={active} />
            )}
          </div>
        ) : (
          <EmptyState title="Select an account" description="Choose an account from the sidebar." />
        )}
      </AppShell>
      <IncomingCallModal />
      <Toaster theme={theme} position="top-right" richColors closeButton />
    </TooltipProvider>
  );
};
