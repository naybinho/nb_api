export type CallStatus = "starting" | "ringing" | "connected" | "ended";

export type CallSummary = {
  sessionId: string;
  callId: string;
  owner: string | null;
  direction: "outbound" | "inbound";
  peer: string;
  startedAt: number;
  status: CallStatus;
  recorded: boolean;
  recordingUrl?: string;
};

export type IncomingPayload = { sessionId: string; callId: string; peer: string; offeredAt: number };
