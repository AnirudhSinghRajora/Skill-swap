/**
 * E2EE initialisation for the auth flow.
 *
 * Called once after login/signup with the user's plaintext password (which is
 * available in the form handler scope). This ensures the private key exists
 * locally and a backup is stored server-side.
 *
 * Flow:
 *   1. Check IndexedDB for an existing key pair.
 *   2. If absent, try to restore from the server backup (login on a new device).
 *   3. If no backup, generate new keys + backup + upload (first-time signup).
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
} from '@/lib/e2ee/keys';
import { encryptKeyBackup, decryptKeyBackup } from '@/lib/e2ee/backup';
import { e2eeKeys as e2eeApi } from '@/lib/api';

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

export async function initializeE2EE(password: string): Promise<void> {
  // 1. Already have keys locally?
  const local = await getStoredKeyPair();
  if (local) {
    notifyE2EEReady();
    return;
  }

  // 2. Try restoring from server backup
  let serverBackup: string | undefined;
  try {
    const resp = await e2eeApi.getBackup();
    serverBackup = resp.encrypted_key_backup || undefined;
  } catch {
    // 404 or network error — treat as "no backup", fall through to generation.
  }

  if (serverBackup) {
    const secretKey = await decryptKeyBackup(serverBackup, password);
    if (secretKey) {
      const restored = keyPairFromSecretKey(secretKey);
      await storeKeyPair(restored);
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

  // 3. No backup on the server — generate fresh keys and upload.
  const fresh = generateKeyPair();
  const backup = await encryptKeyBackup(fresh.secretKey, password);

  await e2eeApi.upload({
    public_key: encodePublicKey(fresh.publicKey),
    encrypted_key_backup: backup,
  });

  await storeKeyPair(fresh);
  notifyE2EEReady();
}
