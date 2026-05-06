# E2EE Implementation Plan — SWE Cycle

> **Prerequisite:** Read [E2EE_RESEARCH.md](./E2EE_RESEARCH.md) for library comparison and architecture rationale.
> **Chosen approach:** NaCl box (X25519 + XSalsa20-Poly1305) via `tweetnacl` (frontend) + `golang.org/x/crypto/nacl/box` (backend).

---

## Phase 0 — Setup & Dependencies

**Goal:** Install packages, add DB migration, create utility modules.

### Tasks

- [ ] **0.1** Install frontend packages
  ```bash
  npm install tweetnacl tweetnacl-util
  npm install -D @types/tweetnacl
  ```

- [ ] **0.2** Install backend package
  ```bash
  go get golang.org/x/crypto
  ```

- [ ] **0.3** Database migration — `005_add_e2ee_keys.sql`
  ```sql
  ALTER TABLE users
    ADD COLUMN public_key TEXT,              -- base64-encoded X25519 public key (32 bytes)
    ADD COLUMN encrypted_key_backup TEXT;    -- base64 blob: salt + nonce + AES-GCM(private key, password-derived key)

  CREATE INDEX idx_users_public_key_not_null
    ON users (id) WHERE public_key IS NOT NULL;
  ```

- [ ] **0.4** Update `User` model in Go (`internal/model/user.go`)
  - Add `PublicKey *string` field with GORM tags
  - Add `EncryptedKeyBackup *string` field with GORM tags

- [ ] **0.5** Update `User` type in TypeScript (`types/user.ts`)
  - Add `public_key?: string` field
  - Add `encrypted_key_backup?: string` field

### Deliverables
- Dependencies installed, migration ready, models updated
- No behavioral changes yet — fully backward compatible

---

## Phase 1 — Key Management (Frontend)

**Goal:** Generate, store, and upload key pairs. Fetch other users' public keys.

### Tasks

- [ ] **1.1** Create `lib/e2ee/keys.ts` — Key pair generation & storage
  ```
  generateKeyPair()        → { publicKey, secretKey }
  storeKeyPair(pair)       → IndexedDB write
  getStoredKeyPair()       → IndexedDB read | null
  exportPublicKey(key)     → base64 string
  ```
  - Use `tweetnacl.box.keyPair()`
  - Store in IndexedDB (not localStorage — binary-friendly, not cleared by "clear site data" in most browsers)
  - Use a dedicated `e2ee-keys` object store

- [ ] **1.1b** Create `lib/e2ee/backup.ts` — Password-based key backup
  ```
  encryptKeyBackup(secretKey, password)     → base64 string (salt + nonce + ciphertext)
  decryptKeyBackup(backupBlob, password)    → secretKey Uint8Array
  ```
  - Derive AES-256-GCM key from password using PBKDF2 (Web Crypto API, 600k iterations, SHA-256)
  - Output: `base64(salt[16] + nonce[12] + AES-GCM(secretKey))`
  - **The user's account password IS the passphrase** — no extra secret to remember

- [ ] **1.2** Create `lib/e2ee/crypto.ts` — Encrypt/decrypt primitives
  ```
  deriveSharedKey(mySecret, theirPublic) → Uint8Array
  encrypt(plaintext, sharedKey)          → base64 string (nonce + ciphertext)
  decrypt(encoded, sharedKey)            → plaintext string
  encryptBinary(data, sharedKey)         → Uint8Array (nonce + ciphertext)
  decryptBinary(data, sharedKey)         → Uint8Array
  ```
  - `nacl.box.before()` for shared key derivation (compute once per conversation, cache)
  - `nacl.secretbox()` with the precomputed shared key
  - Nonce: 24 random bytes, prepended to ciphertext

- [ ] **1.3** Create `hooks/useE2EEKeys.ts` — React hook for key lifecycle
  ```
  useE2EEKeys() → {
    isReady: boolean,
    publicKey: string | null,
    ensureKeysExist(): Promise<void>,  // generate + upload if missing
  }
  ```
  - On first call: check IndexedDB → if no keys, try auto-restore from server backup (see 1.5)
  - If no backup exists: generate new pair → upload public key + encrypted backup to server
  - Memoize with `useRef` to avoid re-generation

