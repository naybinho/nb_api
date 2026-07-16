import { useState } from "react";
import { Headphones, History, Play, Square } from "lucide-react";
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { EmptyState } from "@/components/shared/EmptyState";
import { Badge } from "@/components/ui/badge";
import { useHistory } from "@/hooks/useHistory";

const formatDuration = (startedAt: number, endedAt: number | null): string => {
  if (!endedAt) return "";
  const secs = Math.round((endedAt - startedAt) / 1000);
  if (secs < 1) return "< 1s";
  const m = Math.floor(secs / 60);
  const s = secs % 60;
  return m > 0 ? `${m}m ${s}s` : `${s}s`;
};

export const HistoryDrawer = ({ sid }: { sid: string }) => {
  const [open, setOpen] = useState(false);
  const { data: rows = [] } = useHistory(sid, open);

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="outline" size="sm">
          <History className="h-4 w-4" />
          History
        </Button>
      </SheetTrigger>
      <SheetContent side="right" className="w-full p-0 sm:max-w-md">
        <SheetHeader className="p-6 pb-4">
          <SheetTitle>Call history</SheetTitle>
        </SheetHeader>
        <Separator />
        <ScrollArea className="h-[calc(100vh-5.5rem)] px-6 py-4">
          {rows.length === 0 ? (
            <EmptyState title="No past calls" description="Calls you make or receive will appear here." />
          ) : (
            <ul className="space-y-2">
              {rows.map((r) => (
                <li key={r.callId} className="rounded-lg border p-3">
                  <div className="flex items-start justify-between gap-2">
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-medium">{r.peer}</p>
                      <div className="mt-1 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                        <Badge variant="outline" className="text-[10px]">
                          {r.direction}
                        </Badge>
                        <span>{new Date(r.startedAt).toLocaleString()}</span>
                        {r.endedAt && (
                          <span>· {formatDuration(r.startedAt, r.endedAt)}</span>
                        )}
                        {r.endReason && r.endReason !== "unknown" && (
                          <span>· {r.endReason}</span>
                        )}
                      </div>
                    </div>
                    <div className="flex shrink-0 items-center gap-1">
                      {r.recorded ? (
                        r.recordingUrl ? (
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7"
                            onClick={() => window.open(r.recordingUrl, "_blank")}
                          >
                            <Play className="h-3.5 w-3.5" />
                            <span className="sr-only">Play recording</span>
                          </Button>
                        ) : (
                          <span title="Recorded">
                            <Headphones className="h-3.5 w-3.5 text-muted-foreground" />
                          </span>
                        )
                      ) : (
                        <span title="Not recorded">
                          <Square className="h-3.5 w-3.5 text-muted-foreground/40" />
                        </span>
                      )}
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </ScrollArea>
      </SheetContent>
    </Sheet>
  );
};
