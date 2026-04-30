'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import {
  generateKeyPair,
  keyPairFromSecretKey,
  storeKeyPair,
  getStoredKeyPair,
  encodePublicKey,
  type E2EEKeyPair,
} from '@/lib/e2ee/keys';
import { encryptKeyBackup, decryptKeyBackup } from '@/lib/e2ee/backup';
import { e2eeKeys as e2eeApi } from '@/lib/api';

interface E2EEKeysState {
  /** Whether the key initialisation flow has completed (success or fail). */
  isReady: boolean;
  /** The local key pair, null if not yet loaded or unavailable. */
  keyPair: E2EEKeyPair | null;
}

/**
 * Hook that manages the E2EE key lifecycle:
 *
 * 1. On mount — try to load the key pair from IndexedDB.
 * 2. If absent + password provided — try to restore from server backup.
 * 3. If no backup exists — generate a fresh key pair, create backup, upload both.
 *
 * The password is only needed once per device:
 *   - On **login**: to decrypt the server-side backup.
 *   - On **signup**: to create the first backup.
 *
 * After the key pair is in IndexedDB, subsequent mounts (page navigations) load
 * the key from local storage without needing the password again.
 */
export function useE2EEKeys(password?: string) {
  const [state, setState] = useState<E2EEKeysState>({ isReady: false, keyPair: null });
  const initRef = useRef(false);

  const initKeys = useCallback(async () => {
    // 1. Check IndexedDB first
    const local = await getStoredKeyPair();
    if (local) {
      setState({ isReady: true, keyPair: local });
      return;
    }

    // Without a password we can't restore or create a backup
    if (!password) {
      setState({ isReady: true, keyPair: null });
      return;
    }

    // 2. Try to restore from server backup
    try {
      const { encrypted_key_backup } = await e2eeApi.getBackup();
      if (encrypted_key_backup) {
        const secretKey = await decryptKeyBackup(encrypted_key_backup, password);
        if (secretKey) {
          const restored = keyPairFromSecretKey(secretKey);
          await storeKeyPair(restored);
          setState({ isReady: true, keyPair: restored });
          return;
        }
        // Decryption failed — password mismatch (shouldn't happen if using account password).
        // Fall through to generate new keys.
      }
    } catch {
      // 404 or network error — no backup on server, fall through
    }

    // 3. Generate new key pair, back up, and upload
    const fresh = generateKeyPair();
    const backup = await encryptKeyBackup(fresh.secretKey, password);

    await e2eeApi.upload({
      public_key: encodePublicKey(fresh.publicKey),
      encrypted_key_backup: backup,
    });

    await storeKeyPair(fresh);
    setState({ isReady: true, keyPair: fresh });
  }, [password]);

  useEffect(() => {
    if (initRef.current) return;
    initRef.current = true;
    initKeys().catch(() => {
      setState({ isReady: true, keyPair: null });
    });
  }, [initKeys]);

  // If key setup completes later (e.g. auth flow still running), refresh from storage.
  useEffect(() => {
    if (typeof window === 'undefined') return;

    const onKeysReady = async () => {
      const local = await getStoredKeyPair();
      if (local) {
        setState({ isReady: true, keyPair: local });
      }
    };

    window.addEventListener('e2ee-keys-ready', onKeysReady);
    return () => window.removeEventListener('e2ee-keys-ready', onKeysReady);
  }, []);

  return state;
}
