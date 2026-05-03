import type { HookHandler, HookEvent } from "openclaw";

const CLAWMEMORY_URL = process.env.CLAWMEMORY_URL || "http://localhost:8765";
const CLAWMEMORY_API_KEY = process.env.CLAWMEMORY_API_KEY || "";

const SENSITIVE_PATTERNS = [
  /(?:password|passwd|pwd)\s*[:=]\s*\S+/gi,
  /(?:api[_-]?key|apikey)\s*[:=]\s*\S+/gi,
  /(?:secret|token|auth)\s*[:=]\s*\S+/gi,
  /(?:Bearer\s+)\S+/gi,
  /sk-[a-zA-Z0-9]{20,}/g,
  /cm-[a-zA-Z0-9]{20,}/g,
  /AKIA[A-Z0-9]{16}/g,
];

function filterSensitive(text: string): string {
  let filtered = text;
  for (const pattern of SENSITIVE_PATTERNS) {
    filtered = filtered.replace(pattern, "[REDACTED]");
  }
  return filtered;
}

interface ConversationMessage {
  role: string;
  content: string;
}

interface PushRequest {
  session_id: string;
  title?: string;
  project_path?: string;
  agent_name?: string;
  messages: ConversationMessage[];
  summary?: string;
}

async function pushToClawMemory(
  sessionId: string,
  messages: ConversationMessage[],
  workspacePath?: string
): Promise<void> {
  if (!CLAWMEMORY_API_KEY) {
    console.error("[clawmemory] CLAWMEMORY_API_KEY not set, skipping push");
    return;
  }

  if (messages.length === 0) {
    return;
  }

  const filtered = messages
    .filter((m) => m.content && m.content.trim().length >= 10)
    .map((m) => ({
      role: m.role,
      content: filterSensitive(m.content).slice(0, 5000),
    }));

  if (filtered.length === 0) {
    return;
  }

  const payload: PushRequest = {
    session_id: sessionId,
    agent_name: "openclaw",
    messages: filtered,
  };

  if (workspacePath) {
    payload.project_path = workspacePath;
  }

  const userMsgs = filtered.filter((m) => m.role === "user");
  if (userMsgs.length > 0) {
    payload.title = userMsgs[0].content.slice(0, 100);
  }

  try {
    const response = await fetch(
      `${CLAWMEMORY_URL}/api/v1/external/conversations`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-API-Key": CLAWMEMORY_API_KEY,
        },
        body: JSON.stringify(payload),
      }
    );

    if (!response.ok) {
      const body = await response.text();
      console.error(
        `[clawmemory] Push failed: ${response.status} ${body}`
      );
    } else {
      const result = await response.json();
      console.log(
        `[clawmemory] Pushed ${filtered.length} messages, ${result.created || 0} memories created`
      );
    }
  } catch (err: any) {
    console.error(`[clawmemory] Push error: ${err.message}`);
  }
}

function extractMessagesFromTranscript(
  transcript: any[]
): ConversationMessage[] {
  const messages: ConversationMessage[] = [];

  if (!Array.isArray(transcript)) return messages;

  for (const entry of transcript) {
    if (entry.type === "user" || entry.role === "user") {
      const content =
        entry.content || entry.text || entry.message || "";
      if (content && typeof content === "string") {
        messages.push({ role: "user", content });
      }
    } else if (
      entry.type === "assistant" ||
      entry.role === "assistant"
    ) {
      const content =
        entry.content || entry.text || entry.message || "";
      if (content && typeof content === "string") {
        messages.push({ role: "assistant", content });
      }
    }
  }

  return messages;
}

const handler: HookHandler = async (event: HookEvent) => {
  const validEvents = [
    "command:new",
    "command:reset",
    "session:compact:before",
  ];
  if (!validEvents.includes(event.type)) {
    return;
  }

  const sessionId =
    event.sessionKey || event.sessionId || `session-${Date.now()}`;

  let messages: ConversationMessage[] = [];

  if (event.transcript && Array.isArray(event.transcript)) {
    messages = extractMessagesFromTranscript(event.transcript);
  } else if (event.messages && Array.isArray(event.messages)) {
    messages = event.messages.map((m: any) => ({
      role: m.role || m.type || "user",
      content: m.content || m.text || "",
    }));
  }

  if (messages.length < 2) {
    return;
  }

  const workspacePath =
    event.workspacePath || event.workspace || process.env.OPENCLAW_WORKSPACE;

  await pushToClawMemory(sessionId, messages, workspacePath);
};

export default handler;
