'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import api from '@/lib/api';
import { useAuth } from './useAuth';
import { useWebSocket } from './useWebSocket';
import { encrypt, decrypt } from '@/lib/e2ee/crypto';
import type { MessageType } from '@/types/chat';
import { toast } from 'sonner';

interface WSNewMessage {
  type: 'new_message';
  message: MessageType;
  temp_id?: string;
}

interface WSTypingEvent {
  type: 'user_typing' | 'user_stop_typing';
  conversation_id: string;
  user_id: string;
}

interface WSError {
  type: 'error';
  error: string;
  ref_type?: string;
  temp_id?: string;
}

/**
 * Manages the message state for a single conversation.
 *
 * • Loads the initial page of messages via REST.
 * • Appends real-time messages received via WebSocket.
 * • Sends messages optimistically (instant UI, reconciled on server ack).
 * • Tracks the other participant's typing indicator.
 * • Supports cursor-based "load more" for scrolling back through history.
 */
export function useChat(conversationId: string | null, sharedKey: Uint8Array | null) {
  const { user } = useAuth();
  const { send, subscribe, unsubscribe, isConnected } = useWebSocket();
  const queryClient = useQueryClient();

  // Keep sharedKey in a ref so callbacks always see the latest value
  const sharedKeyRef = useRef(sharedKey);
  sharedKeyRef.current = sharedKey;

  const [messages, setMessages] = useState<MessageType[]>([]);
  const [hasMore, setHasMore] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isOtherTyping, setIsOtherTyping] = useState(false);

  // Track pending optimistic message temp IDs to reconcile on server ack
  const pendingTempIds = useRef(new Set<string>());
  // Track the current conversation in a ref so WS callbacks always see latest
  const conversationIdRef = useRef(conversationId);
  conversationIdRef.current = conversationId;

  const typingTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  /**
   * Attempt to decrypt a single message if it's marked as encrypted.
   * Returns the message with plaintext content, or with _decryptionFailed set.
   */
  const decryptMessage = useCallback(
    (msg: MessageType): MessageType => {
      if (!msg.encrypted) return msg;
      const key = sharedKeyRef.current;
      if (!key) return { ...msg, _decryptionFailed: true };
      const plaintext = decrypt(msg.content, key);
      if (plaintext === null) return { ...msg, _decryptionFailed: true };
      return { ...msg, content: plaintext };
    },
    [],
  );

  /** Decrypt an array of messages. */
  const decryptMessages = useCallback(
    (msgs: MessageType[]): MessageType[] => msgs.map(decryptMessage),
    [decryptMessage],
  );

  // ── Load initial messages ────────────────────────────────────────────────

  useEffect(() => {
    if (!conversationId) {
      setMessages([]);
      setHasMore(false);
      return;
    }

    let cancelled = false;
    setIsLoading(true);

    api.conversations
      .getMessages(conversationId, { limit: 50 })
      .then((res) => {
        if (cancelled) return;
        // API returns oldest-first (ascending created_at)
        setMessages(decryptMessages(res.messages));
        setHasMore(res.has_more);
      })
      .catch(() => {
        if (!cancelled) toast.error('Failed to load messages');
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });

    // Mark conversation as read on open
    api.conversations.markRead(conversationId).then(() => {
      queryClient.invalidateQueries({ queryKey: ['chat-unread-count'] });
      queryClient.invalidateQueries({ queryKey: ['conversations'] });
    }).catch(() => {});

    return () => {
      cancelled = true;
    };
  }, [conversationId]);

  // ── WebSocket: new messages ──────────────────────────────────────────────

  useEffect(() => {
    if (!conversationId) return;

    const handleNewMessage = (raw: unknown) => {
      const data = raw as WSNewMessage;
      const msg = data.message;
      if (!msg || msg.conversation_id !== conversationIdRef.current) return;

      const tempId = data.temp_id;

      const decrypted = decryptMessage(msg);

      // Reconcile optimistic message from this client
      if (tempId && pendingTempIds.current.has(tempId)) {
        pendingTempIds.current.delete(tempId);
        setMessages((prev) =>
          prev.map((m) => (m.message_id === tempId ? decrypted : m)),
        );
        return;
      }

      // Deduplicate (e.g. reconnect replay or message from another tab)
      setMessages((prev) => {
        if (prev.some((m) => m.message_id === msg.message_id)) return prev;
        return [...prev, decrypted];
      });

      // Auto-mark read since this conversation is open
      if (msg.sender_id !== user?.user_id) {
        api.conversations.markRead(conversationIdRef.current!).then(() => {
          queryClient.invalidateQueries({ queryKey: ['chat-unread-count'] });
          queryClient.invalidateQueries({ queryKey: ['conversations'] });
        }).catch(() => {});
      }
    };

    subscribe('new_message', handleNewMessage);
    return () => unsubscribe('new_message', handleNewMessage);
  }, [conversationId, subscribe, unsubscribe, user?.user_id, decryptMessage]);

  // ── WebSocket: typing indicators ─────────────────────────────────────────

  useEffect(() => {
    if (!conversationId) return;

    const handleTyping = (raw: unknown) => {
      const data = raw as WSTypingEvent;
      if (data.conversation_id !== conversationIdRef.current) return;
      if (data.user_id === user?.user_id) return;
      setIsOtherTyping(true);
      // Auto-clear after 4s if no stop_typing arrives
      if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
      typingTimeoutRef.current = setTimeout(() => setIsOtherTyping(false), 4000);
    };

    const handleStopTyping = (raw: unknown) => {
      const data = raw as WSTypingEvent;
      if (data.conversation_id !== conversationIdRef.current) return;
      if (data.user_id === user?.user_id) return;
      setIsOtherTyping(false);
      if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
    };

    subscribe('user_typing', handleTyping);
    subscribe('user_stop_typing', handleStopTyping);
    return () => {
      unsubscribe('user_typing', handleTyping);
      unsubscribe('user_stop_typing', handleStopTyping);
      if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
      setIsOtherTyping(false);
    };
  }, [conversationId, subscribe, unsubscribe, user?.user_id]);

  // ── WebSocket: send errors ───────────────────────────────────────────────

  useEffect(() => {
    const handleError = (raw: unknown) => {
      const data = raw as WSError;
      if (data.ref_type === 'send_message' && data.temp_id) {
        if (pendingTempIds.current.has(data.temp_id)) {
          pendingTempIds.current.delete(data.temp_id);
          setMessages((prev) =>
            prev.filter((m) => m.message_id !== data.temp_id),
          );
          toast.error(data.error || 'Failed to send message');
        }
      }
    };

    subscribe('error', handleError);
    return () => unsubscribe('error', handleError);
  }, [subscribe, unsubscribe]);

  // ── WebSocket: reconnect → refetch to catch missed messages ──────────────

  useEffect(() => {
    if (!conversationId) return;

    const handleReconnect = () => {
      pendingTempIds.current.clear();
      api.conversations
        .getMessages(conversationId, { limit: 50 })
        .then((res) => {
          setMessages(decryptMessages(res.messages));
          setHasMore(res.has_more);
        })
        .catch(() => {});
    };

    subscribe('connected', handleReconnect);
    return () => unsubscribe('connected', handleReconnect);
  }, [conversationId, subscribe, unsubscribe, decryptMessages]);

  // ── Send message (optimistic + WS / REST fallback) ──────────────────────

  const sendMessage = useCallback(
    async (content: string, imageIds: string[] = []) => {
      const cid = conversationIdRef.current;
      if (!cid || !user) return;

      const tempId = crypto.randomUUID();

      // Encrypt content if we have a shared key
      const key = sharedKeyRef.current;
      const isEncrypted = !!key;
      const wireContent = key ? encrypt(content, key) : content;

      const optimistic: MessageType = {
        message_id: tempId,
        conversation_id: cid,
        sender_id: user.user_id,
        content, // Show plaintext locally
        encrypted: isEncrypted,
        has_images: imageIds.length > 0,
        is_edited: false,
        created_at: new Date().toISOString(),
      };

      setMessages((prev) => [...prev, optimistic]);

      if (isConnected) {
        pendingTempIds.current.add(tempId);
        send('send_message', {
          conversation_id: cid,
          content: wireContent,
          encrypted: isEncrypted,
          temp_id: tempId,
          ...(imageIds.length ? { image_ids: imageIds } : {}),
        });
      } else {
        // REST fallback
        try {
          const saved = await api.conversations.sendMessage(
            cid, wireContent, isEncrypted,
            imageIds.length ? imageIds : undefined,
          );
          setMessages((prev) =>
            prev.map((m) => (m.message_id === tempId ? decryptMessage(saved) : m)),
          );
        } catch {
          setMessages((prev) => prev.filter((m) => m.message_id !== tempId));
          toast.error('Failed to send message');
        }
      }
    },
    [user, isConnected, send, decryptMessage],
  );

  // ── Load older messages (cursor pagination) ──────────────────────────────

  const loadMore = useCallback(async () => {
    const cid = conversationIdRef.current;
    if (!cid || !hasMore || messages.length === 0) return;

    const oldest = messages[0];
    try {
      const res = await api.conversations.getMessages(cid, {
        before: oldest.created_at,
        limit: 50,
      });
      setMessages((prev) => [...decryptMessages(res.messages), ...prev]);
      setHasMore(res.has_more);
    } catch {
      toast.error('Failed to load older messages');
    }
  }, [hasMore, messages, decryptMessages]);

  // ── Typing indicator emission ────────────────────────────────────────────

  const typingEmitRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const emitTyping = useCallback(() => {
    const cid = conversationIdRef.current;
    if (!cid) return;

    send('typing', { conversation_id: cid });

    // Auto-send stop_typing after 3 seconds of inactivity
    if (typingEmitRef.current) clearTimeout(typingEmitRef.current);
    typingEmitRef.current = setTimeout(() => {
      send('stop_typing', { conversation_id: cid });
    }, 3000);
  }, [send]);

  const stopTyping = useCallback(() => {
    const cid = conversationIdRef.current;
    if (!cid) return;
    if (typingEmitRef.current) clearTimeout(typingEmitRef.current);
    send('stop_typing', { conversation_id: cid });
  }, [send]);

  return {
    messages,
    sendMessage,
    loadMore,
    hasMore,
    isLoading,
    isOtherTyping,
    emitTyping,
    stopTyping,
  };
}
