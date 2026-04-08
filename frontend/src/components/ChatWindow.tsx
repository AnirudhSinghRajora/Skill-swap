'use client';

import { useEffect, useRef, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { ArrowLeft, ArrowUp, Loader2, Lock, Video } from 'lucide-react';
import { Avatar } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { ChatBubble } from '@/components/ChatBubble';
import { ChatInput } from '@/components/ChatInput';
import { TypingIndicator } from '@/components/TypingIndicator';
import { VideoCall } from '@/components/VideoCall';
import { useAuth } from '@/hooks/useAuth';
import { useChat } from '@/hooks/useChat';
import { useE2EEKeys } from '@/hooks/useE2EEKeys';
import { useConversationKeys } from '@/hooks/useConversationKeys';
import api, { getPhotoUrl } from '@/lib/api';

interface ChatWindowProps {
  conversationId: string;
  onBack?: () => void;
}

/**
 * Full chat view: header + messages + typing indicator + input.
 *
 * - Loads conversation details for the header (participant, skill exchange).
 * - Delegates message state to `useChat`.
 * - Auto-scrolls to newest message on arrival.
 */
export function ChatWindow({ conversationId, onBack }: ChatWindowProps) {
  const { user } = useAuth();
  const [showVideoCall, setShowVideoCall] = useState(false);

  // ── E2EE key derivation ────────────────────────────────────────────────
  const { keyPair } = useE2EEKeys();
  // Fetch conversation details early so we can derive the shared key
  const { data: convo, isLoading: convoLoading } = useQuery({
    queryKey: ['conversation-detail', conversationId],
    queryFn: () => api.conversations.get(conversationId),
    enabled: !!conversationId,
  });
  const otherUserId = convo?.other_user?.user_id;
  const { sharedKey } = useConversationKeys(keyPair, otherUserId);

  const {
    messages,
    sendMessage,
    loadMore,
    hasMore,
    isLoading: messagesLoading,
    isOtherTyping,
    emitTyping,
    stopTyping,
  } = useChat(conversationId, sharedKey);

  // ── Auto-scroll ──────────────────────────────────────────────────────────

  const scrollRef = useRef<HTMLDivElement>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  const isInitialLoad = useRef(true);

  // Reset initial-load flag when conversation changes
  useEffect(() => {
    isInitialLoad.current = true;
  }, [conversationId]);

  useEffect(() => {
    if (messages.length === 0) return;
    // Instant jump on first load, smooth scroll on subsequent messages
    const behavior = isInitialLoad.current ? 'instant' as const : 'smooth' as const;
    isInitialLoad.current = false;
    // Use rAF to ensure DOM has laid out the new messages before scrolling
    requestAnimationFrame(() => {
      bottomRef.current?.scrollIntoView({ behavior });
    });
  }, [messages.length]);

  // ── Header ───────────────────────────────────────────────────────────────

  const otherUser = convo?.other_user;

  if (convoLoading) {
    return (
      <div className="flex h-full flex-col">
        <div className="flex items-center gap-3 border-b border-border px-4 py-3">
          <Skeleton className="h-9 w-9 rounded-full" />
          <div className="space-y-1">
            <Skeleton className="h-4 w-32" />
            <Skeleton className="h-3 w-48" />
          </div>
        </div>
        <div className="flex flex-1 items-center justify-center">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col">
      {/* ── Header ─────────────────────────────────────────────────────── */}
      <div className="flex items-center gap-3 border-b border-border bg-card px-4 py-3">
        {onBack && (
          <Button
            variant="ghost"
            size="icon"
            className="shrink-0 md:hidden"
            onClick={onBack}
            aria-label="Back to conversations"
          >
            <ArrowLeft className="h-5 w-5" />
          </Button>
        )}

        {otherUser && (
          <Avatar
            src={otherUser.has_photo ? getPhotoUrl(otherUser.user_id) : undefined}
            alt={otherUser.name}
            fallback={otherUser.name.charAt(0).toUpperCase()}
            className="h-9 w-9 shrink-0"
          />
        )}

        <div className="min-w-0 flex-1">
          <h2 className="truncate text-sm font-semibold">
            {otherUser?.name ?? 'Chat'}
          </h2>
          {convo && (
            <>
              <p className="truncate text-xs text-muted-foreground">
                {convo.offered_skill.name} ↔ {convo.wanted_skill.name}
                {convo.swap_status === 'completed' && (
                  <Badge variant="secondary" className="ml-2 text-[10px]">
                    Completed
                  </Badge>
                )}
              </p>
              {sharedKey && (
                <p className="flex items-center gap-1 text-[10px] text-green-600 dark:text-green-400">
                  <Lock className="h-2.5 w-2.5" />
                  End-to-end encrypted
                </p>
              )}
            </>
          )}
        </div>

        <Button
          variant="ghost"
          size="icon"
          className="shrink-0"
          onClick={() => setShowVideoCall(true)}
          aria-label="Start video call"
        >
          <Video className="h-5 w-5" />
        </Button>
      </div>

      {/* ── Video Call Overlay ──────────────────────────────────────────── */}
      {showVideoCall && (
        <VideoCall
          conversationId={conversationId}
          onClose={() => setShowVideoCall(false)}
        />
      )}

      {/* ── Messages ───────────────────────────────────────────────────── */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto px-4 py-4" data-lenis-prevent>
        {/* Load more */}
        {hasMore && (
          <div className="mb-4 flex justify-center">
            <Button variant="ghost" size="sm" onClick={loadMore}>
              <ArrowUp className="mr-1 h-3 w-3" />
              Load older messages
            </Button>
          </div>
        )}

        {messagesLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : messages.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-center text-muted-foreground">
            <p className="text-sm">No messages yet.</p>
            <p className="text-xs">Send the first message to get started!</p>
          </div>
        ) : (
          <div className="space-y-2">
            {messages.map((msg) => (
              <ChatBubble
                key={msg.message_id}
                message={msg}
                isOwn={msg.sender_id === user?.user_id}
                sharedKey={sharedKey}
              />
            ))}
          </div>
        )}

        {isOtherTyping && <TypingIndicator name={otherUser?.name} />}

        {/* Scroll anchor */}
        <div ref={bottomRef} />
      </div>

      {/* ── Input ──────────────────────────────────────────────────────── */}
      <ChatInput
        onSend={(content, imageIds) => {
          sendMessage(content, imageIds);
          stopTyping();
        }}
        onTyping={emitTyping}
        sharedKey={sharedKey}
      />
    </div>
  );
}
