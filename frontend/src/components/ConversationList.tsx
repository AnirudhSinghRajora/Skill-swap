'use client';

import { useEffect } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { cn } from '@/lib/utils';
import { Avatar } from '@/components/ui/avatar';
import { Skeleton } from '@/components/ui/skeleton';
import { useAuth } from '@/hooks/useAuth';
import { useWebSocket } from '@/hooks/useWebSocket';
import api, { getPhotoUrl } from '@/lib/api';
import type { ConversationListItem } from '@/types/chat';

interface ConversationListProps {
  selectedId: string | null;
  onSelect: (id: string) => void;
}

/**
 * Sidebar list of conversations.
 *
 * - Fetches conversations via REST.
 * - Refreshes on WebSocket `new_message` so last-message
 *   previews and unread badges stay current.
 */
export function ConversationList({ selectedId, onSelect }: ConversationListProps) {
  const { user, isAuthenticated } = useAuth();
  const { subscribe, unsubscribe } = useWebSocket();
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['conversations'],
    queryFn: () => api.conversations.list(),
    enabled: isAuthenticated,
    refetchInterval: 60_000,
  });

  // Invalidate on any real-time chat event so the list stays fresh
  useEffect(() => {
    const invalidate = () => {
      queryClient.invalidateQueries({ queryKey: ['conversations'] });
    };
    subscribe('new_message', invalidate);
    subscribe('message_read', invalidate);
    return () => {
      unsubscribe('new_message', invalidate);
      unsubscribe('message_read', invalidate);
    };
  }, [subscribe, unsubscribe, queryClient]);

  const conversations: ConversationListItem[] = data?.conversations ?? [];

  if (isLoading) {
    return (
      <div className="space-y-2 p-4">
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={i} className="flex items-center gap-3">
            <Skeleton className="h-10 w-10 rounded-full" />
            <div className="flex-1 space-y-1.5">
              <Skeleton className="h-3.5 w-28" />
              <Skeleton className="h-3 w-40" />
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (conversations.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center gap-2 p-8 text-center text-muted-foreground">
        <p className="font-medium">No conversations yet</p>
        <p className="text-sm">Accept a swap to start chatting!</p>
      </div>
    );
  }

  return (
    <div className="overflow-y-auto" data-lenis-prevent>
      {conversations.map((conv) => (
        <button
          key={conv.conversation_id}
          type="button"
          onClick={() => onSelect(conv.conversation_id)}
          className={cn(
            'flex w-full items-center gap-3 px-4 py-3 text-left transition-colors hover:bg-muted/50',
            selectedId === conv.conversation_id && 'bg-muted',
          )}
        >
          <Avatar
            src={conv.other_user.has_photo ? getPhotoUrl(conv.other_user.user_id) : undefined}
            alt={conv.other_user.name}
            fallback={conv.other_user.name.charAt(0).toUpperCase()}
            className="h-10 w-10 shrink-0"
          />

          <div className="min-w-0 flex-1">
            <div className="flex items-center justify-between gap-2">
              <span className="truncate text-sm font-medium">
                {conv.other_user.name}
              </span>
              {conv.last_message && (
                <span className="shrink-0 text-[11px] text-muted-foreground">
                  {formatRelativeTime(conv.last_message.created_at)}
                </span>
              )}
            </div>

            <div className="flex items-center justify-between gap-2">
              <p className="truncate text-xs text-muted-foreground">
                {conv.last_message
                  ? `${conv.last_message.sender_id === user?.user_id ? 'You: ' : ''}${stripToPlain(conv.last_message.content)}`
                  : `${conv.offered_skill} ↔ ${conv.wanted_skill}`}
              </p>
              {conv.unread_count > 0 && (
                <span className="flex h-5 min-w-5 shrink-0 items-center justify-center rounded-full bg-primary px-1.5 text-[10px] font-bold text-primary-foreground">
                  {conv.unread_count > 99 ? '99+' : conv.unread_count}
                </span>
              )}
            </div>
          </div>
        </button>
      ))}
    </div>
  );
}

// ── Helpers ──────────────────────────────────────────────────────────────────

/** Strip HTML tags and collapse whitespace for the last-message preview. */
function stripToPlain(html: string): string {
  return html.replace(/<[^>]*>/g, '').replace(/\s+/g, ' ').trim();
}

/** Human-friendly relative timestamp. */
function formatRelativeTime(iso: string): string {
  const date = new Date(iso);
  if (isNaN(date.getTime())) return '';
  const diff = Date.now() - date.getTime();
  const mins = Math.floor(diff / 60_000);
  if (mins < 1) return 'now';
  if (mins < 60) return `${mins}m`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d`;
  return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
}
