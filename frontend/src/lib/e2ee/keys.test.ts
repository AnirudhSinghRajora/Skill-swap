/**
 * Unit tests for E2EE key management — lib/e2ee/keys.ts
 *
 * Tests key generation, encoding/decoding, and key pair reconstruction.
 * IndexedDB operations can't be tested in Node — they are tested via
 * integration tests in the browser.
 */

import { describe, it, expect } from 'vitest';
import nacl from 'tweetnacl';
import { encodeBase64, decodeBase64 } from 'tweetnacl-util';
import {
  generateKeyPair,
  keyPairFromSecretKey,
  encodePublicKey,
  decodePublicKey,
} from '@/lib/e2ee/keys';

describe('generateKeyPair', () => {
  it('generates keys of correct length', () => {
    const kp = generateKeyPair();
    expect(kp.publicKey.length).toBe(32);
    expect(kp.secretKey.length).toBe(32);
  });

  it('generates unique key pairs each call', () => {
    const kp1 = generateKeyPair();
    const kp2 = generateKeyPair();
    expect(kp1.publicKey).not.toEqual(kp2.publicKey);
    expect(kp1.secretKey).not.toEqual(kp2.secretKey);
  });

  it('produces valid X25519 keys (can derive shared key)', () => {
    const alice = generateKeyPair();
    const bob = generateKeyPair();
    const shared1 = nacl.box.before(bob.publicKey, alice.secretKey);
    const shared2 = nacl.box.before(alice.publicKey, bob.secretKey);
    expect(shared1).toEqual(shared2);
  });
});

describe('keyPairFromSecretKey', () => {
  it('reconstructs the same public key from a secret key', () => {
    const original = generateKeyPair();
    const restored = keyPairFromSecretKey(original.secretKey);
    expect(restored.publicKey).toEqual(original.publicKey);
    expect(restored.secretKey).toEqual(original.secretKey);
  });

  it('produces a usable key pair for encryption', () => {
    const original = generateKeyPair();
    const restored = keyPairFromSecretKey(original.secretKey);
    const bob = generateKeyPair();

    // Encrypt with original, decrypt with restored — both use same keys
    const nonce = nacl.randomBytes(24);
    const msg = new TextEncoder().encode('test');
    const encrypted = nacl.box(msg, nonce, bob.publicKey, original.secretKey);
    const decrypted = nacl.box.open(encrypted, nonce, restored.publicKey, bob.secretKey);
    expect(decrypted).toEqual(msg);
  });
});

describe('encodePublicKey / decodePublicKey', () => {
  it('roundtrips a public key through base64 encoding', () => {
    const kp = generateKeyPair();
    const encoded = encodePublicKey(kp.publicKey);
    const decoded = decodePublicKey(encoded);
    expect(decoded).toEqual(kp.publicKey);
  });

  it('produces a valid base64 string', () => {
    const kp = generateKeyPair();
    const encoded = encodePublicKey(kp.publicKey);
    // Should be valid base64 (44 chars for 32 bytes with padding)
    expect(encoded.length).toBe(44);
    expect(() => decodeBase64(encoded)).not.toThrow();
  });

  it('returns null for invalid base64', () => {
    expect(decodePublicKey('not-valid-base64!!!')).toBeNull();
  });

  it('returns null for wrong-length key (too short)', () => {
    const short = encodeBase64(new Uint8Array(16));
    expect(decodePublicKey(short)).toBeNull();
  });

  it('returns null for wrong-length key (too long)', () => {
    const long = encodeBase64(new Uint8Array(64));
    expect(decodePublicKey(long)).toBeNull();
  });

  it('returns null for empty string', () => {
    expect(decodePublicKey('')).toBeNull();
  });
});
