'use client';

/**
 * E2EESetupGate — overlay modal that finishes E2EE setup when the
 * `useEnsureE2EE` hook reports it is incomplete.
 *
 * It is mounted once in the root layout but only appears on chat-bearing
 * routes (currently `/messages*` and `/swaps/*/chat`). For every other
 * route it stays out of the way so the rest of the app — browse, profile,
 * onboarding — remains usable even without keys.
 *
 * After the user submits a passphrase we call `initializeE2EE` which
 * either generates+uploads a fresh pair or restores from the encrypted
 * backup. On success the hook re-fetches `/auth/me` and the gate hides
 * itself.
 */

import { useState } from 'react';
import { usePathname } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { useEnsureE2EE } from '@/hooks/useEnsureE2EE';
import { initializeE2EE, E2EERestoreError } from '@/lib/e2ee/init';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

const GATED_PATH_PREFIXES = ['/messages', '/swaps'];

function shouldGate(pathname: string | null): boolean {
  if (!pathname) return false;
  return GATED_PATH_PREFIXES.some((prefix) => pathname.startsWith(prefix));
}

interface CopyForState {
  title: string;
  body: string;
  cta: string;
}

const COPY: Record<'needs-passphrase' | 'needs-restore' | 'mismatch', CopyForState> = {
  'needs-passphrase': {
    title: 'Set up secure messaging',
    body: 'Choose a passphrase to encrypt your messages. We use it to derive a key locally — we never see it. If you forget it you will lose access to past conversations.',
    cta: 'Generate keys',
  },
  'needs-restore': {
    title: 'Unlock secure messaging',
    body: 'We have an encrypted backup of your keys on this account. Enter your passphrase to restore them on this device.',
    cta: 'Restore keys',
  },
  mismatch: {
    title: 'Resync secure messaging',
    body: 'Your device still has keys, but the server lost them. Re-enter your passphrase so we can rebuild the encrypted backup. Old conversations remain readable on this device.',
    cta: 'Rebuild backup',
  },
};

export function E2EESetupGate() {
  const pathname = usePathname();
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  const { status, refresh } = useEnsureE2EE(isAuthenticated);

  const [passphrase, setPassphrase] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (authLoading) return null;
  if (!isAuthenticated) return null;
  if (!shouldGate(pathname)) return null;
  if (status.kind === 'loading' || status.kind === 'ok' || status.kind === 'unauthenticated')
    return null;

  if (status.kind === 'error') {
    return (
      <Overlay>
        <div className="rounded-2xl border bg-card p-6 shadow-lg text-card-foreground">
          <p className="text-sm font-medium">Could not check secure messaging status.</p>
          <p className="mt-2 text-xs text-muted-foreground">{status.message}</p>
          <Button onClick={refresh} className="mt-4 w-full">
            Retry
          </Button>
        </div>
      </Overlay>
    );
  }

  const copy = COPY[status.kind];

  const onSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (passphrase.length < 8) {
      setError('Use at least 8 characters.');
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await initializeE2EE(passphrase);
      setPassphrase('');
      refresh();
    } catch (e) {
      const message =
        e instanceof E2EERestoreError
          ? e.message
          : 'Could not finish secure messaging setup. Check your connection and try again.';
      setError(message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <Overlay>
      <form
        onSubmit={onSubmit}
        className="w-full max-w-md rounded-2xl border bg-card p-6 shadow-lg text-card-foreground"
      >
        <h2 className="text-lg font-semibold tracking-tight">{copy.title}</h2>
        <p className="mt-2 text-sm text-muted-foreground">{copy.body}</p>

        <div className="mt-5 space-y-2">
          <Label htmlFor="e2ee-passphrase" className="text-xs uppercase tracking-wide">
            Passphrase
          </Label>
          <Input
            id="e2ee-passphrase"
            type="password"
            autoComplete="current-password"
            autoFocus
            minLength={8}
            value={passphrase}
            onChange={(e) => setPassphrase(e.target.value)}
            disabled={busy}
            placeholder="At least 8 characters"
          />
        </div>

        {error && (
          <p className="mt-3 text-xs text-destructive" role="alert">
            {error}
          </p>
        )}

        <Button type="submit" className="mt-5 w-full" disabled={busy || passphrase.length < 8}>
          {busy ? 'Working…' : copy.cta}
        </Button>
      </form>
    </Overlay>
  );
}

function Overlay({ children }: { children: React.ReactNode }) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 p-4 backdrop-blur-sm"
      role="dialog"
      aria-modal="true"
    >
      {children}
    </div>
  );
}
