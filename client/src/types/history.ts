export type HistoryRow = {
  callId: string;
  peer: string;
  direction: string;
  startedAt: number;
  endedAt: number | null;
  endReason: string | null;
  recorded: boolean;
  recordingUrl: string;
};
