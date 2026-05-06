# Chat Feature — SWE Implementation Phases

> Reference: [CHAT_PLAN.md](./CHAT_PLAN.md) for full architectural context.

Each phase is a self-contained deliverable. Phases build on each other sequentially. Each phase ends with a working, testable state — no half-broken deploys.

---

## Phase 1: Database Schema + Models + Migration

**Goal**: Lay the data foundation. No UI, no WebSocket — just schema and Go structs.

### Tasks

1. **Create migration `006_create_chat_tables.sql`**
   - `conversations` table (conversation_id, swap_id UNIQUE, created_at, last_message_at, deleted_at)
   - `messages` table (message_id, conversation_id, sender_id, content TEXT, has_images, is_edited, created_at, updated_at, deleted_at)
   - `message_read_status` table (user_id, conversation_id, last_read_at — composite PK)
   - `chat_images` table (image_id, message_id, uploader_id, image_data BYTEA, mime_type, file_size, created_at)
   - All indexes as specified in plan

2. **Create migration `007_update_swap_completion.sql`**
   - Add `requester_completed BOOLEAN DEFAULT FALSE` to swap_requests
   - Add `responder_completed BOOLEAN DEFAULT FALSE` to swap_requests
   - Ensure status enum includes 'completed'

3. **Create Go model structs**
   - `internal/model/conversation.go`: Conversation, Message, MessageReadStatus, ChatImage structs with GORM tags
   - Update `internal/model/swap_request.go`: Add RequesterCompleted, ResponderCompleted fields

4. **Create repository layer**
   - `internal/app/repository/chat_repository.go`:
     - `CreateConversation(conv)`, `GetConversationBySwapID(swapID)`, `GetConversationByID(id)`
     - `GetUserConversations(userID)` — with last message + unread count via subquery
     - `CreateMessage(msg)`, `GetMessages(conversationID, before, limit)` — cursor-based pagination
     - `UpdateReadStatus(userID, conversationID)`, `GetUnreadCount(userID)`
     - `CreateChatImage(img)`, `GetChatImage(imageID)`
     - `DeleteMessage(messageID)`, `EditMessage(messageID, content)`
     - `IsParticipant(userID, conversationID)` — authorization check

5. **Create service layer**
   - `internal/app/service/chat_service.go`:
     - Wraps repository with business logic
     - `SendMessage`: validates participation, creates message, updates conversation.last_message_at
     - `GetConversationForSwap`: returns or creates conversation for an accepted swap
     - `MarkConversationRead`: updates read status
     - `EditMessage`: enforces 15-minute window + ownership
     - `DeleteMessage`: soft delete, ownership check
     - `UploadChatImage`: validate MIME, size, store binary

### Verification
- Run migration against dev database
- Write a simple Go test that creates a conversation, sends a message, retrieves it

### Files Created/Modified
```
backend/skillswap/migrations/006_create_chat_tables.sql        [NEW]
backend/skillswap/migrations/007_update_swap_completion.sql     [NEW]
backend/skillswap/internal/model/conversation.go                [NEW]
backend/skillswap/internal/model/swap_request.go                [MODIFIED]
backend/skillswap/internal/app/repository/chat_repository.go    [NEW]
backend/skillswap/internal/app/service/chat_service.go          [NEW]
```

---

## Phase 2: REST API Endpoints for Chat

**Goal**: Full REST API for chat — create conversations, send/receive messages, upload images. No WebSocket yet. This is also the polling fallback.

### Tasks

