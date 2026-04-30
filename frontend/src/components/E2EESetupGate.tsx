'use client';

/**
 * E2EESetupGate — overlay modal that finishes E2EE setup when the
 * `useEnsureE2EE` hook reports it is incomplete.
 *
 * It is mounted once in the root layout but only appears on chat-bearing
 * routes (currently /messages and /swaps). For every other
 * route it stays out of the way so the rest of the app — browse, profile,
 * onboarding — remains usable even without keys.
 *
 * The user's account password is what we use to derive the local
 * encryption key (PBKDF2). Normally signup/login already do this in the
 * background, but if that step ever fails (or the device storage is
 * cleared) the user lands here and re-enters their account password.
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
    title: 'Finish setting up secure messaging',
    body: 'Re-enter your account password so we can generate your encryption keys on this device. This is the same password you used to sign up — we never send it to the server, it only derives a local key.',
    cta: 'Generate keys',
  },
  'needs-restore': {
    title: 'Unlock secure messaging on this device',
    body: 'Enter your account password so we can decrypt the keys backup we have on the server. Same password you sign in with — it stays on this device.',
    cta: 'Restore keys',
  },
  mismatch: {
    title: 'Resync secure messaging',
    body: 'Your device still has keys, but the server lost its backup. Re-enter your account password so we can rebuild the encrypted backup. Past conversations remain readable on this device.',
    cta: 'Rebuild backup',
  },
};

export function E2EESetupGate() {
  const pathname = usePathname();
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  const { status, refresh } = useEnsureE2EE(isAuthenticated);

  const [password, setPassword] = useState('');
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
    if (password.length < 8) {
      setError('Password must be at least 8 characters.');
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await initializeE2EE(password);
      setPassword('');
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
          <Label htmlFor="e2ee-password" className="text-xs uppercase tracking-wide">
            Account password
          </Label>
          <Input
            id="e2ee-password"
            type="password"
            autoComplete="current-password"
            autoFocus
            minLength={8}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            disabled={busy}
            placeholder="Same password you sign in with"
          />
        </div>

        {error && (
          <p className="mt-3 text-xs text-destructive" role="alert">
            {error}
          </p>
        )}

        <Button type="submit" className="mt-5 w-full" disabled={busy || password.length < 8}>
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
