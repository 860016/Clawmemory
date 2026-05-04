import type { HookHandler } from "openclaw";

const CLAWMEMORY_URL = process.env.CLAWMEMORY_URL || "http://localhost:8765";
const CLAWMEMORY_API_KEY = process.env.CLAWMEMORY_API_KEY || "";

const SENSITIVE_PATTERNS = [
  /(?:password|passwd|pwd)\s*[:=]\s*\S+/gi,
  /(?:api[_-]?key|apikey)\s*[:=]\s*\S+/gi,
  /(?:secret[_-]?key|secret[_-]?token)\s*[:=]\s*\S+/gi,
  /(?:Bearer\s+)\S+/gi,
  /sk-[a-zA-Z0-9]{20,}/g,
  /cm-[a-zA-Z0-9]{20,}/g,
  /AKIA[A-Z0-9]{16}/g,
];

function filterSensitive(text: string): string {
  let filtered = text;
  for (const pattern of SENSITIVE_PATTERNS) {
    const regex = new RegExp(pattern.source, pattern.flags);
    filtered = filtered.replace(regex, "[REDACTED]");
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

function safeSlice(text: string, maxLen: number): string {
  if (text.length <= maxLen) return text;
  let end = maxLen;
  while (end > 0 && (text.charCodeAt(end) & 0xfc00) === 0xdc00) {
    end--;
  }
  return text.slice(0, end);
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
      content: safeSlice(filterSensitive(m.content), 5000),
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
    payload.title = safeSlice(userMsgs[0].content, 100);
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
      let created = 0;
      try {
        const result = await response.json();
        created = result.created || 0;
      } catch {
        console.log("[clawmemory] Push succeeded but response was not JSON");
      }
      console.log(
        `[clawmemory] Pushed ${filtered.length} messages, ${created} memories created`
      );
    }
  } catch (err: any) {
    console.error(`[clawmemory] Push error: ${err.message}`);
  }
}

function extractTextContent(raw: any): string {
  if (typeof raw === "string") return raw;
  if (Array.isArray(raw)) {
    return raw
      .map((item: any) => {
        if (typeof item === "string") return item;
        if (item && typeof item === "object" && typeof item.text === "string") return item.text;
        return "";
      })
      .filter((s: string) => s.length > 0)
      .join("\n");
  }
  if (raw && typeof raw === "object" && typeof raw.text === "string") return raw.text;
  return "";
}

function extractMessagesFromSessionEntry(sessionEntry: any): ConversationMessage[] {
  const messages: ConversationMessage[] = [];
  if (!sessionEntry) return messages;

  const transcript = sessionEntry.transcript || sessionEntry.messages || [];
  if (!Array.isArray(transcript)) return messages;

  for (const entry of transcript) {
    const role = entry.type || entry.role || "";
    const rawContent = entry.content || entry.text || entry.message || "";
    const content = extractTextContent(rawContent);

    if (role === "user" && typeof content === "string" && content.trim()) {
      messages.push({ role: "user", content });
    } else if (
      (role === "assistant" || role === "agent") &&
      typeof content === "string" &&
      content.trim()
    ) {
      messages.push({ role: "assistant", content });
    }
  }

  return messages;
}

function extractMessagesFromContext(context: any): ConversationMessage[] {
  const messages: ConversationMessage[] = [];

  if (context.previousSessionEntry) {
    const prev = extractMessagesFromSessionEntry(context.previousSessionEntry);
    messages.push(...prev);
  }

  if (context.sessionEntry) {
    const curr = extractMessagesFromSessionEntry(context.sessionEntry);
    messages.push(...curr);
  }

  return messages;
}

const handler: HookHandler = async (event: any) => {
  if (event.type === "command") {
    if (event.action !== "new" && event.action !== "reset") {
      return;
    }
  } else if (event.type === "session:compact:before") {
    // continue
  } else {
    return;
  }

  const sessionId =
    event.sessionKey || (event.context && event.context.sessionEntry && event.context.sessionEntry.key) || `session-${Date.now()}`;

  let messages: ConversationMessage[] = [];

  if (event.context) {
    messages = extractMessagesFromContext(event.context);
  }

  if (messages.length === 0 && event.transcript && Array.isArray(event.transcript)) {
    for (const entry of event.transcript) {
      const role = entry.type || entry.role || "";
      const rawContent = entry.content || entry.text || entry.message || "";
      const content = extractTextContent(rawContent);
      if ((role === "user" || role === "assistant" || role === "agent") && typeof content === "string" && content.trim()) {
        messages.push({ role: role === "agent" ? "assistant" : role, content });
      }
    }
  }

  if (messages.length < 2) {
    return;
  }

  const workspacePath =
    (event.context && event.context.workspaceDir) ||
    event.workspacePath ||
    event.workspace ||
    process.env.OPENCLAW_WORKSPACE;

  await pushToClawMemory(sessionId, messages, workspacePath);
};

export default handler;
