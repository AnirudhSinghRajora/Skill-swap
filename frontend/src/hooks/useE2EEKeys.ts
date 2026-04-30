'use client';

import { useState, useEffect, useCallback } from 'react';
import { getStoredKeyPair, type E2EEKeyPair } from '@/lib/e2ee/keys';
import { initializeE2EE } from '@/lib/e2ee/init';

interface E2EEKeysState {
  /** Whether the key initialisation flow has completed (success or fail). */
  isReady: boolean;
  /** The local key pair, null if not yet loaded or unavailable. */
  keyPair: E2EEKeyPair | null;
}

/**
 * Hook that reads the current user's local E2EE identity key:
 *
 * 1. On mount — try to load the user-scoped key pair from IndexedDB.
 * 2. If absent + password provided — delegate to initializeE2EE(), which
 *    reconciles local storage with the server public key and backup state.
 *
 * The password is only needed once per device:
 *   - On **login**: to decrypt the server-side backup.
 *   - On **signup**: to create the first backup.
 *
 * After the key pair is in IndexedDB, subsequent mounts (page navigations) load
 * the key from local storage without needing the password again.
 */
export function useE2EEKeys(userId?: string, password?: string) {
  const [state, setState] = useState<E2EEKeysState>({ isReady: false, keyPair: null });

  const initKeys = useCallback(async () => {
    if (!userId) {
      setState({ isReady: true, keyPair: null });
      return;
    }

    const local = await getStoredKeyPair(userId);
    if (local) {
      setState({ isReady: true, keyPair: local });
      return;
    }

    // Without a password we can't restore or create a backup
    if (!password) {
      setState({ isReady: true, keyPair: null });
      return;
    }

    await initializeE2EE(password, userId);
    const refreshed = await getStoredKeyPair(userId);
    setState({ isReady: true, keyPair: refreshed });
  }, [password, userId]);

  useEffect(() => {
    initKeys().catch(() => {
      setState({ isReady: true, keyPair: null });
    });
  }, [initKeys]);

  // If key setup completes later (e.g. auth flow still running), refresh from storage.
  useEffect(() => {
    if (typeof window === 'undefined') return;

    const onKeysReady = async () => {
      if (!userId) return;
      const local = await getStoredKeyPair(userId);
      if (local) {
        setState({ isReady: true, keyPair: local });
      }
    };

    window.addEventListener('e2ee-keys-ready', onKeysReady);
    return () => window.removeEventListener('e2ee-keys-ready', onKeysReady);
  }, [userId]);

  return state;
}
