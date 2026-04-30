/**
 * E2EE Key Management — IndexedDB storage for X25519 key pairs.
 *
 * The private key NEVER leaves the device except as a password-encrypted backup.
 * IndexedDB is used over localStorage because:
 *   - It handles binary data (Uint8Array) natively
 *   - It persists across sessions (survives logout)
 *   - It's origin-scoped (same-origin policy)
 */

import nacl from 'tweetnacl';
import { encodeBase64, decodeBase64 } from 'tweetnacl-util';

// ── Types ────────────────────────────────────────────────────────────────────

export interface E2EEKeyPair {
  publicKey: Uint8Array; // 32 bytes — shared with server + other users
  secretKey: Uint8Array; // 32 bytes — NEVER sent to server in plaintext
}

// ── IndexedDB helpers ────────────────────────────────────────────────────────

const DB_NAME = 'e2ee-keystore';
const DB_VERSION = 1;
const STORE_NAME = 'keys';
const LEGACY_RECORD_KEY = 'identity';

function recordKeyForUser(userId?: string): string {
  return userId ? `identity:${userId}` : LEGACY_RECORD_KEY;
}

function openDB(): Promise<IDBDatabase> {
  if (typeof globalThis.window !== 'undefined' && !globalThis.window.isSecureContext) {
    return Promise.reject(new Error('E2EE keys require a secure context (HTTPS)'));
  }
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION);

    request.onupgradeneeded = () => {
      const db = request.result;
      if (!db.objectStoreNames.contains(STORE_NAME)) {
        db.createObjectStore(STORE_NAME);
      }
    };

    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}

// ── Public API ───────────────────────────────────────────────────────────────

/** Generate a new X25519 key pair using TweetNaCl. */
export function generateKeyPair(): E2EEKeyPair {
  const kp = nacl.box.keyPair();
  return { publicKey: kp.publicKey, secretKey: kp.secretKey };
}

/** Reconstruct the full key pair from a secret key (e.g. after backup restore). */
export function keyPairFromSecretKey(secretKey: Uint8Array): E2EEKeyPair {
  const kp = nacl.box.keyPair.fromSecretKey(secretKey);
  return { publicKey: kp.publicKey, secretKey: kp.secretKey };
}

/** Store a key pair in IndexedDB, scoped to the authenticated user. */
export async function storeKeyPair(kp: E2EEKeyPair, userId?: string): Promise<void> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, 'readwrite');
    const store = tx.objectStore(STORE_NAME);
    // Store as plain object with Array copies (structured-cloneable)
    store.put(
      {
        publicKey: Array.from(kp.publicKey),
        secretKey: Array.from(kp.secretKey),
      },
      recordKeyForUser(userId),
    );
    tx.oncomplete = () => {
      db.close();
      resolve();
    };
    tx.onerror = () => {
      db.close();
      reject(tx.error);
    };
  });
}

/** Retrieve the stored key pair for a user, or null if none exists. */
export async function getStoredKeyPair(userId?: string): Promise<E2EEKeyPair | null> {
  try {
    const db = await openDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(STORE_NAME, 'readonly');
      const store = tx.objectStore(STORE_NAME);
      const request = store.get(recordKeyForUser(userId));
      request.onsuccess = () => {
        db.close();
        const record = request.result;
        if (!record?.publicKey || !record?.secretKey) {
          resolve(null);
          return;
        }
        resolve({
          publicKey: new Uint8Array(record.publicKey),
          secretKey: new Uint8Array(record.secretKey),
        });
      };
      request.onerror = () => {
        db.close();
        reject(request.error);
      };
    });
  } catch {
    // IndexedDB may be unavailable (private browsing in some browsers)
    return null;
  }
}

/** Delete the stored key pair for a user (used on account deletion or key rotation). */
export async function clearStoredKeyPair(userId?: string): Promise<void> {
  try {
    const db = await openDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(STORE_NAME, 'readwrite');
      tx.objectStore(STORE_NAME).delete(recordKeyForUser(userId));
      tx.oncomplete = () => {
        db.close();
        resolve();
      };
      tx.onerror = () => {
        db.close();
        reject(tx.error);
      };
    });
  } catch {
    // Silently fail if IndexedDB is unavailable
  }
}

// ── Encoding helpers ─────────────────────────────────────────────────────────

/** Encode a 32-byte public key to base64 for API transport. */
export function encodePublicKey(key: Uint8Array): string {
  return encodeBase64(key);
}

/** Decode a base64 public key from the API. Returns null on invalid input. */
export function decodePublicKey(b64: string): Uint8Array | null {
  try {
    const decoded = decodeBase64(b64);
    if (decoded.length !== 32) return null;
    return decoded;
  } catch {
    return null;
  }
}
