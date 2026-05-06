'use client';

// ── WebSocket manager — singleton per browser tab ────────────────────────────
//
// Provides a single WebSocket connection with:
//   • Automatic reconnect (exponential backoff: 1s → 2s → 4s → … → 30s max)
//   • Event-based message dispatch (on / off)
//   • Token refresh on every reconnect attempt
//   • Safe no-op when called during SSR
// ---------------------------------------------------------------------------

type WSListener = (data: unknown) => void;

const MAX_RECONNECT_DELAY = 30_000;

class ChatWebSocket {
  private socket: WebSocket | null = null;
  private wsUrl = '';
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private reconnectAttempts = 0;
  private intentionalClose = false;
  private listeners = new Map<string, Set<WSListener>>();
  private _isConnected = false;

  // ---- public API ----------------------------------------------------------

  /** Open (or re-open) the WebSocket connection with the given JWT. */
  connect(token: string) {
    if (this.socket?.readyState === WebSocket.OPEN) return;

    this.intentionalClose = false;
    this.wsUrl = buildWsUrl(token);
    this.doConnect();
  }

  /** Gracefully close the connection and stop reconnection. */
  disconnect() {
    this.intentionalClose = true;
    this.clearReconnect();
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
    this._isConnected = false;
  }

  /** Send a typed JSON payload. Returns true if the socket was open. */
  send(type: string, payload: Record<string, unknown>): boolean {
    if (this.socket?.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify({ type, ...payload }));
      return true;
    }
    return false;
  }

  /** Subscribe to a WS event type (e.g. 'new_message', 'connected'). */
  on(event: string, handler: WSListener) {
    if (!this.listeners.has(event)) this.listeners.set(event, new Set());
    this.listeners.get(event)!.add(handler);
  }

  /** Remove a previously registered handler. */
  off(event: string, handler: WSListener) {
    this.listeners.get(event)?.delete(handler);
  }

  get isConnected() {
    return this._isConnected;
  }

  // ---- internals -----------------------------------------------------------

  private doConnect() {
    if (this.intentionalClose || !this.wsUrl) return;

    try {
      this.socket = new WebSocket(this.wsUrl);
    } catch {
      this.scheduleReconnect();
      return;
    }

    this.socket.onopen = () => {
      this._isConnected = true;
      this.reconnectAttempts = 0;
      this.emit('connected', undefined);
    };

    this.socket.onclose = () => {
      this._isConnected = false;
      this.emit('disconnected', undefined);
      if (!this.intentionalClose) this.scheduleReconnect();
    };

    // onclose fires after onerror — reconnect handled there
    this.socket.onerror = () => {};

    this.socket.onmessage = (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data as string);
        if (typeof data.type === 'string') this.emit(data.type, data);
      } catch {
        /* malformed frame — ignore */
      }
    };
  }

  private scheduleReconnect() {
    this.clearReconnect();
    const delay = Math.min(
      1000 * Math.pow(2, this.reconnectAttempts),
      MAX_RECONNECT_DELAY,
    );
    this.reconnectTimer = setTimeout(() => {
      this.reconnectAttempts++;
      // Grab a fresh token in case it was refreshed by an API call
      const freshToken =
        typeof window !== 'undefined'
          ? localStorage.getItem('access_token')
          : null;
      if (freshToken) {
        this.wsUrl = buildWsUrl(freshToken);
        this.doConnect();
      }
    }, delay);
  }

  private clearReconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  private emit(event: string, data: unknown) {
    this.listeners.get(event)?.forEach((fn) => fn(data));
  }
}

// ── Helpers ──────────────────────────────────────────────────────────────────

function buildWsUrl(token: string): string {
  const apiBase =
    process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080/api/v1';
  // http → ws, https → wss (replacing the leading "http" covers both)
  const wsBase = apiBase.replace(/^http/, 'ws');
  return `${wsBase}/ws?token=${encodeURIComponent(token)}`;
}

// ── Singleton (client-side only) ─────────────────────────────────────────────

export const chatWS: ChatWebSocket | null =
  typeof window !== 'undefined' ? new ChatWebSocket() : null;
