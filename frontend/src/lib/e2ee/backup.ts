/**
 * E2EE Key Backup — Password-based encryption of the private key.
 *
 * Uses the user's account password as the passphrase:
 *   1. PBKDF2 (600,000 iterations, SHA-256) derives a 256-bit AES key from the password.
 *   2. AES-256-GCM encrypts the 32-byte X25519 secret key.
 *   3. Output: base64( salt[16] || iv[12] || ciphertext+tag[48] )
 *
 * The server stores this opaque blob. Without the password, it's infeasible to recover
 * the secret key. On login, the password is available client-side, so decryption is automatic.
 *
 * All operations use the Web Crypto API (no extra dependencies).
 */

import { encodeBase64, decodeBase64 } from 'tweetnacl-util';

// ── Constants ────────────────────────────────────────────────────────────────

const PBKDF2_ITERATIONS = 600_000;
const SALT_BYTES = 16;
const IV_BYTES = 12;

// ── Internal helpers ─────────────────────────────────────────────────────────

/** Import a password string as a CryptoKey for PBKDF2 derivation. */
async function importPassword(password: string): Promise<CryptoKey> {
  const enc = new TextEncoder();
  return crypto.subtle.importKey('raw', enc.encode(password), 'PBKDF2', false, [
    'deriveKey',
  ]);
}

/** Derive a 256-bit AES-GCM key from a password and salt via PBKDF2. */
async function deriveAESKey(password: string, salt: Uint8Array): Promise<CryptoKey> {
  const baseKey = await importPassword(password);
  return crypto.subtle.deriveKey(
    {
      name: 'PBKDF2',
      salt,
      iterations: PBKDF2_ITERATIONS,
      hash: 'SHA-256',
    },
    baseKey,
    { name: 'AES-GCM', length: 256 },
    false, // non-extractable
    ['encrypt', 'decrypt'],
  );
}

// ── Public API ───────────────────────────────────────────────────────────────

/**
 * Encrypt a secret key with the user's password.
 *
 * @returns base64-encoded blob: salt(16) + iv(12) + ciphertext+tag(48)
 */
export async function encryptKeyBackup(
  secretKey: Uint8Array,
  password: string,
): Promise<string> {
  const salt = crypto.getRandomValues(new Uint8Array(SALT_BYTES));
  const iv = crypto.getRandomValues(new Uint8Array(IV_BYTES));
  const aesKey = await deriveAESKey(password, salt);

  const ciphertext = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv },
    aesKey,
    secretKey,
  );

  // Concatenate: salt || iv || ciphertext (includes GCM auth tag)
  const blob = new Uint8Array(SALT_BYTES + IV_BYTES + ciphertext.byteLength);
  blob.set(salt, 0);
  blob.set(iv, SALT_BYTES);
  blob.set(new Uint8Array(ciphertext), SALT_BYTES + IV_BYTES);

  return encodeBase64(blob);
}

/**
 * Decrypt a secret key from a backup blob using the user's password.
 *
 * @returns The 32-byte X25519 secret key, or null if decryption fails
 *          (wrong password, corrupted data, etc.)
 */
export async function decryptKeyBackup(
  backupBlob: string,
  password: string,
): Promise<Uint8Array | null> {
  try {
    const blob = decodeBase64(backupBlob);
    if (blob.length < SALT_BYTES + IV_BYTES + 1) return null;

    const salt = blob.slice(0, SALT_BYTES);
    const iv = blob.slice(SALT_BYTES, SALT_BYTES + IV_BYTES);
    const ciphertext = blob.slice(SALT_BYTES + IV_BYTES);

    const aesKey = await deriveAESKey(password, salt);

    const plaintext = await crypto.subtle.decrypt(
      { name: 'AES-GCM', iv },
      aesKey,
      ciphertext,
    );

    const secretKey = new Uint8Array(plaintext);
    if (secretKey.length !== 32) return null;

    return secretKey;
  } catch {
    // Decryption failed — wrong password or corrupted backup
    return null;
  }
}
