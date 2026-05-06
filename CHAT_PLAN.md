# Chat Feature — Implementation Plan

## Problem Statement

The current swap flow is shallow: one user requests a swap, the other accepts, and we assume the swap is "done." There's no communication channel for users to coordinate schedules, discuss skill levels, share resources, or confirm completion. The notification system is one-way alerts only.

## Goals

1. Real-time messaging between two users once a swap is accepted
2. Rich text support (bold, italic, links, code blocks, lists)
3. Image sharing within chat
4. Chat tied to swap context (each accepted swap gets a conversation)
5. Revised swap lifecycle: accepted ≠ completed — users chat, then confirm completion
6. Online presence and typing indicators
7. Unread message counts in navigation

---

## Architecture Decisions

### Real-time Transport: WebSocket

**Choice**: `gorilla/websocket` on the backend, native `WebSocket` API on the frontend.

**Why not polling?** Polling at intervals (e.g., every 3s) creates unnecessary load, stale data, and poor UX. WebSocket gives instant delivery and enables typing indicators and presence.

**Why not SSE?** SSE is server→client only. Chat needs bidirectional communication (sending + receiving). WebSocket is the natural fit.

**Fallback**: If WebSocket connection fails, the frontend falls back to REST polling (GET messages endpoint) with a 5-second interval. This ensures chat works even behind restrictive proxies.

### Message Storage: PostgreSQL

Messages stored in PostgreSQL alongside existing data. No need for a separate message store at this scale. Conversations are scoped to swap requests.

### Rich Text: TipTap (Frontend)

