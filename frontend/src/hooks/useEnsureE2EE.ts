'use client';

/**
 * useEnsureE2EE — reconciles client-side E2EE state with the server.
 *
 * On mount (when authenticated) it asks `GET /auth/me` whether the server
 * has a public key on file, and inspects IndexedDB to see whether the
 * device holds a key pair. From that combination it produces one of:
 *
 *   - `ok`               — local keys present and server agrees. Nothing to do.
 *   - `needs-passphrase` — server has no public key, no local keys either.
 *                          The user must supply a passphrase so we can
 *                          generate + upload a fresh pair (first-time
 *                          OAuth signup, or a password user whose initial
 *                          upload silently failed).
 *   - `needs-restore`    — server has a backup but the device has no local
 *                          keys yet. Prompt for passphrase to decrypt the
 *                          backup (cross-device login).
 *   - `mismatch`         — local keys exist but the server forgot us
 *                          (backup wiped, public_key cleared). Prompt for
 *                          passphrase to rebuild the backup; we keep the
 *                          existing identity so old conversations remain
 *                          decryptable.
 *
 * The state is published verbatim so an `<E2EESetupGate />` can render the
 * matching prompt. After a successful `initializeE2EE()` the gate calls
 * `refresh()` and we transition to `ok`.
 */

import { useCallback, useEffect, useState } from 'react';
import { auth as authApi } from '@/lib/api';
import { getStoredKeyPair } from '@/lib/e2ee/keys';

export type E2EEStatus =
  | { kind: 'loading' }
  | { kind: 'unauthenticated' }
  | { kind: 'ok' }
  | { kind: 'needs-passphrase' }
  | { kind: 'needs-restore' }
  | { kind: 'mismatch' }
  | { kind: 'error'; message: string };

interface EnsureE2EEResult {
  status: E2EEStatus;
  refresh: () => void;
}

export function useEnsureE2EE(isAuthenticated: boolean): EnsureE2EEResult {
  const [status, setStatus] = useState<E2EEStatus>({ kind: 'loading' });
  const [tick, setTick] = useState(0);

  const refresh = useCallback(() => setTick((n) => n + 1), []);

  useEffect(() => {
    if (!isAuthenticated) {
      setStatus({ kind: 'unauthenticated' });
      return;
    }

    let cancelled = false;
    setStatus({ kind: 'loading' });

    (async () => {
      try {
        const [me, local] = await Promise.all([authApi.me(), getStoredKeyPair()]);
        if (cancelled) return;

        if (local && me.has_public_key) {
          setStatus({ kind: 'ok' });
          return;
        }
        if (local && !me.has_public_key) {
          setStatus({ kind: 'mismatch' });
          return;
        }
        if (!local && me.has_key_backup) {
          setStatus({ kind: 'needs-restore' });
          return;
        }
        setStatus({ kind: 'needs-passphrase' });
      } catch (e) {
        if (cancelled) return;
        const message = e instanceof Error ? e.message : 'Unknown error';
        setStatus({ kind: 'error', message });
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [isAuthenticated, tick]);

  // Re-evaluate when initializeE2EE() finishes elsewhere.
  useEffect(() => {
    if (typeof window === 'undefined') return;
    const handler = () => refresh();
    window.addEventListener('e2ee-keys-ready', handler);
    return () => window.removeEventListener('e2ee-keys-ready', handler);
  }, [refresh]);

  return { status, refresh };
}
