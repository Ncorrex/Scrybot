import { ref, onMounted, onUnmounted } from 'vue'

export function useWebSocket(url: string, onMessage: (data: string) => void) {
  const isConnected = ref(false)
  let ws: WebSocket | null = null
  let retryTimer: ReturnType<typeof setTimeout> | null = null
  let retryDelay = 1000
  let destroyed = false

  function connect() {
    if (destroyed) return
    ws = new WebSocket(url)

    ws.onopen = () => {
      isConnected.value = true
      retryDelay = 1000
    }

    ws.onmessage = (event: MessageEvent) => {
      onMessage(event.data as string)
    }

    ws.onclose = () => {
      isConnected.value = false
      if (!destroyed) {
        retryTimer = setTimeout(connect, retryDelay)
        retryDelay = Math.min(retryDelay * 2, 30_000)
      }
    }

    ws.onerror = () => {
      ws?.close()
    }
  }

  onMounted(connect)

  onUnmounted(() => {
    destroyed = true
    if (retryTimer !== null) clearTimeout(retryTimer)
    ws?.close()
  })

  return { isConnected }
}