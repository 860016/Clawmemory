import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

// Mock WebSocket before importing the composable
class MockWebSocket {
  static OPEN = 1
  static CLOSED = 3
  static instances: MockWebSocket[] = []
  onopen: (() => void) | null = null
  onclose: (() => void) | null = null
  onerror: (() => void) | null = null
  onmessage: ((e: { data: string }) => void) | null = null
  readyState = 0
  sentData: string[] = []

  constructor(public url: string) {
    MockWebSocket.instances.push(this)
  }

  send(data: string) {
    this.sentData.push(data)
  }

  close() {
    this.readyState = 3
  }

  static reset() {
    MockWebSocket.instances = []
  }

  static get last() {
    return MockWebSocket.instances[MockWebSocket.instances.length - 1]
  }
}

// Stub global
vi.stubGlobal('WebSocket', MockWebSocket)

// Must import after stubbing
const { useWebSocket } = await import('../useWebSocket')

describe('useWebSocket', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    MockWebSocket.reset()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('should initialize with default state', () => {
    const { connected, messages, reconnectAttempts } = useWebSocket()
    expect(connected.value).toBe(false)
    expect(messages.value).toEqual([])
    expect(reconnectAttempts.value).toBe(0)
  })

  it('should create WebSocket connection on connect', () => {
    const { connect, connected } = useWebSocket()
    connect('test-token')

    expect(MockWebSocket.last.url).toContain('test-token')
    expect(connected.value).toBe(false)

    // Simulate open
    MockWebSocket.last.onopen!()
    expect(connected.value).toBe(true)
  })

  it('should handle messages', () => {
    const { connect, messages } = useWebSocket()
    connect('test-token')

    MockWebSocket.last.onopen!()
    MockWebSocket.last.onmessage!({ data: JSON.stringify({ type: 'chat', content: 'hello' }) })

    expect(messages.value).toHaveLength(1)
    expect(messages.value[0]).toEqual({ type: 'chat', content: 'hello' })
  })

  it('should attempt reconnection on unexpected close', () => {
    const { connect, connected, reconnectAttempts } = useWebSocket({
      baseDelay: 100,
      maxRetries: 3,
    })
    connect('test-token')

    MockWebSocket.last.onopen!()
    expect(connected.value).toBe(true)

    MockWebSocket.last.onclose!()
    expect(connected.value).toBe(false)

    vi.advanceTimersByTime(200)
    expect(reconnectAttempts.value).toBeGreaterThanOrEqual(1)
  })

  it('should not reconnect on manual disconnect', () => {
    const { connect, disconnect, reconnectAttempts } = useWebSocket()
    connect('test-token')
    MockWebSocket.last.onopen!()

    disconnect()
    expect(reconnectAttempts.value).toBe(0)

    vi.advanceTimersByTime(5000)
    expect(reconnectAttempts.value).toBe(0)
  })

  it('should send data when connected', () => {
    const { connect, send } = useWebSocket()
    connect('test-token')
    const ws = MockWebSocket.last
    ws.onopen!()
    ws.readyState = 1 // OPEN

    send({ type: 'message', content: 'test' })
    expect(ws.sentData).toContain('{"type":"message","content":"test"}')
  })

  it('should not send data when not connected', () => {
    const { connect, send } = useWebSocket()
    connect('test-token')
    // Don't call onopen - simulate not connected

    send({ type: 'message', content: 'test' })
    expect(MockWebSocket.last.sentData).toHaveLength(0)
  })
})