- [ ] **1.5** Integrate key backup into auth flow
  - **Sign up / first key generation:**
    1. Password is available in the sign-up form state
    2. Generate key pair
    3. `encryptKeyBackup(secretKey, password)` → backup blob
    4. Upload public key + encrypted backup to server in one call
    5. Store key pair in IndexedDB
  - **Login on new device (no keys in IndexedDB):**
    1. Password is available in the login form state
    2. After successful auth, fetch encrypted backup from `GET /users/me/key-backup`
    3. `decryptKeyBackup(backup, password)` → secret key restored
    4. Reconstruct full key pair (public key derived from secret key via `nacl.box.keyPair.fromSecretKey()`)
    5. Store in IndexedDB — **all transparent, zero extra prompts**
  - **Login on same device:** IndexedDB has keys → skip backup/restore entirely
  - **Password change:**
    1. Read existing key pair from IndexedDB
    2. `encryptKeyBackup(secretKey, newPassword)` → new backup blob
    3. Upload new backup to server (replaces old one)
    4. This must be done atomically with the password change API call

- [ ] **1.4** Create `hooks/useConversationKeys.ts` — Per-conversation shared key cache
  ```
  useConversationKey(conversationId, otherUserId) → {
    sharedKey: Uint8Array | null,
    isReady: boolean,
  }
  ```
  - Fetches other user's public key via API (React Query, cached)
  - Derives shared key with `nacl.box.before()`
  - Caches derived key in a `Map<conversationId, Uint8Array>` ref

### Deliverables
- Key pair generated on first chat usage
- Public key + encrypted backup uploaded to server
- Automatic key restore on login (new device) — zero friction
- Password change re-encrypts backup automatically
- Shared key derived per conversation
- All crypto operations in isolated, testable modules

---

## Phase 2 — Key Management (Backend)

**Goal:** Store and serve public keys. Backend never touches private keys.

### Tasks

- [ ] **2.1** Add API endpoints in `internal/user/handler.go`
  ```
  PUT  /users/me/e2ee-keys     — Upload public key + encrypted backup (atomic, authed)
  GET  /users/:id/public-key   — Fetch any user's public key (authed)
  GET  /users/me/key-backup    — Fetch own encrypted backup (authed)
  ```
  - Validate public key: base64-encoded, decodes to exactly 32 bytes
  - Encrypted backup: opaque blob, just validate non-empty string
  - Allow key rotation (new key pair + new backup replaces old)

- [ ] **2.2** Add routes in `internal/router/user_routes.go`

- [ ] **2.3** Add service methods in `internal/app/service/user_service.go`
  ```
  SetE2EEKeys(userID, publicKeyBase64, encryptedBackup) → error
  GetPublicKey(userID)        → string, error
  GetKeyBackup(userID)        → string, error
  UpdateKeyBackup(userID, encryptedBackup) → error  // for password change
  ```

- [ ] **2.4** Add repository methods in `internal/app/repository/user_repository.go`
  ```
  UpsertE2EEKeys(userID, publicKey, encryptedBackup) → error
  GetPublicKey(userID)     → string, error
  GetKeyBackup(userID)     → string, error
  UpdateKeyBackup(userID, backup) → error
  ```

- [ ] **2.5** Modify password change endpoint
  - Accept optional `encrypted_key_backup` field in password change request
  - If provided, update the backup atomically with the password hash

### Deliverables
- Public key + backup upload in one call
- Backup retrieval for auto-restore on new device
- Password change re-encrypts backup atomically
- No private key ever touches the server (only encrypted blob)

---

## Phase 3 — Encrypt Messages

**Goal:** Encrypt message content before sending, decrypt on receive.

### Tasks

- [ ] **3.1** Modify `useChat.ts` — `sendMessage()`
  - Before sending: `content = encrypt(htmlContent, sharedKey)`
  - The `content` field in WebSocket/REST payload becomes a base64 ciphertext blob
  - Add an `encrypted: true` flag to the message payload (for backward compat)

- [ ] **3.2** Modify `useChat.ts` — message receive handler
  - On `new_message` event: if `encrypted` flag is true, `content = decrypt(content, sharedKey)`
  - On initial message fetch (REST): decrypt each message's content
  - Decrypted HTML then flows into existing sanitize → render pipeline unchanged