1. **Create chat handler**
   - `internal/chat/handler.go`:
     - `GetConversations` — GET /conversations (list with last message, unread count, other user's info)
     - `GetConversation` — GET /conversations/:id (details + participants)
     - `GetMessages` — GET /conversations/:id/messages (?before=timestamp&limit=50)
     - `SendMessage` — POST /conversations/:id/messages (body: {content})
     - `MarkRead` — PUT /conversations/:id/read
     - `EditMessage` — PUT /messages/:id (body: {content})
     - `DeleteMessage` — DELETE /messages/:id
     - `UploadImage` — POST /chat/images (multipart, returns image_id + url)
     - `ServeImage` — GET /chat/images/:id (serve binary)
     - All endpoints check participation via middleware or service layer

2. **Create chat routes**
   - `internal/router/chat_routes.go`:
     - All routes under auth middleware
     - Rate limit on SendMessage (30/min)
     - Rate limit on UploadImage (10/min)

3. **Wire into main router**
   - Update `internal/router/routes.go`: Initialize ChatRepository → ChatService → ChatHandler → SetupChatRoutes

4. **Update swap accept flow**
   - Modify `internal/app/service/swap_service.go`: When swap status changes to "accepted", auto-create a conversation
   - Add notification for conversation creation

5. **Add swap completion endpoints**
   - Modify `internal/swap/handler.go`: Add `MarkSwapComplete` endpoint
   - PUT /swaps/:id/complete — toggles requester_completed or responder_completed
   - When both true → status = "completed", create notification, prompt rating

6. **Update frontend API client**
   - Add to `frontend/src/lib/api.ts`:
     ```
     conversations.list()
     conversations.get(id)
     conversations.getMessages(id, { before?, limit? })
     conversations.sendMessage(id, { content })
     conversations.markRead(id)
     messages.edit(id, { content })
     messages.delete(id)
     chat.uploadImage(file) → { image_id, url }
     swaps.markComplete(id)
     ```

7. **Add frontend types**
   - `frontend/src/types/message.ts`:
     ```typescript
     interface Conversation { conversation_id, swap_id, other_user, last_message, unread_count, created_at }
     interface Message { message_id, conversation_id, sender_id, content, has_images, is_edited, created_at }
     interface ChatImage { image_id, url, mime_type }
     ```

### Verification
- Test all REST endpoints via curl/Postman
- Accept a swap → verify conversation auto-created
- Send messages → retrieve paginated history
- Upload image → verify served correctly
- Complete swap from both sides → verify status changes

### Files Created/Modified
```
backend/skillswap/internal/chat/handler.go                     [NEW]
backend/skillswap/internal/router/chat_routes.go               [NEW]
backend/skillswap/internal/router/routes.go                    [MODIFIED]
backend/skillswap/internal/app/service/swap_service.go         [MODIFIED]
backend/skillswap/internal/swap/handler.go                     [MODIFIED]
backend/skillswap/internal/model/swap_request.go               [MODIFIED]
frontend/src/lib/api.ts                                        [MODIFIED]
frontend/src/types/message.ts                                  [NEW]
```

---

## Phase 3: WebSocket Real-Time Layer

**Goal**: Add WebSocket support for instant message delivery, typing indicators, and online presence. REST endpoints from Phase 2 remain as fallback.

### Tasks

1. **Add gorilla/websocket dependency**
   - `go get github.com/gorilla/websocket`

2. **Create WebSocket hub**
   - `internal/chat/hub.go`:
     - `Hub` struct: connections map[string][]*Client (user_id → clients), register/unregister channels
     - `Run()` goroutine: handle register, unregister, and targeted send
     - `SendToUser(userID, message)` — sends to all of a user's connections
     - `SendToConversationParticipants(conversationID, message, excludeUserID)` — look up participants, send

3. **Create WebSocket client**
   - `internal/chat/client.go`:
     - `Client` struct: hub, conn, userID, send channel
     - `ReadPump()`: read messages from WS, parse type, route to handler
     - `WritePump()`: drain send channel to WS connection
     - Ping/pong heartbeat (30s interval, 10s timeout)

4. **Create WebSocket handler**
   - `internal/chat/ws_handler.go`:
     - `HandleWebSocket` — HTTP upgrade handler:
       - Extract token from query param
       - Validate JWT
       - Create Client, register with Hub
       - Launch read/write pumps
     - Message type routing:
       - `send_message` → validate participation, save to DB via ChatService, broadcast to other participant
       - `typing` / `stop_typing` → forward to other participant (no DB write)
       - `mark_read` → update read status via ChatService

5. **Wire WebSocket into router**
   - Add `GET /api/v1/ws` route (no auth middleware — auth handled in upgrade handler since WS can't send custom headers)
   - Hub initialized as singleton in routes.go, passed to handler

6. **Add rate limiting to WebSocket messages**
   - Per-client rate limiter: 30 messages/minute
   - If exceeded, send error message over WS and drop the message

### Verification
- Connect via WebSocket from browser console / wscat
- Send a message → verify other participant receives instantly
- Open two tabs → verify both receive messages
- Disconnect → verify reconnection works
- Exceed rate limit → verify error returned

### Files Created/Modified
```
backend/skillswap/go.mod                                       [MODIFIED]
backend/skillswap/internal/chat/hub.go                         [NEW]
backend/skillswap/internal/chat/client.go                      [NEW]
backend/skillswap/internal/chat/ws_handler.go                  [NEW]
backend/skillswap/internal/router/routes.go                    [MODIFIED]
backend/skillswap/internal/router/chat_routes.go               [MODIFIED]
```

---

## Phase 4: Chat UI — Messages Page

**Goal**: Build the full chat interface. Conversation list + chat window with message bubbles. No rich text yet — plain text first to nail the layout.

### Tasks

1. **Create WebSocket hook**
   - `frontend/src/lib/websocket.ts`:
     - `WebSocketClient` class: connect, disconnect, send, onMessage callbacks
     - Auto-reconnect with exponential backoff (1s → 2s → 4s → 8s → 30s max)
     - Heartbeat ping every 30s
   - `frontend/src/hooks/useWebSocket.ts`:
     - React hook wrapping WebSocketClient
     - Manages connection lifecycle with component mount/unmount
     - Exposes: `send()`, `isConnected`, `lastMessage`

2. **Create chat hooks**
   - `frontend/src/hooks/useChat.ts`:
     - Loads conversation messages via REST (initial + pagination)
     - Listens for new messages via WebSocket
     - Optimistic message sending (show immediately, confirm on WS ack)
     - Handles typing indicators (debounced 3s auto-stop)
     - Exposes: `messages`, `sendMessage()`, `loadMore()`, `isTyping`, `setTyping()`
   - `frontend/src/hooks/useUnreadCount.ts`:
     - Fetches total unread count
     - Updates on WebSocket `new_message` events
     - Exposes: `unreadCount`

3. **Build conversation list component**
   - `frontend/src/components/ConversationList.tsx`:
     - Fetches conversations via `api.conversations.list()`
     - Shows: other user's avatar + name, last message preview (truncated, strip HTML), unread badge, relative time
     - Click navigates to conversation
     - Highlights active conversation on desktop

4. **Build chat message bubble**
   - `frontend/src/components/ChatBubble.tsx`:
     - Own messages: right-aligned, primary color
     - Other's messages: left-aligned, muted color
     - Shows: content (plain text for now), timestamp, edited indicator
     - Context menu: Edit / Delete (own messages only)

5. **Build chat input**
   - `frontend/src/components/ChatInput.tsx`:
     - Simple textarea for Phase 4 (TipTap added in Phase 5)
     - Send on Enter (Shift+Enter for newline)
     - Typing indicator emission (debounced)
     - Send button

6. **Build typing indicator**
   - `frontend/src/components/TypingIndicator.tsx`:
     - Animated dots "User is typing..."
     - Shows below message list when other user is typing

7. **Build chat window**
   - `frontend/src/components/ChatWindow.tsx`:
     - Composes: message list + typing indicator + chat input
     - Header: other user's name/avatar, swap context (skills being exchanged)
     - Auto-scroll to bottom on new message
     - "Load older messages" button at top (cursor pagination)
     - Mark conversation as read when viewed

8. **Build messages page**
   - `frontend/src/app/messages/page.tsx`:
     - Desktop: split layout — ConversationList (left 1/3) + ChatWindow (right 2/3)
     - Mobile: show ConversationList, click navigates to `/messages/[conversationId]`
     - Empty state: "No conversations yet. Accept a swap to start chatting!"
   - `frontend/src/app/messages/[conversationId]/page.tsx`:
     - Full-screen ChatWindow for mobile
     - Back button to conversation list

9. **Add unread badge to Navigation**
   - Update `frontend/src/components/Navigation.tsx`:
     - Use `useUnreadCount` hook
     - Show red badge with count on Messages nav item

### Verification
- Accept a swap → conversation appears in list
- Send messages back and forth (via REST fallback if WS not yet wired)
- Verify WebSocket delivers messages instantly in two browser windows
- Typing indicator shows and disappears
- Unread counts update correctly
- Mobile layout works (conversation list → tap → chat view → back)

### Files Created/Modified
```
frontend/src/lib/websocket.ts                                  [NEW]
frontend/src/hooks/useWebSocket.ts                             [NEW]
frontend/src/hooks/useChat.ts                                  [NEW]
frontend/src/hooks/useUnreadCount.ts                           [NEW]
frontend/src/components/ConversationList.tsx                    [NEW]
frontend/src/components/ChatBubble.tsx                          [NEW]
frontend/src/components/ChatInput.tsx                           [NEW]
frontend/src/components/ChatWindow.tsx                          [NEW]
frontend/src/components/TypingIndicator.tsx                     [NEW]
frontend/src/app/messages/page.tsx                              [NEW]
frontend/src/app/messages/[conversationId]/page.tsx             [NEW]
frontend/src/components/Navigation.tsx                          [MODIFIED]
```

---

## Phase 5: Rich Text Editor (TipTap) + Image Upload

**Goal**: Replace the plain textarea with TipTap WYSIWYG editor. Add image upload in chat. Sanitize HTML on render.

### Tasks

1. **Install TipTap packages**
   ```bash
   npm install @tiptap/react @tiptap/starter-kit @tiptap/extension-link 
               @tiptap/extension-image @tiptap/extension-placeholder
               dompurify @types/dompurify
   ```

2. **Replace ChatInput with TipTap editor**
   - Update `frontend/src/components/ChatInput.tsx`:
     - TipTap editor instance with extensions: StarterKit (bold, italic, lists, code, blockquote, headings), Link, Image, Placeholder
     - Formatting toolbar: Bold, Italic, Link, Code, List (bulleted/ordered), Image upload button
     - Send on Ctrl/Cmd+Enter (Enter for newlines since rich text needs line breaks)
     - `editor.getHTML()` for message content
     - Image upload: click button → file picker → compress → upload via api.chat.uploadImage → insert image URL into editor

3. **Update ChatBubble for HTML rendering**
   - Update `frontend/src/components/ChatBubble.tsx`:
     - Import DOMPurify
     - Configure allowed tags: p, strong, em, a, code, pre, ul, ol, li, blockquote, h1-h3, img, br
     - Configure allowed attributes: href (on a), src (on img), alt (on img), class
     - Render via `dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(content, config) }}`
     - Style rendered HTML with Tailwind prose classes (`prose prose-sm dark:prose-invert`)
     - Images: lazy loading, click to expand, max-width constraint

4. **Style the TipTap editor**
   - Create `frontend/src/components/TipTapStyles.css` (or use Tailwind):
     - Editor area matches chat theme
     - Toolbar compact, icon-based
     - Bubble menu for link editing (select text → link button appears)
     - Mobile: toolbar scrollable horizontally

5. **Image upload flow**
   - Client: compress with `browser-image-compression` (already in deps), max 5MB/2048px
   - Upload: `POST /api/v1/chat/images` → returns `{ image_id, url }`
   - Insert: TipTap image extension inserts `<img src="/api/v1/chat/images/{id}">`
   - Display: on render, images show inline
   - Backend: serves from `chat_images` table, validates requester is conversation participant

### Verification
- Type bold text → verify renders bold for recipient
- Insert link → verify clickable
- Upload image → verify displays inline for both users
- Verify DOMPurify strips any script tags or malicious HTML
- Verify formatting toolbar works on mobile

### Files Created/Modified
```
frontend/package.json                                          [MODIFIED - new deps]
frontend/src/components/ChatInput.tsx                          [MODIFIED - TipTap]
frontend/src/components/ChatBubble.tsx                         [MODIFIED - HTML render]
```

---

## Phase 6: Swap Flow Revision + Integration Polish

**Goal**: Update the swap page for the new lifecycle. Connect chat to swaps. Dashboard integration. Final polish.

### Tasks

1. **Update swap page UI**
   - Modify `frontend/src/app/swaps/page.tsx`:
     - Accepted swaps: show "Open Chat" button → navigates to conversation
     - Accepted swaps: show "Mark as Complete" button
     - Show completion status: "You marked complete" / "Waiting for partner" / "Both completed ✓"
     - Completed swaps: show rating form (existing behavior, relocated to after completion)
     - Remove assumption that accepted = done

2. **Update swap status display**
   - New status badge colors: pending (yellow), accepted (blue), completed (green), rejected (red), cancelled (gray)
   - Accepted status text: "In Progress" (instead of implying done)

3. **Dashboard integration**
   - Modify `frontend/src/app/dashboard/page.tsx`:
     - Add "Recent Messages" card: last 3 conversations with preview
     - Update stats: add "Active Chats" count
     - Existing notification card: handle `new_message` notification type

4. **Notification integration**
   - Backend: Add `CreateNewMessageNotification(receiverID, senderName, conversationID)` to notification service
   - Only send if recipient has no active WebSocket connection (offline notification)
   - Frontend: Handle `new_message` notification type → link to `/messages?conversation=<id>`

5. **Auto-open chat on swap accept**
   - When a user accepts a swap, redirect to the newly created conversation
   - Show a toast: "Swap accepted! Start chatting to coordinate."

6. **Polish and edge cases**
   - Handle conversation with deleted/banned user
   - Handle case where swap is cancelled while chat is open
   - Empty message prevention (TipTap can produce empty `<p></p>`)
   - Message character limit (10,000 chars)
   - Conversation list sorting (most recent first)
   - Mark WebSocket reconnection state in UI (subtle banner: "Reconnecting...")

### Verification
- Full flow: Create swap → Accept → Chat opens → Exchange messages → Both mark complete → Rate
- Verify notifications arrive for offline users
- Verify dashboard shows recent messages
- Verify swap page shows correct completion state
- Mobile responsive throughout

### Files Created/Modified
```
frontend/src/app/swaps/page.tsx                                [MODIFIED]
frontend/src/app/dashboard/page.tsx                            [MODIFIED]
backend/skillswap/internal/app/service/notification_service.go [MODIFIED]
backend/skillswap/internal/notification/handler.go             [MODIFIED]
frontend/src/types/notification.ts                             [MODIFIED]
```

---

## Phase Summary

| Phase | Deliverable | Backend | Frontend | Testable? |
|-------|------------|---------|----------|-----------|
| 1 | Schema + Models + Repository | ✅ | — | DB + unit tests |
| 2 | REST API for chat | ✅ | API client + types | curl/Postman |
| 3 | WebSocket real-time | ✅ | — | wscat/browser console |
| 4 | Chat UI (plain text) | — | ✅ | Full chat flow |
| 5 | Rich text + images | — | ✅ | Formatting + uploads |
| 6 | Swap flow + integration | ✅ | ✅ | End-to-end flow |

### Dependency Graph

```
Phase 1 ─→ Phase 2 ─→ Phase 3
                │           │
                └─→ Phase 4 ←┘
                        │
                        ↓
                    Phase 5
                        │
                        ↓
                    Phase 6
```

Phase 4 can start after Phase 2 (using REST polling) and incorporate Phase 3's WebSocket when ready. Phases 5 and 6 are strictly sequential after Phase 4.
