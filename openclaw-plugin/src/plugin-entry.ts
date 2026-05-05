import register from "./index";

export default {
  id: "clawmemory",
  name: "ClawMemory Context Engine",
  description:
    "Use ClawMemory as OpenClaw's memory backend. Automatically records conversations and injects relevant memories into context.",
  kind: "context-engine",
  register,
};
