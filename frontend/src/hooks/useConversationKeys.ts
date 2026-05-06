'use client';

import { useState, useEffect, useRef } from 'react';
import { e2eeKeys as e2eeApi } from '@/lib/api';
import { decodePublicKey } from '@/lib/e2ee/keys';
import { deriveSharedKey } from '@/lib/e2ee/crypto';
import type { E2EEKeyPair } from '@/lib/e2ee/keys';

interface ConversationKeysState {
  /** The 32-byte precomputed shared key for this conversation. */
  sharedKey: Uint8Array | null;
  /** Whether the shared key derivation has completed. */
  isReady: boolean;
}

// In-memory cache of derived shared keys per other-user, survives re-renders.
// Keyed by the other user's ID — the shared key is the same for all
// conversations with the same user (same key pair).
const sharedKeyCache = new Map<string, Uint8Array>();

/**
 * Hook that derives and caches the NaCl shared key for a conversation.
 *
 * @param myKeyPair   Our local E2EE key pair (from useE2EEKeys)
 * @param otherUserId The ID of the other participant
 */
export function useConversationKeys(
  myKeyPair: E2EEKeyPair | null,
  otherUserId: string | undefined,
): ConversationKeysState {
  const [state, setState] = useState<ConversationKeysState>({
    sharedKey: null,
    isReady: false,
  });
  const initRef = useRef(false);

  useEffect(() => {
    if (!myKeyPair || !otherUserId) {
      setState({ sharedKey: null, isReady: true });
      return;
    }

    // Check cache first
    const cached = sharedKeyCache.get(otherUserId);
    if (cached) {
      setState({ sharedKey: cached, isReady: true });
      return;
    }

    if (initRef.current) return;
    initRef.current = true;

    let cancelled = false;

    (async () => {
      try {
        const { public_key } = await e2eeApi.getPublicKey(otherUserId);
        const theirPublicKey = decodePublicKey(public_key);

        if (cancelled) return;

        if (!theirPublicKey) {
          // Other user hasn't set up E2EE yet
          setState({ sharedKey: null, isReady: true });
          return;
        }

        const shared = deriveSharedKey(myKeyPair.secretKey, theirPublicKey);
        sharedKeyCache.set(otherUserId, shared);
        setState({ sharedKey: shared, isReady: true });
      } catch {
        if (!cancelled) {
          setState({ sharedKey: null, isReady: true });
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [myKeyPair, otherUserId]);

  return state;
}
