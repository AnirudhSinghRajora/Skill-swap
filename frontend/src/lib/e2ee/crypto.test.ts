/**
 * Unit tests for E2EE crypto primitives — lib/e2ee/crypto.ts
 *
 * Covers Phase 5 § 5.1:
 *   - Encrypt → decrypt roundtrip (text)
 *   - Encrypt → decrypt roundtrip (binary/images)
 *   - Wrong key → decryption fails gracefully
 *   - Nonce uniqueness (encrypt same message twice → different ciphertext)
 *   - Empty content edge case
 *   - Large content (10KB HTML)
 */

import { describe, it, expect } from 'vitest';
import nacl from 'tweetnacl';
import { decodeBase64 } from 'tweetnacl-util';
import {
  deriveSharedKey,
  encrypt,
  decrypt,
  encryptBinary,
  decryptBinary,
} from '@/lib/e2ee/crypto';

// ── Helpers ──────────────────────────────────────────────────────────────────

/** Create a pair of users and derive their shared keys. */
function createSharedKeys() {
  const alice = nacl.box.keyPair();
  const bob = nacl.box.keyPair();
  const aliceShared = deriveSharedKey(alice.secretKey, bob.publicKey);
  const bobShared = deriveSharedKey(bob.secretKey, alice.publicKey);
  return { alice, bob, aliceShared, bobShared };
}

// ── Tests ────────────────────────────────────────────────────────────────────

describe('deriveSharedKey', () => {
  it('produces the same shared key for both parties', () => {
    const { aliceShared, bobShared } = createSharedKeys();
    expect(aliceShared).toEqual(bobShared);
  });

  it('produces 32-byte keys', () => {
    const { aliceShared } = createSharedKeys();
    expect(aliceShared.length).toBe(32);
  });

  it('produces different keys for different key pairs', () => {
    const keys1 = createSharedKeys();
    const keys2 = createSharedKeys();
    expect(keys1.aliceShared).not.toEqual(keys2.aliceShared);
  });
});

describe('encrypt / decrypt (text)', () => {
  it('roundtrips a simple string', () => {
    const { aliceShared, bobShared } = createSharedKeys();
    const plaintext = 'Hello, World! 🌍';
    const ciphertext = encrypt(plaintext, aliceShared);
    const decrypted = decrypt(ciphertext, bobShared);
    expect(decrypted).toBe(plaintext);
  });

  it('roundtrips HTML content (like TipTap output)', () => {
    const { aliceShared, bobShared } = createSharedKeys();
    const html = '<p>Hello <strong>bold</strong> and <em>italic</em></p><ul><li>Item 1</li></ul>';
    const ciphertext = encrypt(html, aliceShared);
    const decrypted = decrypt(ciphertext, bobShared);
    expect(decrypted).toBe(html);
  });

  it('roundtrips Unicode / emoji content', () => {
    const { aliceShared, bobShared } = createSharedKeys();
    const text = '日本語テスト 🔐🎉 مرحبا Γεια σας';
    const ciphertext = encrypt(text, aliceShared);
    expect(decrypt(ciphertext, bobShared)).toBe(text);
  });

  it('handles empty string', () => {
    const { aliceShared, bobShared } = createSharedKeys();
    const ciphertext = encrypt('', aliceShared);
    const decrypted = decrypt(ciphertext, bobShared);
    expect(decrypted).toBe('');
  });

  it('handles large content (10KB HTML)', () => {
    const { aliceShared, bobShared } = createSharedKeys();
    // Generate ~10KB of HTML content
    const paragraph = '<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.</p>';
    const largeHtml = paragraph.repeat(80); // ~10KB
    expect(largeHtml.length).toBeGreaterThan(10_000);

    const ciphertext = encrypt(largeHtml, aliceShared);
    const decrypted = decrypt(ciphertext, bobShared);
    expect(decrypted).toBe(largeHtml);
  });

  it('returns null when decrypting with wrong key', () => {
    const { aliceShared } = createSharedKeys();
    const wrongKey = createSharedKeys().aliceShared; // Different key pair

    const ciphertext = encrypt('secret message', aliceShared);
    const result = decrypt(ciphertext, wrongKey);
    expect(result).toBeNull();
  });

  it('returns null for corrupted ciphertext', () => {
    const { aliceShared, bobShared } = createSharedKeys();
    const ciphertext = encrypt('test', aliceShared);

    // Corrupt the ciphertext by flipping a byte in the middle
    const decoded = decodeBase64(ciphertext);
    decoded[decoded.length - 5] ^= 0xff;
    const corrupted = Buffer.from(decoded).toString('base64');

    expect(decrypt(corrupted, bobShared)).toBeNull();
  });

  it('returns null for truncated ciphertext', () => {
    const { bobShared } = createSharedKeys();
    // Too short to contain nonce (24 bytes)
    const tooShort = Buffer.from(new Uint8Array(20)).toString('base64');
    expect(decrypt(tooShort, bobShared)).toBeNull();
  });

  it('returns null for empty base64 input', () => {
    const { bobShared } = createSharedKeys();
    expect(decrypt('', bobShared)).toBeNull();
  });

  it('produces unique ciphertext for the same plaintext (nonce uniqueness)', () => {
    const { aliceShared } = createSharedKeys();
    const plaintext = 'Same message encrypted twice';

    const ct1 = encrypt(plaintext, aliceShared);
    const ct2 = encrypt(plaintext, aliceShared);

    // Ciphertext must be different (random nonces)
    expect(ct1).not.toBe(ct2);

    // But both decrypt to the same plaintext
    expect(decrypt(ct1, aliceShared)).toBe(plaintext);
    expect(decrypt(ct2, aliceShared)).toBe(plaintext);
  });

  it('ciphertext is base64 encoded', () => {
    const { aliceShared } = createSharedKeys();
    const ciphertext = encrypt('test', aliceShared);
    // Verify it's valid base64
    expect(() => decodeBase64(ciphertext)).not.toThrow();
  });

  it('ciphertext starts with a 24-byte nonce', () => {
    const { aliceShared } = createSharedKeys();
    const ciphertext = encrypt('test', aliceShared);
    const decoded = decodeBase64(ciphertext);
    // Must be at least nonce (24) + ciphertext (>= 16 for Poly1305 tag)
    expect(decoded.length).toBeGreaterThan(24 + 16);
  });
});