**Choice**: [TipTap](https://tiptap.dev/) — a headless, extensible rich text editor built on ProseMirror.

**Why TipTap?**
- Headless = full control over styling (fits our Tailwind/shadcn design)
- Built-in extensions for bold, italic, links, code, lists, blockquotes
- Image extension for inline images
- Outputs HTML or JSON — we store HTML in DB, render safely on the receiving end
- Active ecosystem, MIT licensed
- Works with React 19

**Alternative considered**: `react-markdown` + toolbar. Simpler but requires users to know/type Markdown. TipTap gives WYSIWYG which is better UX.

### Message Content Sanitization

Messages stored as HTML (from TipTap). On render, we use `DOMPurify` to sanitize HTML before display. This prevents XSS from malicious message content.

### Image Uploads in Chat

Reuse the existing binary-in-PostgreSQL pattern (like profile photos). Chat images stored as `bytea` in a `chat_images` table. TipTap's image extension handles inline display. Images are uploaded via REST endpoint, and the returned URL is embedded in the message HTML.

**Size limit**: 5MB per image (same as profile photos), compressed client-side before upload.

---

## Database Schema

### New Tables

```sql
-- Conversations link to accepted swaps
CREATE TABLE conversations (
    conversation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    swap_id         UUID NOT NULL UNIQUE REFERENCES swap_requests(swap_id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_message_at TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_conversations_swap_id ON conversations(swap_id);

-- Messages within conversations
CREATE TABLE messages (
    message_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(conversation_id),
    sender_id       UUID NOT NULL REFERENCES users(user_id),
    content         TEXT NOT NULL,          -- HTML from TipTap
    has_images      BOOLEAN DEFAULT FALSE,
    is_edited       BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_messages_created_at ON messages(conversation_id, created_at DESC);

-- Track read status per user per conversation
CREATE TABLE message_read_status (
    user_id         UUID NOT NULL REFERENCES users(user_id),
    conversation_id UUID NOT NULL REFERENCES conversations(conversation_id),
    last_read_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, conversation_id)
);

-- Chat images (binary storage, same pattern as profile photos)
CREATE TABLE chat_images (
    image_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id      UUID REFERENCES messages(message_id),
    uploader_id     UUID NOT NULL REFERENCES users(user_id),
    image_data      BYTEA NOT NULL,
    mime_type       VARCHAR(50) NOT NULL,
    file_size       INTEGER NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chat_images_message_id ON chat_images(message_id);
```

### Schema Design Notes

- **`conversations.swap_id` is UNIQUE**: One conversation per swap. Simple, no ambiguity.
- **`message_read_status`**: Instead of per-message read receipts (expensive), we track the last time each user "read" a conversation. Any message after that timestamp is unread. Efficient for unread counts.
- **`messages.content` is HTML**: TipTap serializes to HTML. Sanitized on render with DOMPurify.
- **`chat_images` separate table**: Keeps the `messages` table lean. Images fetched on demand via their own endpoint.
- **Soft deletes**: `deleted_at` on conversations and messages for safety.

---

## Revised Swap Lifecycle

```
Current:   pending → accepted | rejected | cancelled
                     (assumed "done")

Proposed:  pending → accepted → completed
                   ↘ rejected    ↑
                   ↘ cancelled   |
                                 |
           User clicks "Accept" → conversation auto-created
           Both users chat, coordinate, do the swap
           Either user can mark as "completed"
           Both must confirm → status becomes "completed"
           Then rating prompt appears
```

### Changes to `swap_requests` table

```sql
ALTER TABLE swap_requests 
    ALTER COLUMN status TYPE VARCHAR(20);
-- Status values: pending, accepted, rejected, cancelled, completed

-- Track who has confirmed completion
ALTER TABLE swap_requests 
    ADD COLUMN requester_completed BOOLEAN DEFAULT FALSE,
    ADD COLUMN responder_completed BOOLEAN DEFAULT FALSE;
```

### Completion Logic

- When user A marks complete → their `_completed` flag = true, notification sent to user B
- When user B also marks complete → status changes to "completed", rating prompt shown
- Either user can "undo" their completion mark before both confirm
- This prevents one-sided "done" claims

---

## Backend API Design

### WebSocket Endpoint

```
GET /api/v1/ws — Upgrade to WebSocket
  Auth: JWT token passed as query param (?token=xxx) since WebSocket
        doesn't support custom headers in the browser API
  
  After connection:
  - Server authenticates token, associates connection with user_id
  - Client joins conversations they're part of
  - Server pushes: new messages, typing indicators, presence updates
  - Client sends: messages, typing events, read receipts
```

### WebSocket Message Protocol

```json
// Client → Server
{ "type": "send_message",    "conversation_id": "...", "content": "<p>Hello!</p>" }
{ "type": "typing",          "conversation_id": "..." }
{ "type": "stop_typing",     "conversation_id": "..." }
{ "type": "mark_read",       "conversation_id": "..." }

// Server → Client  
{ "type": "new_message",     "message": { ...full message object } }
{ "type": "user_typing",     "conversation_id": "...", "user_id": "..." }
{ "type": "user_stop_typing","conversation_id": "...", "user_id": "..." }
{ "type": "presence",        "user_id": "...", "online": true }
{ "type": "error",           "message": "..." }
```

### REST Endpoints (History + Fallback)

```
GET    /api/v1/conversations                    — List user's conversations (with last message, unread count)
GET    /api/v1/conversations/:id                — Get conversation details
GET    /api/v1/conversations/:id/messages       — Paginated message history (?before=timestamp&limit=50)
PUT    /api/v1/conversations/:id/read           — Mark conversation as read
POST   /api/v1/chat/images                      — Upload chat image (returns image URL)
GET    /api/v1/chat/images/:id                  — Serve chat image
DELETE /api/v1/messages/:id                      — Soft-delete own message
PUT    /api/v1/messages/:id                      — Edit own message (within 15 min)
```

---

## Backend Architecture

### WebSocket Hub Pattern

```
                    ┌─────────────┐
                    │   WS Hub    │
                    │             │
                    │ connections │ ← map[user_id]*Client
                    │ register    │ ← chan *Client
                    │ unregister  │ ← chan *Client
                    │ broadcast   │ ← chan Message
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
         ┌────┴────┐ ┌────┴────┐ ┌────┴────┐
         │ Client1 │ │ Client2 │ │ Client3 │
         │ (UserA) │ │ (UserB) │ │ (UserA) │  ← same user, multiple tabs
         └─────────┘ └─────────┘ └─────────┘
```

- **Hub**: Singleton, runs in its own goroutine. Manages all active connections.
- **Client**: One per WebSocket connection. Has read/write goroutines and a send channel.
- **Multi-tab support**: A user can have multiple connections. Messages sent to all of that user's clients.
- **Authorization**: Before delivering a message, Hub verifies the sender is a participant of the conversation.

### New Backend Packages

```
internal/
├── chat/
│   ├── handler.go        — REST endpoints (history, image upload)
│   ├── ws_handler.go     — WebSocket upgrade + message routing
│   ├── hub.go            — Connection registry + broadcast
│   └── client.go         — Per-connection read/write pumps
├── model/
│   ├── conversation.go   — Conversation + Message + ChatImage structs
│   └── (existing files)
├── app/
│   ├── repository/
│   │   └── chat_repository.go
│   └── service/
│       └── chat_service.go
└── router/
    └── chat_routes.go
```

---

## Frontend Architecture

### New Packages

| Package | Purpose |
|---------|---------|
| `@tiptap/react` | React integration for TipTap editor |
| `@tiptap/starter-kit` | Bold, italic, headings, lists, code, blockquote |
| `@tiptap/extension-link` | Clickable hyperlinks |
| `@tiptap/extension-image` | Inline images |
| `@tiptap/extension-placeholder` | Input placeholder text |
| `dompurify` | HTML sanitization for rendering messages |
| `@types/dompurify` | TypeScript types |

### New Files

```
frontend/src/
├── app/
│   └── messages/
│       ├── page.tsx              — Conversation list + chat view (split layout)
│       └── [conversationId]/
│           └── page.tsx          — Full-screen chat view (mobile)
├── components/
│   ├── ChatWindow.tsx            — Message list + input area
│   ├── ChatBubble.tsx            — Single message bubble (with sanitized HTML)
│   ├── ChatInput.tsx             — TipTap editor with toolbar
│   ├── ConversationList.tsx      — Sidebar list of conversations
│   └── TypingIndicator.tsx       — "User is typing..." animation
├── hooks/
│   ├── useWebSocket.ts           — WebSocket connection + reconnection logic
│   ├── useChat.ts                — Chat state management (messages, send, typing)
│   └── useUnreadCount.ts         — Global unread count for nav badge
├── lib/
│   └── websocket.ts              — WebSocket client class (connect, reconnect, parse)
└── types/
    └── message.ts                — Message, Conversation, ChatImage types
```

### WebSocket Client Design

```
┌──────────────────────────────────────┐
│         useWebSocket Hook            │
│                                      │
│  connect() → ws://host/api/v1/ws     │
│  onMessage → dispatch to handlers    │
│  send(type, payload)                 │
│  reconnect logic (exponential backoff│
│    1s → 2s → 4s → 8s → max 30s)     │
│  heartbeat (ping every 30s)          │
│  fallback → REST polling if WS fails │
└──────────────────────────────────────┘
```

### Chat UI Layout

```
Desktop (≥768px):                    Mobile (<768px):
┌──────────┬─────────────────┐       ┌─────────────────┐
│ Convos   │  Chat Window    │       │  Conversation    │
│ List     │                 │       │  List            │
│          │  [messages...]  │       │                  │
│ • User1  │                 │       │  → click →       │
│ • User2◉ │  ──────────     │       └─────────────────┘
│ • User3  │  [TipTap input] │       ┌─────────────────┐
│          │  [Send]         │       │  ← Back          │
└──────────┴─────────────────┘       │  Chat Window     │
                                     │  [messages...]   │
                                     │  ──────────      │
                                     │  [TipTap input]  │
                                     └─────────────────┘
```

---

## Security Considerations

1. **WebSocket Auth**: Token passed via query param (necessary for browser WS API). Token validated on upgrade. Connection closed if token expires — client must reconnect with fresh token.
2. **Message Sanitization**: All HTML content sanitized with DOMPurify before rendering. TipTap's output is already structured HTML, but we sanitize to prevent any injection.
3. **Authorization**: Users can only access conversations they're participants of. Hub checks participation before delivering messages. REST endpoints check ownership.
4. **Rate Limiting**: Max 30 messages per minute per user via WebSocket. Prevents spam.
5. **Image Validation**: Validate MIME type server-side (only image/*). Size limit 5MB. Client-side compression before upload.
6. **Message Editing Window**: Messages can only be edited within 15 minutes of sending. Prevents retroactive tampering.

---

## Integration Points

### Navigation Badge
`useUnreadCount` hook polls `GET /conversations` or uses WebSocket push to show total unread count on the Messages nav item.

### Swap Page
- Accepted swaps show a "Chat" button linking to `/messages?swap=<swap_id>`
- Accepted swaps show "Mark as Complete" button (replaces current assumption of done)
- Completed status requires both parties to confirm

### Notifications
- Existing notification system sends a notification when a new message arrives (if user is offline / WebSocket disconnected)
- Notification type: `new_message` with `related_id` = conversation_id

### Dashboard
- "Recent Messages" card showing last 3 conversations with previews
- Unread count in the sidebar/nav
