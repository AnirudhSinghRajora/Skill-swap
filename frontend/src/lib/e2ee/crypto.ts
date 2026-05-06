/**
 * E2EE Message Crypto — NaCl-based encrypt/decrypt for chat messages and images.
 *
 * Uses X25519 + XSalsa20-Poly1305 via TweetNaCl:
 *   1. Precompute a shared secret from (mySecretKey, theirPublicKey) via nacl.box.before()
 *   2. Encrypt with nacl.secretbox() using the shared secret + random 24-byte nonce
 *   3. Output: base64( nonce[24] || ciphertext )
 *
 * The shared secret is symmetric — both sides derive the same key:
 *   nacl.box.before(theirPublic, mySecret) === nacl.box.before(myPublic, theirSecret)
 */

import nacl from 'tweetnacl';
import { encodeBase64, decodeBase64, decodeUTF8, encodeUTF8 } from 'tweetnacl-util';

// ── Shared key derivation ────────────────────────────────────────────────────

/**
 * Derive a shared secret for a conversation.
 * Compute once per conversation, then reuse for all messages.
 *
 * @param mySecretKey  Our 32-byte X25519 private key
 * @param theirPublicKey  The other user's 32-byte X25519 public key
 * @returns 32-byte shared secret
 */
export function deriveSharedKey(
  mySecretKey: Uint8Array,
  theirPublicKey: Uint8Array,
): Uint8Array {
  return nacl.box.before(theirPublicKey, mySecretKey);
}

// ── Text encrypt/decrypt ─────────────────────────────────────────────────────

/**
 * Encrypt a text message (HTML content) with the shared key.
 *
 * @returns base64-encoded string: nonce(24) + ciphertext
 */
export function encrypt(plaintext: string, sharedKey: Uint8Array): string {
  const nonce = nacl.randomBytes(nacl.secretbox.nonceLength); // 24 bytes
  const messageBytes = decodeUTF8(plaintext);
  const ciphertext = nacl.secretbox(messageBytes, nonce, sharedKey);

  // Concatenate nonce + ciphertext
  const combined = new Uint8Array(nonce.length + ciphertext.length);
  combined.set(nonce, 0);
  combined.set(ciphertext, nonce.length);

  return encodeBase64(combined);
}

/**
 * Decrypt a text message from its base64 envelope.
 *
 * @returns The plaintext string, or null if decryption fails
 */
export function decrypt(encoded: string, sharedKey: Uint8Array): string | null {
  try {
    const combined = decodeBase64(encoded);
    if (combined.length <= nacl.secretbox.nonceLength) return null;

    const nonce = combined.slice(0, nacl.secretbox.nonceLength);
    const ciphertext = combined.slice(nacl.secretbox.nonceLength);

    const plaintext = nacl.secretbox.open(ciphertext, nonce, sharedKey);
    if (!plaintext) return null;

    return encodeUTF8(plaintext);
  } catch {
    return null;
  }
}

// ── Binary encrypt/decrypt (for images) ──────────────────────────────────────

/**
 * Encrypt binary data (e.g. image bytes) with the shared key.
 *
 * @returns Uint8Array: nonce(24) + ciphertext
 */
export function encryptBinary(data: Uint8Array, sharedKey: Uint8Array): Uint8Array {
  const nonce = nacl.randomBytes(nacl.secretbox.nonceLength);
  const ciphertext = nacl.secretbox(data, nonce, sharedKey);

  const combined = new Uint8Array(nonce.length + ciphertext.length);
  combined.set(nonce, 0);
  combined.set(ciphertext, nonce.length);

  return combined;
}

/**
 * Decrypt binary data from its encrypted envelope.
 *
 * @returns The plaintext bytes, or null if decryption fails
 */
export function decryptBinary(
  encrypted: Uint8Array,
  sharedKey: Uint8Array,
): Uint8Array | null {
  try {
    if (encrypted.length <= nacl.secretbox.nonceLength) return null;

    const nonce = encrypted.slice(0, nacl.secretbox.nonceLength);
    const ciphertext = encrypted.slice(nacl.secretbox.nonceLength);

    return nacl.secretbox.open(ciphertext, nonce, sharedKey) ?? null;
  } catch {
    return null;
  }
}
