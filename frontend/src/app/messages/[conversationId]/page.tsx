'use client';

import { use } from 'react';
import { Loader2 } from 'lucide-react';
import { ChatWindow } from '@/components/ChatWindow';
import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';

/**
 * `/messages/[conversationId]` — direct-link entry point for a conversation.
 *
 * Used when navigating from a notification, swap page, or shared link.
 * Shows a full-screen ChatWindow with a back button to `/messages`.
 */
export default function ConversationPage({
  params,
}: {
  params: Promise<{ conversationId: string }>;
}) {
  const { conversationId } = use(params);
  const { isLoading } = useAuth(true);
  const router = useRouter();

  if (isLoading) {
    return (
      <div className="flex h-[calc(100vh-4rem)] items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="h-[calc(100vh-4rem)] bg-background">
      <ChatWindow
        conversationId={conversationId}
        onBack={() => router.push('/messages')}
      />
    </div>
  );
}
