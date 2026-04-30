'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { MailWarning, Loader2 } from 'lucide-react';
import { toast } from 'sonner';
import { Button } from '@/components/ui/button';
import api from '@/lib/api';

/**
 * Banner shown to authenticated users whose email is not yet verified.
 *
 * Reads verification status from /auth/me (the JWT does not carry it, and the
 * localStorage 'user' object is a display cache only, so it cannot be trusted
 * for verification state).
 *
 * Safe to mount on any authenticated page. Renders nothing while the status
 * is loading, on error, or when the user is already verified.
 */
export function UnverifiedEmailBanner() {
  const [sending, setSending] = useState(false);

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['auth', 'me'],
    queryFn: () => api.auth.me(),
    // Verification status is sticky enough that a fresh fetch on mount is
    // sufficient; no polling.
    staleTime: 60_000,
  });

  if (isLoading || isError || !data || data.email_verified) return null;

  const handleResend = async () => {
    setSending(true);
    try {
      await api.auth.resendVerification();
      toast.success('Verification email sent. Check your inbox.');
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Could not send email';
      toast.error(msg);
    } finally {
      setSending(false);
      // Re-fetch so that if the user verified in another tab, the banner
      // disappears on next interaction.
      refetch();
    }
  };

  return (
    <div className="mb-6 flex flex-col gap-3 rounded-lg border border-amber-300/60 bg-amber-50 p-4 text-amber-900 sm:flex-row sm:items-center sm:justify-between dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-100">
      <div className="flex items-start gap-3">
        <MailWarning className="mt-0.5 h-5 w-5 shrink-0" />
        <div className="text-sm">
          <p className="font-medium">Verify your email address</p>
          <p className="text-amber-800/90 dark:text-amber-100/80">
            We sent a verification link to <span className="font-medium">{data.email}</span>.
            Verifying unlocks swap requests and protects your account.
          </p>
        </div>
      </div>
      <Button
        variant="outline"
        size="sm"
        onClick={handleResend}
        disabled={sending}
        className="self-start border-amber-400 bg-white/60 text-amber-900 hover:bg-white sm:self-center dark:bg-amber-500/10 dark:text-amber-100 dark:hover:bg-amber-500/20"
      >
        {sending ? (
          <>
            <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" />
            Sending…
          </>
        ) : (
          'Resend email'
        )}
      </Button>
    </div>
  );
}