- [ ] **3.3** Modify backend `Message` model
  - Add `Encrypted bool` field (default false for old messages)
  - Backend **does not** decrypt — just stores/relays the blob

- [ ] **3.4** Modify backend `handleSendMessage` in `ws_handler.go`
  - If `encrypted: true`, skip content validation (it's ciphertext, not HTML)
  - Still enforce max length on the encrypted blob (ciphertext ≈ plaintext + 40 bytes overhead)

- [ ] **3.5** Modify `ChatBubble.tsx` — graceful fallback
  - If decryption fails (wrong key, corrupted): show "🔒 Unable to decrypt this message" placeholder
  - Old plaintext messages (pre-E2EE) render normally

- [ ] **3.6** Add E2EE indicator to chat UI
  - Small lock icon (🔒) in the chat header when E2EE is active
  - Tooltip: "Messages are end-to-end encrypted"

### Deliverables
- Messages encrypted client-side before transmission
- Server stores opaque ciphertext
- Old plaintext messages still render
- Visual indicator for encrypted conversations

---

## Phase 4 — Encrypt Images

**Goal:** Encrypt image data client-side before upload, decrypt on display.

### Tasks

- [ ] **4.1** Modify `ChatInput.tsx` — image upload
  - After compression: `encryptedBlob = encryptBinary(imageBytes, sharedKey)`
  - Upload encrypted blob to `POST /chat/images`
  - Server stores encrypted BYTEA (cannot view the image)

- [ ] **4.2** Change image tag format in messages
  - Before: `<img src="/chat/images/uuid">`
  - After: `<img data-encrypted-image-id="uuid" alt="Encrypted image">`
  - This prevents the browser from auto-fetching (no `src` attribute)

- [ ] **4.3** Create `components/EncryptedImage.tsx`
  ```
  EncryptedImage({ imageId, sharedKey }) → renders decrypted image
  ```
  - Fetch encrypted blob from `/chat/images/:id`
  - Decrypt in browser → create `Blob` → `URL.createObjectURL()`
  - Show skeleton/spinner while decrypting
  - Revoke blob URL on unmount (memory cleanup)

- [ ] **4.4** Modify `ChatBubble.tsx` — render encrypted images
  - After decrypting message HTML, scan for `data-encrypted-image-id` attributes
  - Replace those elements with `<EncryptedImage>` React components
  - This requires switching from `dangerouslySetInnerHTML` to a parsed approach (or post-render DOM manipulation)

- [ ] **4.5** Update `ImageLightbox.tsx`
  - Accept `blob:` URLs (already works since `<img src="blob:...">` is valid)
  - Download button: use the blob directly instead of re-fetching

- [ ] **4.6** Backend: skip image linking regex for encrypted messages
  - In `SendMessage()`, if `encrypted: true`, skip the regex that extracts `/chat/images/uuid` from content (content is ciphertext, not HTML)
  - Instead, accept an optional `image_ids: []string` field in the message payload for explicit linking

### Deliverables
- Images encrypted before upload, decrypted in browser
- Server cannot see image contents
- Seamless UX with loading states
- Lightbox works with decrypted blob URLs

---

## Phase 5 — Testing & Hardening

### Tasks

- [ ] **5.1** Unit tests — `lib/e2ee/crypto.test.ts`
  - Encrypt → decrypt roundtrip (text)
  - Encrypt → decrypt roundtrip (binary/images)
  - Wrong key → decryption fails gracefully
  - Nonce uniqueness (encrypt same message twice → different ciphertext)
  - Empty content edge case
  - Large content (10KB HTML)

- [ ] **5.2** Integration tests
  - Full send/receive flow with two browser contexts (Playwright)
  - Key generation → upload → fetch → derive → encrypt → send → receive → decrypt
  - Image encrypt → upload → fetch → decrypt → render

- [ ] **5.3** Backward compatibility verification
  - Old plaintext messages render correctly
  - Mixed conversation (some encrypted, some not) displays properly
  - User without E2EE keys can still receive plaintext messages

- [ ] **5.4** Security review checklist
  - [ ] Private key never sent to server in plaintext (only encrypted backup blob)
  - [ ] Nonces are never reused (random 24 bytes each time)
  - [ ] Shared key derivation is deterministic (same inputs → same key)
  - [ ] Ciphertext in DB is indistinguishable from random
  - [ ] No timing side channels in decrypt path
  - [ ] IndexedDB key storage uses secure origin check
  - [ ] PBKDF2 iteration count is sufficient (≥600k)
  - [ ] Password change correctly re-encrypts backup
  - [ ] Backup decryption failure on login is handled gracefully (fallback: generate new keys)

- [ ] **5.5** Performance benchmarks
  - Measure encrypt/decrypt latency for typical message sizes
  - Measure image encrypt/decrypt for 1MB images
  - Ensure no perceptible UI lag (target: <5ms for text, <50ms for images)

---

## Rollout Strategy

### Feature Flag

```typescript
const E2EE_ENABLED = process.env.NEXT_PUBLIC_E2EE_ENABLED === 'true';
```

- Phase 0-2: Flag off (no visible changes)
- Phase 3-4: Flag on for development/testing
- Phase 5: Flag on for all users
- Post-launch: Remove flag, E2EE is default

### Backward Compatibility

| Scenario | Behavior |
|---|---|
| Both users have keys | Full E2EE |
| One user has keys, one doesn't | Messages sent unencrypted (fallback) |
| Old message (plaintext) | Rendered normally, no lock icon |
| New encrypted message | Lock icon, decrypted client-side |
| Key mismatch / corruption | "Unable to decrypt" placeholder |

### Migration Path

1. Deploy Phase 0-2 (backend changes, key infrastructure + backup endpoints)
2. Deploy Phase 3 with feature flag OFF
3. Enable flag for beta testers
4. Monitor error rates, performance
5. Enable for all users
6. Deploy Phase 4 (images)

> **Key backup is NOT a separate phase** — it's built into Phase 1-2 (key generation + auth flow).
> Users get automatic cross-device key recovery from day one with zero extra friction.

---

## File Map (New & Modified)

### New Files
```
frontend/src/lib/e2ee/keys.ts             — Key generation, IndexedDB storage
frontend/src/lib/e2ee/crypto.ts            — Encrypt/decrypt primitives
frontend/src/lib/e2ee/backup.ts            — Password-based key backup (encrypt/decrypt with account password)
frontend/src/hooks/useE2EEKeys.ts          — Key lifecycle hook (auto-restore on login)
frontend/src/hooks/useConversationKeys.ts  — Per-conversation shared key
frontend/src/components/EncryptedImage.tsx  — Decrypted image renderer
backend/skillswap/migrations/005_add_e2ee_keys.sql  — public_key + encrypted_key_backup columns
```

### Modified Files
```
frontend/package.json                      — tweetnacl dependency
frontend/src/types/user.ts                 — public_key field
frontend/src/hooks/useChat.ts              — encrypt on send, decrypt on receive
frontend/src/components/ChatInput.tsx       — encrypt images before upload
frontend/src/components/ChatBubble.tsx      — decrypt + render encrypted images
frontend/src/components/ChatWindow.tsx      — E2EE lock indicator
frontend/src/components/ImageLightbox.tsx   — support blob: URLs
backend/skillswap/go.mod                   — golang.org/x/crypto
backend/skillswap/internal/model/user.go   — PublicKey field
backend/skillswap/internal/user/handler.go — public key endpoints
backend/skillswap/internal/router/user_routes.go
backend/skillswap/internal/app/service/chat_service.go  — skip HTML regex for encrypted
backend/skillswap/internal/chat/handler.go              — skip content validation for encrypted
```

---

## Estimated Complexity per Phase

| Phase | Scope | Risk |
|---|---|---|
| 0 — Setup | Low | None |
| 1 — Frontend keys + auto-backup | Medium | IndexedDB browser compat, PBKDF2 in login flow |
| 2 — Backend keys + backup endpoints | Low | Password change atomicity |
| 3 — Encrypt messages | Medium | Backward compat, fallback logic |
| 4 — Encrypt images | High | DOM parsing change in ChatBubble, blob lifecycle |
| 5 — Testing | Medium | Playwright E2E test setup |

**Recommended order:** 0 → 1 → 2 → 3 → 5 (test text) → 4 → 5 (test images)