describe('encryptBinary / decryptBinary', () => {
  it('roundtrips binary data', () => {
    const { aliceShared, bobShared } = createSharedKeys();
    const original = new Uint8Array([0, 1, 2, 3, 255, 128, 64, 32]);
    const encrypted = encryptBinary(original, aliceShared);
    const decrypted = decryptBinary(encrypted, bobShared);

    expect(decrypted).toEqual(original);
  });

  it('roundtrips a simulated 1KB image', () => {
    const { aliceShared, bobShared } = createSharedKeys();
    const imageData = new Uint8Array(1024);
    // Fill with pseudo-random data
    for (let i = 0; i < imageData.length; i++) {
      imageData[i] = (i * 17 + 31) % 256;
    }

    const encrypted = encryptBinary(imageData, aliceShared);
    const decrypted = decryptBinary(encrypted, bobShared);
    expect(decrypted).toEqual(imageData);
  });

  it('roundtrips empty binary data', () => {
    const { aliceShared, bobShared } = createSharedKeys();
    const empty = new Uint8Array(0);
    const encrypted = encryptBinary(empty, aliceShared);
    const decrypted = decryptBinary(encrypted, bobShared);
    expect(decrypted).toEqual(empty);
  });

  it('returns null for wrong key', () => {
    const { aliceShared } = createSharedKeys();
    const wrongKey = createSharedKeys().aliceShared;
    const data = new Uint8Array([1, 2, 3]);

    const encrypted = encryptBinary(data, aliceShared);
    expect(decryptBinary(encrypted, wrongKey)).toBeNull();
  });

  it('returns null for corrupted data', () => {
    const { aliceShared, bobShared } = createSharedKeys();
    const encrypted = encryptBinary(new Uint8Array([1, 2, 3]), aliceShared);

    // Corrupt
    encrypted[encrypted.length - 1] ^= 0xff;
    expect(decryptBinary(encrypted, bobShared)).toBeNull();
  });

  it('returns null for truncated data', () => {
    const { bobShared } = createSharedKeys();
    // Too short for nonce
    expect(decryptBinary(new Uint8Array(20), bobShared)).toBeNull();
  });

  it('encrypted binary is larger than plaintext (nonce + poly1305 overhead)', () => {
    const { aliceShared } = createSharedKeys();
    const plain = new Uint8Array(100);
    const encrypted = encryptBinary(plain, aliceShared);
    // nonce (24) + ciphertext (100 + 16 auth tag) = 140
    expect(encrypted.length).toBe(100 + 24 + 16);
  });

  it('produces unique output for the same input (nonce uniqueness)', () => {
    const { aliceShared } = createSharedKeys();
    const data = new Uint8Array([42, 42, 42]);

    const enc1 = encryptBinary(data, aliceShared);
    const enc2 = encryptBinary(data, aliceShared);

    // Different nonces → different output
    expect(enc1).not.toEqual(enc2);

    // Both decrypt successfully
    expect(decryptBinary(enc1, aliceShared)).toEqual(data);
    expect(decryptBinary(enc2, aliceShared)).toEqual(data);
  });
});

describe('cross-compatibility', () => {
  it('text encrypted by Alice can be decrypted by Bob and vice versa', () => {
    const { aliceShared, bobShared } = createSharedKeys();

    const msgFromAlice = encrypt('From Alice', aliceShared);
    const msgFromBob = encrypt('From Bob', bobShared);

    expect(decrypt(msgFromAlice, bobShared)).toBe('From Alice');
    expect(decrypt(msgFromBob, aliceShared)).toBe('From Bob');
  });

  it('binary encrypted by Alice can be decrypted by Bob and vice versa', () => {
    const { aliceShared, bobShared } = createSharedKeys();

    const dataA = new Uint8Array([10, 20, 30]);
    const dataB = new Uint8Array([40, 50, 60]);

    const encA = encryptBinary(dataA, aliceShared);
    const encB = encryptBinary(dataB, bobShared);

    expect(decryptBinary(encA, bobShared)).toEqual(dataA);
    expect(decryptBinary(encB, aliceShared)).toEqual(dataB);
  });
});
