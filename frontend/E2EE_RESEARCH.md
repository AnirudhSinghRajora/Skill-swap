# E2EE for SkillSwap Chat — Research & Options

## Current State

| Aspect | Details |
|---|---|
| **Message format** | HTML string (TipTap rich text) |
| **Transport** | WebSocket (primary) + REST fallback |
| **Storage** | PostgreSQL — `messages.content` as TEXT, `chat_images.image_data` as BYTEA |
| **Participants** | Always 1:1 (tied to a swap request) |
| **Auth** | JWT (HS256) in `Authorization` header (REST) / query param (WS) |
| **Images** | Uploaded separately, served publicly via UUID URL, linked to messages server-side |

**What gets encrypted:** `messages.content` (HTML), `chat_images.image_data` (binary).
**What stays plaintext:** metadata — `message_id`, `sender_id`, `conversation_id`, `created_at`, `has_images`, `is_edited`.

---

## Library / Approach Comparison

### Option A — Web Crypto API (Zero Dependencies)

Use the browser's native [`SubtleCrypto`](https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto) API. No npm packages.

| | |
|---|---|
| **Key exchange** | ECDH with P-256 curve (`generateKey`, `deriveKey`) |
| **Encryption** | AES-256-GCM (authenticated encryption) |
| **Key storage** | IndexedDB (private key never leaves the device) |
| **Go backend** | `crypto/ecdh`, `crypto/aes`, `crypto/cipher` (stdlib) |
| **Pros** | Zero dependencies, audited by browser vendors, fast (hardware-accelerated), no supply chain risk |
| **Cons** | Lower-level API (more code to wire up), P-256 curve (NSA-designed, debated), no forward secrecy without manual ratcheting |

**NPM packages:** None required.
**Go packages:** None required (stdlib only).

---

### Option B — TweetNaCl.js / libsodium.js

Audited, minimal NaCl (Networking and Cryptography Library) implementation.

