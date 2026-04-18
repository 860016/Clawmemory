import { ref, onUnmounted } from 'vue'

interface WebSocketOptions {
  /** Max reconnection attempts (default: 5) */
  maxRetries?: number
  /** Base delay in ms for exponential backoff (default: 1000) */
  baseDelay?: number
  /** Max delay cap in ms (default: 30000) */
  maxDelay?: number
  /** Whether to auto-reconnect (default: true) */
  autoReconnect?: boolean
}

export function useWebSocket(options: WebSocketOptions = {}) {
  const {
    maxRetries = 5,
    baseDelay = 1000,
    maxDelay = 30000,
    autoReconnect = true,
  } = options

  const ws = ref<WebSocket | null>(null)
  const connected = ref(false)
  const messages = ref<any[]>([])
  const reconnectAttempts = ref(0)
  const isManualClose = ref(false)

  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let currentToken = ''

  function getReconnectDelay(): number {
    const delay = baseDelay * Math.pow(2, reconnectAttempts.value)
    const jitter = delay * 0.2 * Math.random()
    return Math.min(delay + jitter, maxDelay)
  }

  function connect(token: string) {
    currentToken = token
    isManualClose.value = false
    doConnect(token)
  }

  function doConnect(token: string) {
    if (ws.value) {
      ws.value.onclose = null
      ws.value.close()
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const socket = new WebSocket(`${protocol}//${window.location.host}/api/v1/chat/ws?token=${token}`)

    socket.onopen = () => {
      connected.value = true
      reconnectAttempts.value = 0
    }

    socket.onclose = () => {
      connected.value = false
      ws.value = null

      if (autoReconnect && !isManualClose.value && reconnectAttempts.value < maxRetries) {
        const delay = getReconnectDelay()
        reconnectAttempts.value++
        reconnectTimer = setTimeout(() => {
          doConnect(currentToken)
        }, delay)
      }
    }

    socket.onerror = () => {
      // onclose will fire after onerror, reconnection handled there
    }

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        messages.value.push(data)
      } catch {
        // ignore non-JSON messages
      }
    }

    ws.value = socket
  }

  function send(data: object) {
    if (ws.value && ws.value.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify(data))
    }
  }

  function disconnect() {
    isManualClose.value = true
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    reconnectAttempts.value = 0
    if (ws.value) {
      ws.value.onclose = null
      ws.value.close()
      ws.value = null
    }
    connected.value = false
  }

  onUnmounted(() => {
    disconnect()
  })

  return {
    ws,
    connected,
    messages,
    reconnectAttempts,
    connect,
    send,
    disconnect,
  }
}
