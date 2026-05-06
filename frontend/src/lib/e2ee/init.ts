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

/**
 * Ensure the current device has an E2EE key pair and the server has the
 * public key + encrypted backup.
 *
 * This is intentionally fire-and-forget-safe: if it fails, the user can
 * still use the app — E2EE will just be unavailable until the next login.
 */
export async function initializeE2EE(password: string): Promise<void> {
  // 1. Already have keys locally?
  const local = await getStoredKeyPair();
  if (local) return;

  // 2. Try restoring from server backup
  try {
    const { encrypted_key_backup } = await e2eeApi.getBackup();
    if (encrypted_key_backup) {
      const secretKey = await decryptKeyBackup(encrypted_key_backup, password);
      if (secretKey) {
        const restored = keyPairFromSecretKey(secretKey);
        await storeKeyPair(restored);
        return;
      }
    }
  } catch {
    // No backup on server (404) or network error — fall through to generate
  }

  // 3. Generate fresh keys, encrypt backup, upload both
  const fresh = generateKeyPair();
  const backup = await encryptKeyBackup(fresh.secretKey, password);

  await e2eeApi.upload({
    public_key: encodePublicKey(fresh.publicKey),
    encrypted_key_backup: backup,
  });

  await storeKeyPair(fresh);
}
