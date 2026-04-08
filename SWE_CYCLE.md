# SWE Cycle: Google OAuth, Resend Email, LiveKit Video Calls

## Feature 1: Google OAuth (Server-Side Redirect Flow)

### Architecture
```
Frontend                    Backend                     Google
───────                    ───────                     ──────
Click "Google" btn ───→ GET /auth/google ──────────→ Google consent screen
                         (302 redirect)
                    ←── GET /auth/google/callback ←── Authorization code
                         Exchange code for user info
                         Find or create user
                         Generate JWT pair
                         302 → frontend /auth/callback?tokens=...
Parse tokens from URL
Store in localStorage
Redirect to /dashboard
```

### Backend Changes
1. **Config** — Add `GoogleClientID`, `GoogleClientSecret`, `GoogleRedirectURL`, `FrontendURL` env vars
2. **Auth Service** — Add `GoogleLogin(code string) (*AuthResponse, error)` method
   - Exchange auth code → Google tokens → fetch user profile
   - Find user by email (upsert: create if new, skip password requirement)
   - Generate JWT pair (same as regular login)
3. **Auth Handler** — Two new endpoints:
   - `GET /auth/google` → Redirect to Google consent URL with state param
   - `GET /auth/google/callback` → Exchange code, generate tokens, redirect to frontend
4. **User Model** — Add `AuthProvider` field (`local` | `google`) + `GoogleID`
5. **Migration** — `008_add_oauth_fields.sql`

### Frontend Changes
1. **Callback Page** — `src/app/auth/callback/page.tsx` parses tokens from URL, stores them, redirects
2. **Login/Signup Forms** — Already have Google button, just needs correct URL

### Dependencies
- `golang.org/x/oauth2` + `golang.org/x/oauth2/google`
- `google.golang.org/api/oauth2/v2`

---

## Feature 2: Resend Email API

### Architecture
```
Event triggers                EmailService              Resend API
─────────────               ────────────              ──────────
User registers ──────────→ SendWelcomeEmail() ──────→ POST api.resend.com/emails
Swap requested ──────────→ SendSwapRequestEmail()
Swap accepted/rejected ──→ SendSwapStatusEmail()
New message (offline) ───→ SendNewMessageEmail()
Password reset request ──→ SendPasswordResetEmail()
Email verification ──────→ SendVerificationEmail()
```

### Backend Changes
1. **Config** — Add `ResendAPIKey`, `FromEmail`
2. **Email Service** — `internal/email/service.go`
   - HTTP client wrapper for Resend REST API (no SDK needed — simple POST)
   - Template methods: Welcome, SwapRequest, SwapStatus, NewMessage, PasswordReset, EmailVerification
   - Async sending via goroutine (non-blocking)
3. **Hook into NotificationService** — After creating in-app notification, also fire email
4. **Password Reset Flow**:
   - New model: `PasswordResetToken` (token, user_id, expires_at, used)
   - `POST /auth/forgot-password` → generate token, send email
   - `POST /auth/reset-password` → validate token, update password
5. **Email Verification Flow**:
   - New model: `EmailVerification` (token, user_id, expires_at, verified)
   - On register → send verification email
   - `GET /auth/verify-email?token=...` → mark verified
   - Add `email_verified` field to User model
6. **Migration** — `008_add_email_fields.sql` (reset tokens table, verification, email_verified on users)

### Frontend Changes
1. **Forgot Password Page** — `src/app/auth/forgot-password/page.tsx`
2. **Reset Password Page** — `src/app/auth/reset-password/page.tsx`
3. **Email Verification Banner** — Show in dashboard if not verified
4. **Login Form** — Add "Forgot password?" link

### Dependencies
- None (Resend is a simple REST API — we'll use net/http)

---

## Feature 3: LiveKit Video Calls

### Architecture
```
User A (Caller)          Backend                 LiveKit Cloud         User B (Callee)
───────────────         ───────                 ─────────────        ───────────────
Click "Call" ──────→ POST /api/v1/calls/token                       
                     Validate participants                          
                     Generate room name                             
                  ←── { token, room_name } ←── LiveKit Server SDK   
                                                                    
Connect to LiveKit ────────────────────────→ JOIN room              
                                                                    
                     ──── WebSocket ────────────────────────────────→ call_invite
                                                                     Show ringing UI
                                                                     Click "Accept"
                                              ←──────── POST /api/v1/calls/token
                                              { token } ────────────→
                                                                     Connect to LiveKit
                                              ←── LiveKit handles ──→
                                                   media routing      media routing
```

### Backend Changes
1. **Config** — Add `LiveKitAPIKey`, `LiveKitAPISecret`, `LiveKitURL`
2. **Video Service** — `internal/app/service/video_service.go`
   - `CreateRoomToken(userID, conversationID) (token, roomName, error)`
   - Validates user is participant in conversation
   - Generates LiveKit access token with room join grants
3. **Video Handler** — `internal/video/handler.go`
   - `POST /api/v1/calls/token` → returns LiveKit token for a conversation
4. **WebSocket Signaling** — Add call signal types to ws_handler.go:
   - `call_invite` → forward to other participant
   - `call_accept` → forward to caller
   - `call_reject` → forward to caller
   - `call_hangup` → forward to other participant
5. **Routes** — `internal/router/video_routes.go`

### Frontend Changes
1. **VideoCall Component** — `src/components/VideoCall.tsx`
   - LiveKit Room component with local/remote video tracks
   - Controls: mute audio, mute video, screen share, hang up
2. **Incoming Call UI** — `src/components/IncomingCall.tsx`
   - Ring overlay with accept/reject
3. **useVideoCall Hook** — `src/hooks/useVideoCall.ts`
   - Manages call state machine: idle → ringing → connecting → connected → ended
   - Listens for WS call events
4. **ChatWindow Integration** — Add video call button to chat header
5. **Call Provider** — Global context for incoming call detection

### Dependencies
- Backend: `github.com/livekit/server-sdk-go` (token generation)
- Frontend: `@livekit/components-react`, `livekit-client`

---

## Implementation Order

1. **Google OAuth** (lowest risk, unblocks user growth)
2. **Resend Email** (depends on auth flow for verification)
3. **LiveKit Video** (most complex, builds on chat infrastructure)

## Environment Variables Needed
```env
# Google OAuth
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback
FRONTEND_URL=http://localhost:3000

# Resend
RESEND_API_KEY=
FROM_EMAIL=noreply@skillswap.com

# LiveKit
LIVEKIT_API_KEY=
LIVEKIT_API_SECRET=
LIVEKIT_URL=wss://your-app.livekit.cloud
```