| | |
|---|---|
| **Key exchange** | X25519 (Curve25519 ECDH) |
| **Encryption** | XSalsa20-Poly1305 (tweetnacl) or XChaCha20-Poly1305 (libsodium) |
| **Libraries (frontend)** | [`tweetnacl`](https://www.npmjs.com/package/tweetnacl) (7KB, audited) or [`libsodium-wrappers`](https://www.npmjs.com/package/libsodium-wrappers) (~190KB WASM) |
| **Libraries (backend)** | [`golang.org/x/crypto/nacl/box`](https://pkg.go.dev/golang.org/x/crypto/nacl/box) |
| **Pros** | Curve25519 is universally trusted, tiny bundle (tweetnacl), simple `box`/`open` API, well-audited |
| **Cons** | Extra dependency, no built-in forward secrecy, tweetnacl hasn't had updates since 2020 (stable, but some view it as a risk) |

```
npm install tweetnacl tweetnacl-util   # ~10KB total
# OR
npm install libsodium-wrappers         # ~190KB (WASM, faster)
```

```
go get golang.org/x/crypto             # nacl/box, curve25519
```

---

### Option C — Signal Protocol (Double Ratchet)

The gold standard for messaging E2EE — used by Signal, WhatsApp, Google Messages.

| | |
|---|---|
| **Key exchange** | X3DH (Extended Triple Diffie-Hellman) |
| **Encryption** | Double Ratchet (AES-256-CBC + HMAC-SHA256) |
| **Forward secrecy** | Yes — per-message keys, compromising one key doesn't reveal past/future |
| **Libraries (frontend)** | [`@nicolo-ribaudo/libsignal-protocol`](https://www.npmjs.com/package/@nicolo-ribaudo/libsignal-protocol) or [`@nicolo-ribaudo/signal-protocol`](https://www.npmjs.com/package/signal-protocol) |
| **Libraries (backend)** | No official Go implementation; would need [`signal-protocol`](https://github.com/nicolo-ribaudo/libsignal-protocol) with server acting as key distribution only |
| **Pros** | Gold standard, forward secrecy, future secrecy, battle-tested at WhatsApp scale |
| **Cons** | Significant complexity (prekey bundles, session management, ratchet state), overkill for 1:1 swap-based chat, limited Go ecosystem, harder to debug |

---

### Option D — Matrix Olm/Megolm (via Vodozemac)

Encryption layer used by Element/Matrix.

| | |
|---|---|
| **Libraries** | [`@nicolo-ribaudo/vodozemac`](https://www.npmjs.com/package/@nicolo-ribaudo/vodozemac) (Rust→WASM) |
| **Pros** | Designed for federated messaging, audited |
| **Cons** | Designed for Matrix protocol, heavy WASM bundle, overkill for our use case, poor Go support |

**Verdict:** Not recommended for this project.

---

## Recommendation: Option B (TweetNaCl / NaCl box)

**Why:**

1. **Simplicity** — `nacl.box(message, nonce, theirPublicKey, mySecretKey)` / `nacl.box.open(...)`. That's the entire API.
2. **Trusted curve** — Curve25519 / X25519 is universally recommended and avoids P-256 debates.
3. **Tiny bundle** — `tweetnacl` is 7KB minified. No WASM loading delay.
4. **Go stdlib adjacent** — `golang.org/x/crypto/nacl/box` is the official Go extended library.
5. **Right-sized** — Signal Protocol's forward secrecy is overkill for a skill-swap platform where conversations are tied to transient swap requests.
6. **Audited** — tweetnacl has a [published audit](https://tweetnacl.js.org/audits/cure53.pdf) by Cure53.

### If forward secrecy becomes a requirement later

Upgrade from static X25519 key pairs to a simplified ratchet:
- Rotate keys every N messages or every session
- Store previous keys briefly for decrypting in-flight messages
- This is a **non-breaking incremental upgrade** on top of Option B

---

## Key Concepts for Our Architecture

### Key Pair Lifecycle

```
User registers/first login
  → Generate X25519 key pair in browser
  → Store private key in IndexedDB (never sent to server)
  → Upload public key to server (GET /users/:id/public-key, PUT /users/:id/public-key)

User opens conversation with Bob
  → Fetch Bob's public key from server
  → Derive shared secret: X25519(myPrivate, bobPublic)
  → Encrypt messages with shared secret + random nonce
  → Send encrypted ciphertext over WebSocket/REST

Bob receives message
  → Fetch Alice's public key (cached)
  → Derive same shared secret: X25519(bobPrivate, alicePublic)
  → Decrypt ciphertext
```

### What Changes in the Message Flow

| Step | Before (plaintext) | After (E2EE) |
|---|---|---|
| **Frontend send** | `content: "<p>Hello</p>"` | `content: "base64(nonce + ciphertext)"` |
| **WebSocket transport** | Server reads content | Server sees opaque blob |
| **Database storage** | `content = '<p>Hello</p>'` | `content = 'base64(encrypted)'` |
| **Server relay** | Server can log/moderate | Server **cannot** read messages |
| **Frontend receive** | Render HTML directly | Decrypt → then render HTML |
| **Image upload** | Raw binary to server | Encrypted binary to server |
| **Image serve** | Server returns raw image | Server returns encrypted blob, frontend decrypts |

### Multi-Device Consideration

If a user logs in from a new device:
- **Problem:** Private key is in the old device's IndexedDB
- **Solution: Password-encrypted key backup (automatic)**
  1. On first key generation, encrypt private key with user's **account password** (PBKDF2 → AES-256-GCM)
  2. Upload encrypted backup blob to server (server cannot decrypt it — doesn't know the raw password)
  3. On new device login, password is already available (user just typed it) → auto-fetch backup → auto-decrypt → keys restored **invisibly**
  4. On password change, re-encrypt backup with new password atomically

  **Result:** Zero extra prompts. Zero extra passphrases. Fully transparent cross-device key recovery.

- **Edge case:** If backup decryption fails (e.g., corrupted data), fall back to generating a new key pair. Old messages become undecryptable on the new device (shown as "🔒 Encrypted with a previous key").

---

## Image Encryption Strategy

Images need special handling because they're uploaded separately before being sent in a message.

**Approach: Encrypt before upload**

1. Sender compresses image (existing flow)
2. Sender encrypts image bytes with the conversation's shared secret + random nonce
3. Upload encrypted blob to `POST /chat/images`
4. Server stores encrypted BYTEA (can't see the image)
5. Insert `<img data-encrypted-image-id="uuid">` into message HTML (not `<img src=...>`)
6. Recipient receives message, decrypts content HTML
7. For each `data-encrypted-image-id`, fetch encrypted blob from `/chat/images/:id`
8. Decrypt blob in browser → create `blob:` URL → render

**Trade-off:** Server can no longer generate thumbnails or transcode images. All processing must happen client-side (already the case — we use `browser-image-compression`).

---

## What the Server Still Knows (Metadata)

Even with E2EE, the server knows:
- Who is talking to whom (`sender_id`, `conversation_id`)
- When messages are sent (`created_at`)
- Message count and size
- Whether images are attached (`has_images`)
- Read receipts (`message_read_status`)

This is inherent to any E2EE system (Signal has the same limitation). Full metadata protection requires onion routing (Tor), which is out of scope.
