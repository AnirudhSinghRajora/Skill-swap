/**
 * Unit tests for E2EE key backup — lib/e2ee/backup.ts
 *
 * Tests password-based encryption/decryption of the secret key using
 * PBKDF2 + AES-256-GCM via the Web Crypto API.
 *
 * Note: Node.js has a global `crypto` with SubtleCrypto that mirrors
 * the browser API, so these tests run in Node without polyfills.
 */

import { describe, it, expect } from 'vitest';
import nacl from 'tweetnacl';
import { decodeBase64 } from 'tweetnacl-util';
import { encryptKeyBackup, decryptKeyBackup } from '@/lib/e2ee/backup';

describe('encryptKeyBackup / decryptKeyBackup', () => {
  it('roundtrips a secret key with the correct password', async () => {
    const kp = nacl.box.keyPair();
    const password = 'correct-horse-battery-staple';

    const backup = await encryptKeyBackup(kp.secretKey, password);
    const restored = await decryptKeyBackup(backup, password);

    expect(restored).toEqual(kp.secretKey);
  });

  it('returns null with the wrong password', async () => {
    const kp = nacl.box.keyPair();
    const backup = await encryptKeyBackup(kp.secretKey, 'right-password');
    const restored = await decryptKeyBackup(backup, 'wrong-password');

    expect(restored).toBeNull();
  });

  it('produces valid base64 output', async () => {
    const kp = nacl.box.keyPair();
    const backup = await encryptKeyBackup(kp.secretKey, 'test');

    expect(typeof backup).toBe('string');
    expect(() => decodeBase64(backup)).not.toThrow();
  });

  it('backup blob contains salt(16) + iv(12) + ciphertext+tag(48)', async () => {
    const kp = nacl.box.keyPair();
    const backup = await encryptKeyBackup(kp.secretKey, 'test');
    const decoded = decodeBase64(backup);

    // salt(16) + iv(12) + AES-GCM(32 bytes secret key + 16 bytes auth tag) = 76
    expect(decoded.length).toBe(76);
  });

  it('produces different backups each time (random salt and IV)', async () => {
    const kp = nacl.box.keyPair();
    const password = 'same-password';

    const backup1 = await encryptKeyBackup(kp.secretKey, password);
    const backup2 = await encryptKeyBackup(kp.secretKey, password);

    // Different salt + IV → different output
    expect(backup1).not.toBe(backup2);

    // Both decrypt to same key
    expect(await decryptKeyBackup(backup1, password)).toEqual(kp.secretKey);
    expect(await decryptKeyBackup(backup2, password)).toEqual(kp.secretKey);
  });

  it('handles empty password gracefully', async () => {
    const kp = nacl.box.keyPair();
    const backup = await encryptKeyBackup(kp.secretKey, '');
    const restored = await decryptKeyBackup(backup, '');
    expect(restored).toEqual(kp.secretKey);
  });

  it('returns null for corrupted backup blob', async () => {
    const kp = nacl.box.keyPair();
    const backup = await encryptKeyBackup(kp.secretKey, 'password');

    // Corrupt the ciphertext portion
    const decoded = decodeBase64(backup);
    decoded[decoded.length - 1] ^= 0xff;
    const corrupted = Buffer.from(decoded).toString('base64');

    const result = await decryptKeyBackup(corrupted, 'password');
    expect(result).toBeNull();
  });

  it('returns null for truncated backup', async () => {
    // Too short — less than salt(16) + iv(12) + 1 byte
    const short = Buffer.from(new Uint8Array(20)).toString('base64');
    const result = await decryptKeyBackup(short, 'password');
    expect(result).toBeNull();
  });

  it('returns null for invalid base64 input', async () => {
    const result = await decryptKeyBackup('!!not-base64!!', 'password');
    expect(result).toBeNull();
  });

  it('handles Unicode passwords', async () => {
    const kp = nacl.box.keyPair();
    const password = '密码パスワード🔐';

    const backup = await encryptKeyBackup(kp.secretKey, password);
    const restored = await decryptKeyBackup(backup, password);
    expect(restored).toEqual(kp.secretKey);
  });

  it('the restored key can be used for encryption', async () => {
    const alice = nacl.box.keyPair();
    const bob = nacl.box.keyPair();
    const password = 'backup-test';

    // Backup alice's key, then restore
    const backup = await encryptKeyBackup(alice.secretKey, password);
    const restoredSecret = await decryptKeyBackup(backup, password);
    expect(restoredSecret).not.toBeNull();

    // Use restored key to derive shared secret — must match original
    const originalShared = nacl.box.before(bob.publicKey, alice.secretKey);
    const restoredShared = nacl.box.before(bob.publicKey, restoredSecret!);
    expect(restoredShared).toEqual(originalShared);
  });
});
