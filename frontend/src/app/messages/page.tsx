'use client';

import { useState, useEffect, Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import { MessageCircle, Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { ConversationList } from '@/components/ConversationList';
import { ChatWindow } from '@/components/ChatWindow';
import { useAuth } from '@/hooks/useAuth';
import api from '@/lib/api';

/**
 * `/messages` — split-panel messages page.
 *
 * Query params:
 *   ?conversation=<id>  — open a conversation directly
 *   ?swap=<id>          — resolve swap → conversation and open it
 *
 * Desktop (>= md): Conversation list (left 1/3) + Chat window (right 2/3).
 * Mobile  (< md):  One panel at a time, toggled by selecting / backing out.
 */
export default function MessagesPage() {
  return (
    <Suspense
      fallback={
        <div className="fixed inset-0 top-16 flex items-center justify-center">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      }
    >
      <MessagesContent />
    </Suspense>
  );
}

function MessagesContent() {
  const { isLoading } = useAuth(true);
  const searchParams = useSearchParams();
  const [selectedId, setSelectedId] = useState<string | null>(null);

  // Resolve ?conversation= or ?swap= query params on mount
  useEffect(() => {
    const convoParam = searchParams.get('conversation');
    const swapParam = searchParams.get('swap');

    if (convoParam) {
      setSelectedId(convoParam);
    } else if (swapParam) {
      // Resolve swap → conversation
      api.conversations
        .getBySwap(swapParam)
        .then((detail) => setSelectedId(detail.conversation_id))
        .catch(() => {
          // Conversation may not exist yet (swap not accepted)
        });
    }
  }, [searchParams]);

  if (isLoading) {
    return (
      <div className="fixed inset-0 top-16 flex items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="fixed inset-0 top-16 flex overflow-hidden bg-background">
      {/* ── Conversation list ──────────────────────────────────────────── */}
      <div
        className={cn(
          'h-full border-r border-border',
          'md:block md:w-1/3 lg:w-[320px] lg:min-w-[280px]',
          selectedId ? 'hidden' : 'w-full',
        )}
      >
        <div className="flex h-14 items-center border-b border-border px-4">
          <h1 className="text-lg font-semibold">Messages</h1>
        </div>
        <div className="h-[calc(100%-3.5rem)]">
          <ConversationList selectedId={selectedId} onSelect={setSelectedId} />
        </div>
      </div>

      {/* ── Chat window ────────────────────────────────────────────────── */}
      <div
        className={cn(
          'h-full flex-1',
          selectedId ? 'block' : 'hidden md:flex',
        )}
      >
        {selectedId ? (
          <ChatWindow
            conversationId={selectedId}
            onBack={() => setSelectedId(null)}
          />
        ) : (
          <EmptyState />
        )}
      </div>
    </div>
  );
}

function EmptyState() {
  return (
    <div className="flex h-full flex-col items-center justify-center gap-3 text-muted-foreground">
      <MessageCircle className="h-12 w-12 opacity-30" />
      <p className="text-sm">Select a conversation to start chatting</p>
    </div>
  );
}
