'use client';

import { useEffect } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import api from '@/lib/api';
import { useAuth } from './useAuth';
import { useWebSocket } from './useWebSocket';
import type { MessageType } from '@/types/chat';

/**
 * Provides the total unread chat message count.
 *
 * - Fetches the initial count via REST.
 * - Subscribes to WebSocket `new_message` events to
 *   invalidate the query (server becomes single source of truth).
 * - Polling fallback every 60 s when WS is disconnected.
 */
export function useUnreadCount() {
  const { isAuthenticated } = useAuth();
  const { subscribe, unsubscribe, isConnected } = useWebSocket();
  const queryClient = useQueryClient();

  const { data } = useQuery({
    queryKey: ['chat-unread-count'],
    queryFn: () => api.chatUnread.count(),
    enabled: isAuthenticated,
    // Poll at 60 s as fallback if WS is down
    refetchInterval: isConnected ? false : 60_000,
    staleTime: 30_000,
  });

  // Invalidate on real-time events so the count stays fresh
  useEffect(() => {
    const onNewMessage = (raw: unknown) => {
      const msg = (raw as { message?: MessageType }).message;
      if (msg) {
        queryClient.invalidateQueries({ queryKey: ['chat-unread-count'] });
      }
    };

    const onRead = () => {
      queryClient.invalidateQueries({ queryKey: ['chat-unread-count'] });
    };

    subscribe('new_message', onNewMessage);
    subscribe('message_read', onRead);
    return () => {
      unsubscribe('new_message', onNewMessage);
      unsubscribe('message_read', onRead);
    };
  }, [subscribe, unsubscribe, queryClient]);

  return data?.unread_count ?? 0;
}
