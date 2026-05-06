'use client';

import { useState, useEffect, useCallback } from 'react';
import { chatWS } from '@/lib/websocket';
import { getAccessToken } from '@/lib/api';
import { useAuth } from './useAuth';

/**
 * Manages the global WebSocket connection.
 *
 * - Connects automatically when the user is authenticated.
 * - Disconnects on logout.
 * - The underlying ChatWebSocket singleton persists across SPA navigations.
 */
export function useWebSocket() {
  const { isAuthenticated } = useAuth();
  const [isConnected, setIsConnected] = useState(false);

  // Connect / disconnect in response to auth state
  useEffect(() => {
    if (!chatWS) return;

    if (isAuthenticated) {
      const token = getAccessToken();
      if (token) chatWS.connect(token);
    } else {
      chatWS.disconnect();
    }
  }, [isAuthenticated]);

  // Mirror the singleton's connection state into React state
  useEffect(() => {
    if (!chatWS) return;

    const onConnect = () => setIsConnected(true);
    const onDisconnect = () => setIsConnected(false);

    chatWS.on('connected', onConnect);
    chatWS.on('disconnected', onDisconnect);

    // Sync initial state
    if (chatWS.isConnected) setIsConnected(true);

    return () => {
      chatWS?.off('connected', onConnect);
      chatWS?.off('disconnected', onDisconnect);
    };
  }, []);

  const send = useCallback(
    (type: string, payload: Record<string, unknown>) =>
      chatWS?.send(type, payload) ?? false,
    [],
  );

  const subscribe = useCallback(
    (event: string, handler: (data: unknown) => void) => {
      chatWS?.on(event, handler);
    },
    [],
  );

  const unsubscribe = useCallback(
    (event: string, handler: (data: unknown) => void) => {
      chatWS?.off(event, handler);
    },
    [],
  );

  return { isConnected, send, subscribe, unsubscribe };
}
