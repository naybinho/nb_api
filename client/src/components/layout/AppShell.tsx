import { useState, type ReactNode } from "react";
import { BookOpen, Cpu, Menu, PhoneCall, QrCode, Server } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import { Sidebar } from "./Sidebar";
import { ThemeToggle } from "./ThemeToggle";
import { cn } from "@/lib/utils";

export type ViewType = "api";
export type ApiSubView = "instancias" | "calls" | "pix";

export const AppShell = ({
  children,
  view,
  onViewChange,
  subView,
  onSubViewChange,
}: {
  children: ReactNode;
  view: ViewType;
  onViewChange: (v: ViewType) => void;
  subView: ApiSubView;
  onSubViewChange: (v: ApiSubView) => void;
}) => {
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <div className="flex min-h-screen flex-col">
      {/* ── Header ── */}
      <header className="sticky top-0 z-30 flex items-center justify-between border-b bg-background/80 px-4 py-3 backdrop-blur sm:px-6">
        <div className="flex items-center gap-3">
          <Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
            <SheetTrigger asChild>
              <Button variant="outline" size="icon" className="md:hidden" aria-label="Accounts">
                <Menu className="h-4 w-4" />
              </Button>
            </SheetTrigger>
            <SheetContent side="left" className="w-72 p-0">
              <SheetTitle className="px-3 pt-3">Accounts</SheetTitle>
              <Sidebar onNavigate={() => setMobileOpen(false)} />
            </SheetContent>
          </Sheet>
          <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <Cpu className="h-4 w-4" />
          </span>
          <div className="flex items-baseline gap-1.5">
            <span className="text-lg font-semibold tracking-tight">NB_API</span>
            <span className="rounded-md border px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
              v2.0.0
            </span>
          </div>
        </div>

        <ThemeToggle />
      </header>

      {/* ── Sub-navigation ── */}
      <nav className="flex items-center gap-1 border-b bg-muted/30 px-4 py-1.5 sm:px-6">
          <button
            onClick={() => onSubViewChange("instancias")}
            className={cn(
              "flex items-center gap-1.5 rounded-md px-3 py-1 text-xs font-medium transition-colors",
              subView === "instancias"
                ? "bg-primary text-primary-foreground"
                : "text-muted-foreground hover:text-foreground",
            )}
          >
            <Server className="h-3.5 w-3.5" />
            Instâncias
          </button>
          <button
            onClick={() => onSubViewChange("calls")}
            className={cn(
              "flex items-center gap-1.5 rounded-md px-3 py-1 text-xs font-medium transition-colors",
              subView === "calls"
                ? "bg-primary text-primary-foreground"
                : "text-muted-foreground hover:text-foreground",
            )}
          >
            <PhoneCall className="h-3.5 w-3.5" />
            Chamadas
          </button>
          <button
            onClick={() => onSubViewChange("pix")}
            className={cn(
              "flex items-center gap-1.5 rounded-md px-3 py-1 text-xs font-medium transition-colors",
              subView === "pix"
                ? "bg-primary text-primary-foreground"
                : "text-muted-foreground hover:text-foreground",
            )}
          >
            <QrCode className="h-3.5 w-3.5" />
            PIX
          </button>
          <a
            href="/swagger"
            target="_blank"
            rel="noopener noreferrer"
            className={cn(
              "flex items-center gap-1.5 rounded-md px-3 py-1 text-xs font-medium transition-colors",
              "text-muted-foreground hover:text-foreground",
            )}
          >
            <BookOpen className="h-3.5 w-3.5" />
            API Docs
          </a>
        </nav>

      {/* ── Body ── */}
      <div className="flex flex-1">
        <aside className="hidden w-64 shrink-0 border-r md:block">
          <Sidebar />
        </aside>
        <main className="flex-1 px-4 py-6 sm:px-6">{children}</main>
      </div>
    </div>
  );
};
