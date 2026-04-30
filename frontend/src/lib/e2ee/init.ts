/**
 * E2EE initialisation for the auth flow.
 *
 * Called once after login/signup with the user's plaintext password and user
 * id. This ensures the private key exists locally for that account and a
 * backup is stored server-side.
 *
 * Flow:
 *   1. Check IndexedDB for an existing key pair scoped to this user id.
 *   2. Reconcile local public key with the server's public key.
 *   3. If absent, try to restore from the server backup (login on a new device).
 *   4. If no backup and the server has no key, generate + upload fresh keys.
 *
 * Important behaviour around password reset:
 *   When the user resets their password, the backend clears
 *   `encrypted_key_backup`, so step 2 returns an empty backup and we fall
 *   straight to step 3. Old ciphertext in past conversations is permanently
 *   unrecoverable — that's a fundamental property of password-derived key
 *   escrow without a separate recovery secret.
 *
 *   If a backup IS present but decryption fails (wrong password, corrupted
 *   blob, password changed without server-side clear), we now refuse to
 *   silently rotate. We throw — the caller surfaces an error and the user
 *   keeps a chance to re-enter the right password rather than instantly
 *   losing their old conversations.
 */

import {
  generateKeyPair,
  keyPairFromSecretKey,
  storeKeyPair,
  getStoredKeyPair,
  encodePublicKey,
  type E2EEKeyPair,
} from '@/lib/e2ee/keys';
import { encryptKeyBackup, decryptKeyBackup } from '@/lib/e2ee/backup';
import { auth as authApi, e2eeKeys as e2eeApi } from '@/lib/api';

export class E2EERestoreError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'E2EERestoreError';
  }
}

function notifyE2EEReady() {
  if (typeof window === 'undefined') return;
  window.dispatchEvent(new Event('e2ee-keys-ready'));
}

async function uploadKeyBackup(userId: string, password: string, keyPair: E2EEKeyPair) {
  const backup = await encryptKeyBackup(keyPair.secretKey, password);

  await e2eeApi.upload({
    public_key: encodePublicKey(keyPair.publicKey),
    encrypted_key_backup: backup,
  });

  await storeKeyPair(keyPair, userId);
}

async function getServerPublicKey(userId: string, hasPublicKey: boolean): Promise<string | null> {
  if (!hasPublicKey) return null;
  const { public_key: publicKey } = await e2eeApi.getPublicKey(userId);
  return publicKey || null;
}

export async function initializeE2EE(password: string, userId: string): Promise<void> {
  const me = await authApi.me();
  if (me.user_id !== userId) {
    throw new Error('Authenticated user changed during secure messaging setup. Please try again.');
  }

  const serverPublicKey = await getServerPublicKey(userId, me.has_public_key);

  // 1. Already have keys locally for this user?
  const local = await getStoredKeyPair(userId);
  if (local) {
    const localPublicKey = encodePublicKey(local.publicKey);
    if (serverPublicKey && serverPublicKey !== localPublicKey) {
      // The scoped local key does not match the server identity. Try the
      // backup path below rather than silently rotating or using wrong keys.
    } else {
      if (!serverPublicKey || !me.has_key_backup) {
        await uploadKeyBackup(userId, password, local);
      }
      notifyE2EEReady();
      return;
    }
  }

  // Legacy builds stored a single global `identity` record. Only migrate it
  // when its public key exactly matches the current user's server public key.
  if (serverPublicKey) {
    const legacy = await getStoredKeyPair();
    if (legacy && encodePublicKey(legacy.publicKey) === serverPublicKey) {
      if (!me.has_key_backup) {
        await uploadKeyBackup(userId, password, legacy);
      } else {
        await storeKeyPair(legacy, userId);
      }
      notifyE2EEReady();
      return;
    }
  }

  // 2. Try restoring from server backup
  let serverBackup: string | undefined;
  if (me.has_key_backup) {
    const resp = await e2eeApi.getBackup();
    serverBackup = resp.encrypted_key_backup || undefined;
  }

  if (serverBackup) {
    const secretKey = await decryptKeyBackup(serverBackup, password);
    if (secretKey) {
      const restored = keyPairFromSecretKey(secretKey);
      const restoredPublicKey = encodePublicKey(restored.publicKey);
      if (serverPublicKey && restoredPublicKey !== serverPublicKey) {
        throw new E2EERestoreError(
          'Secure messaging backup does not match this account. Please sign out and try again.',
        );
      }
      if (!serverPublicKey) {
        await uploadKeyBackup(userId, password, restored);
      } else {
        await storeKeyPair(restored, userId);
      }
      notifyE2EEReady();
      return;
    }
    // A backup exists but decryption failed. Do NOT silently rotate —
    // that would lose the user's identity key against the backup.
    throw new E2EERestoreError(
      'Could not unlock secure messaging with this password. ' +
        'If you recently reset your password, sign out and sign back in to regenerate keys.',
    );
  }

  if (serverPublicKey) {
    throw new E2EERestoreError(
      'Secure messaging is already set up for this account, but its encrypted backup is unavailable. ' +
        'Use a device that already has your keys to rebuild the backup.',
    );
  }

  // 3. No backup or public key on the server — generate fresh keys and upload.
  const fresh = generateKeyPair();
  await uploadKeyBackup(userId, password, fresh);
  notifyE2EEReady();
}
