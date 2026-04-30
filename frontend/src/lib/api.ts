const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080/api/v1';

export function getPhotoUrl(userId: string): string {
  return `${API_BASE_URL}/files/users/${userId}/photo`;
}

interface ApiError {
  error: string;
  status: number;
}

class ApiClientError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = 'ApiClientError';
    this.status = status;
  }
}

function getAccessToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('access_token');
}

function getRefreshToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('refresh_token');
}

function setTokens(accessToken: string, refreshToken: string) {
  localStorage.setItem('access_token', accessToken);
  localStorage.setItem('refresh_token', refreshToken);
}

// Decode JWT payload without verifying signature (verification is server-side).
// This gives us the tamper-proof user_id: if the token is modified, the backend
// rejects it on the next API call, so we can trust the decoded claims for display.
function decodeTokenPayload(token: string): { user_id: string; email: string } | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;
    // JWT uses base64url encoding
    const base64 = parts[1].replace(/-/g, '+').replace(/_/g, '/');
    const payload = JSON.parse(atob(base64));
    if (!payload.user_id || !payload.email) return null;
    return { user_id: payload.user_id, email: payload.email };
  } catch {
    return null;
  }
}

function notifyAuthChange() {
  window.dispatchEvent(new Event('auth-change'));
}

function updateStoredUser(updates: Partial<UserInfo>) {
  try {
    const stored = localStorage.getItem('user');
    if (stored) {
      const user = JSON.parse(stored);
      Object.assign(user, updates);
      localStorage.setItem('user', JSON.stringify(user));
      notifyAuthChange();
    }
  } catch { /* ignore */ }
}

function clearAuth() {
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
  localStorage.removeItem('user');
  notifyAuthChange();
}

async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) return null;

  try {
    const res = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!res.ok) {
      clearAuth();
      return null;
    }

    const data = await res.json();
    setTokens(data.access_token, data.refresh_token);
    // Update display cache from server's fresh user data
    if (data.user) {
      localStorage.setItem('user', JSON.stringify(data.user));
      notifyAuthChange();
    }
    return data.access_token;
  } catch {
    clearAuth();
    return null;
  }
}

async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {},
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;
  const token = getAccessToken();

  const headers: Record<string, string> = {
    ...(options.headers as Record<string, string>),
  };

  // Don't set Content-Type for FormData (browser sets boundary automatically)
  if (!(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
  }

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  let res = await fetch(url, { ...options, headers });

  // If 401, try refreshing the token once
  if (res.status === 401 && token) {
    const newToken = await refreshAccessToken();
    if (newToken) {
      headers['Authorization'] = `Bearer ${newToken}`;
      res = await fetch(url, { ...options, headers });
    } else {
      clearAuth();
      if (typeof window !== 'undefined') {
        window.location.href = '/auth/signin';
      }
      throw new ApiClientError('Session expired. Please sign in again.', 401);
    }
  }

  if (res.status === 204) {
    return undefined as T;
  }

  const data = await res.json();

  if (!res.ok) {
    throw new ApiClientError(
      (data as ApiError).error || `Request failed with status ${res.status}`,
      res.status,
    );
  }

  return data as T;
}

// ─── Auth ────────────────────────────────────────────────────────────────────

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
  user: UserInfo;
}

export interface UserInfo {
  user_id: string;
  name: string;
  email: string;
  location: string | null;
  has_photo: boolean;
  is_public: boolean;
  public_key?: string;
  has_key_backup?: boolean;
}

export const auth = {
  login(email: string, password: string) {
    return apiRequest<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
  },

  register(data: { name: string; email: string; password: string; location?: string }) {
    return apiRequest<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  refresh(refreshToken: string) {
    return apiRequest<AuthResponse>('/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
  },

  logout() {
    return apiRequest<{ message: string }>('/auth/logout', { method: 'POST' });
  },

  me() {
    return apiRequest<{
      user_id: string;
      email: string;
      name: string;
      has_public_key: boolean;
      has_key_backup: boolean;
      email_verified: boolean;
    }>('/auth/me');
  },

  forgotPassword(email: string) {
    return apiRequest<{ message: string }>('/auth/forgot-password', {
      method: 'POST',
      body: JSON.stringify({ email }),
    });
  },

  resetPassword(token: string, newPassword: string) {
    return apiRequest<{ message: string }>('/auth/reset-password', {
      method: 'POST',
      body: JSON.stringify({ token, new_password: newPassword }),
    });
  },

  // Verify the email-verification token. Used by the /auth/verify page after
  // the user clicks the link in their inbox.
  verifyEmail(token: string) {
    return apiRequest<{ message: string }>(
      `/auth/verify-email?token=${encodeURIComponent(token)}`,
      { method: 'GET' },
    );
  },

  // Re-issue and re-send a verification email. Auth-required; rate-limited
  // to 3/hour per user on the server.
  resendVerification() {
    return apiRequest<{ message: string }>('/auth/resend-verification', {
      method: 'POST',
    });
  },
};

// ─── Users ───────────────────────────────────────────────────────────────────

export interface UserProfileResponse {
  user_id: string;
  name: string;
  email: string;
  location: string | null;
  has_photo: boolean;
  is_public: boolean;
  skills_offered: SkillResponse[];
  skills_wanted: SkillResponse[];
  created_at: string;
}

export interface SkillResponse {
  skill_id: string;
  name: string;
}

export interface SearchUsersResponse {
  users: UserProfileResponse[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export const users = {
  getProfile() {
    return apiRequest<UserProfileResponse>('/users/profile');
  },

  getPublicProfile(userId: string) {
    return apiRequest<UserProfileResponse>(`/public/users/${encodeURIComponent(userId)}`);
  },

  updateProfile(data: { name?: string; email?: string; location?: string; is_public?: boolean }) {
    return apiRequest<{ message: string }>('/users/profile', {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  searchPublic(params: { search_term?: string; location?: string; page?: number; limit?: number }) {
    const query = new URLSearchParams();
    if (params.search_term) query.set('search_term', params.search_term);
    if (params.location) query.set('location', params.location);
    if (params.page) query.set('page', String(params.page));
    if (params.limit) query.set('limit', String(params.limit));
    return apiRequest<SearchUsersResponse>(`/public/users/search?${query.toString()}`);
  },
};

// ─── Skills ──────────────────────────────────────────────────────────────────

export interface Skill {
  skill_id: string;
  name: string;
  created_at: string;
}

export const skills = {
  list() {
    return apiRequest<Skill[]>('/skills');
  },

  get(id: string) {
    return apiRequest<Skill>(`/skills/${encodeURIComponent(id)}`);
  },

  listOffered() {
    return apiRequest<SkillResponse[]>('/users/skills/offered');
  },

  addOffered(skillId: string) {
    return apiRequest<SkillResponse>('/users/skills/offered', {
      method: 'POST',
      body: JSON.stringify({ skill_id: skillId }),
    });
  },

  removeOffered(skillId: string) {
    return apiRequest<void>(`/users/skills/offered/${encodeURIComponent(skillId)}`, {
      method: 'DELETE',
    });
  },

  listWanted() {
    return apiRequest<SkillResponse[]>('/users/skills/wanted');
  },

  addWanted(skillId: string) {
    return apiRequest<SkillResponse>('/users/skills/wanted', {
      method: 'POST',
      body: JSON.stringify({ skill_id: skillId }),
    });
  },

  removeWanted(skillId: string) {
    return apiRequest<void>(`/users/skills/wanted/${encodeURIComponent(skillId)}`, {
      method: 'DELETE',
    });
  },
};

// ─── Swaps ───────────────────────────────────────────────────────────────────

export type SwapStatus = 'pending' | 'accepted' | 'rejected' | 'cancelled' | 'completed';

export interface SwapRequestResponse {
  swap_id: string;
  requester_id: string;
  responder_id: string;
  offered_skill_id: string;
  wanted_skill_id: string;
  status: SwapStatus;
  requester_completed: boolean;
  responder_completed: boolean;
  created_at: string;
  updated_at: string;
  requester: UserInfo;
  responder: UserInfo;
  offered_skill: SkillResponse;
  wanted_skill: SkillResponse;
}

export interface SwapListResponse {
  sent: SwapRequestResponse[];
  received: SwapRequestResponse[];
}

export const swaps = {
  create(data: {
    responder_id: string;
    offered_skill_id: string;
    wanted_skill_id: string;
  }) {
    return apiRequest<SwapRequestResponse>('/swaps', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  list(params?: { status?: SwapStatus; limit?: number; offset?: number }) {
    const query = new URLSearchParams();
    if (params?.status) query.set('status', params.status);
    if (params?.limit) query.set('limit', String(params.limit));
    if (params?.offset) query.set('offset', String(params.offset));
    const qs = query.toString();
    return apiRequest<SwapListResponse>(`/swaps${qs ? `?${qs}` : ''}`);
  },

  get(id: string) {
    return apiRequest<SwapRequestResponse>(`/swaps/${encodeURIComponent(id)}`);
  },

  updateStatus(id: string, status: 'accepted' | 'rejected' | 'cancelled') {
    return apiRequest<SwapRequestResponse>(`/swaps/${encodeURIComponent(id)}/status`, {
      method: 'PUT',
      body: JSON.stringify({ status }),
    });
  },

  delete(id: string) {
    return apiRequest<void>(`/swaps/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    });
  },

  matches() {
    return apiRequest<SwapRequestResponse[]>('/swaps/matches');
  },

  markComplete(id: string) {
    return apiRequest<import('../types/chat').SwapCompletionResponse>(
      `/chat/swaps/${encodeURIComponent(id)}/complete`,
      { method: 'PUT' },
    );
  },

  undoComplete(id: string) {
    return apiRequest<import('../types/chat').SwapCompletionResponse>(
      `/chat/swaps/${encodeURIComponent(id)}/complete`,
      { method: 'DELETE' },
    );
  },

  reportNoShow(id: string, reason?: string) {
    return apiRequest<{ swap_id: string; no_show_flag: boolean }>(
      `/swaps/${encodeURIComponent(id)}/no-show`,
      { method: 'POST', body: JSON.stringify({ reason: reason ?? '' }) },
    );
  },

  raiseDispute(id: string, reason: string) {
    return apiRequest<{ swap_id: string }>(
      `/swaps/${encodeURIComponent(id)}/dispute`,
      { method: 'POST', body: JSON.stringify({ reason }) },
    );
  },

  getReliability(userId: string) {
    return apiRequest<{
      qualifies: boolean;
      successful_swaps: number;
      reported_no_shows: number;
      score?: number;
    }>(`/users/${encodeURIComponent(userId)}/reliability`);
  },
};

// ─── Ratings ─────────────────────────────────────────────────────────────────

export interface RatingResponse {
  rating_id: string;
  swap_id: string;
  rater_id: string;
  ratee_id: string;
  score: number;
  comment: string | null;
  created_at: string;
  rater?: { user_id: string; name: string; has_photo: boolean };
  ratee?: { user_id: string; name: string; has_photo: boolean };
}

export interface UserRatingStats {
  average_rating: number;
  total_ratings: number;
}

export const ratings = {
  create(data: { swap_id: string; ratee_id: string; score: number; comment?: string }) {
    return apiRequest<RatingResponse>('/ratings', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  get(id: string) {
    return apiRequest<RatingResponse>(`/ratings/${encodeURIComponent(id)}`);
  },

  update(id: string, data: { score: number; comment?: string }) {
    return apiRequest<RatingResponse>(`/ratings/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  delete(id: string) {
    return apiRequest<void>(`/ratings/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    });
  },

  getForSwap(swapId: string) {
    return apiRequest<{ ratings: RatingResponse[] }>(`/ratings/swap/${encodeURIComponent(swapId)}`);
  },

  getForUser(userId: string, params?: { limit?: number; offset?: number }) {
    const query = new URLSearchParams();
    query.set('as_ratee', 'true');
    if (params?.limit) query.set('limit', String(params.limit));
    if (params?.offset) query.set('offset', String(params.offset));
    const qs = query.toString();
    return apiRequest<{ ratings: RatingResponse[] }>(
      `/users/${encodeURIComponent(userId)}/ratings${qs ? `?${qs}` : ''}`,
    ).then((res) => res.ratings);
  },

  getStatsForUser(userId: string) {
    return apiRequest<UserRatingStats>(`/users/${encodeURIComponent(userId)}/ratings/stats`);
  },
};

// ─── Notifications ───────────────────────────────────────────────────────────

export type NotificationType =
  | 'swap_request'
  | 'swap_accepted'
  | 'swap_rejected'
  | 'swap_completed'
  | 'new_rating'
  | 'skill_matched'
  | 'system_alert'
  | 'admin_notice';

export interface NotificationResponse {
  notification_id: string;
  type: NotificationType;
  title: string;
  message: string;
  is_read: boolean;
  related_id: string | null;
  created_at: string;
}

export interface NotificationListResponse {
  notifications: NotificationResponse[];
  pagination: {
    page: number;
    limit: number;
    total: number;
  };
}

export interface NotificationStats {
  total_notifications: number;
  unread_count: number;
  read_count: number;
}

export const notifications = {
  list(params?: { page?: number; limit?: number; unread_only?: boolean }) {
    const query = new URLSearchParams();
    if (params?.page) query.set('page', String(params.page));
    if (params?.limit) query.set('limit', String(params.limit));
    if (params?.unread_only) query.set('unread_only', 'true');
    const qs = query.toString();
    return apiRequest<NotificationListResponse>(`/notifications${qs ? `?${qs}` : ''}`);
  },

  get(id: string) {
    return apiRequest<NotificationResponse>(`/notifications/${encodeURIComponent(id)}`);
  },

  markRead(notificationIds: string[]) {
    return apiRequest<{ message: string }>('/notifications/mark-read', {
      method: 'PUT',
      body: JSON.stringify({ notification_ids: notificationIds }),
    });
  },

  markAllRead() {
    return apiRequest<{ message: string }>('/notifications/mark-all-read', {
      method: 'PUT',
    });
  },

  delete(id: string) {
    return apiRequest<{ message: string }>(
      `/notifications/${encodeURIComponent(id)}`,
      { method: 'DELETE' },
    );
  },

  stats() {
    return apiRequest<NotificationStats>('/notifications/stats');
  },
};

// ─── Availability ────────────────────────────────────────────────────────────

export interface AvailabilitySlot {
  slot_id: string;
  user_id: string;
  label: string;
  day_bitmask: number;
  start_time: string;
  end_time: string;
  created_at: string;
}

export const availability = {
  create(data: { label: string; day_bitmask: number; start_time: string; end_time: string }) {
    return apiRequest<AvailabilitySlot>('/availability', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  list() {
    return apiRequest<AvailabilitySlot[]>('/availability');
  },

  get(id: string) {
    return apiRequest<AvailabilitySlot>(`/availability/${encodeURIComponent(id)}`);
  },

  update(id: string, data: Partial<{ label: string; day_bitmask: number; start_time: string; end_time: string }>) {
    return apiRequest<AvailabilitySlot>(`/availability/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  delete(id: string) {
    return apiRequest<void>(`/availability/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    });
  },

  findCommon(userId: string) {
    return apiRequest<AvailabilitySlot[]>(`/availability/common/${encodeURIComponent(userId)}`);
  },
};

// ─── Files ───────────────────────────────────────────────────────────────────

export interface FileUploadResponse {
  filename: string;
  url: string;
  size: number;
  mime_type: string;
}

export const files = {
  uploadPhoto(file: File) {
    const formData = new FormData();
    formData.append('file', file);
    return apiRequest<FileUploadResponse>('/files/users/photo', {
      method: 'POST',
      body: formData,
    });
  },

  deletePhoto() {
    return apiRequest<{ message: string; success: boolean }>('/files/users/photo', {
      method: 'DELETE',
    });
  },
};

// ─── Chat ────────────────────────────────────────────────────────────────────

export const conversations = {
  list() {
    return apiRequest<{ conversations: import('../types/chat').ConversationListItem[] }>(
      '/chat/conversations',
    );
  },

  get(id: string) {
    return apiRequest<import('../types/chat').ConversationDetail>(
      `/chat/conversations/${encodeURIComponent(id)}`,
    );
  },

  getBySwap(swapId: string) {
    return apiRequest<import('../types/chat').ConversationDetail>(
      `/chat/conversations/by-swap/${encodeURIComponent(swapId)}`,
    );
  },

  getMessages(id: string, params?: { before?: string; limit?: number }) {
    const query = new URLSearchParams();
    if (params?.before) query.set('before', params.before);
    if (params?.limit) query.set('limit', String(params.limit));
    const qs = query.toString();
    return apiRequest<import('../types/chat').MessagesResponse>(
      `/chat/conversations/${encodeURIComponent(id)}/messages${qs ? `?${qs}` : ''}`,
    );
  },

  sendMessage(id: string, content: string, encrypted?: boolean, imageIds?: string[]) {
    return apiRequest<import('../types/chat').MessageType>(
      `/chat/conversations/${encodeURIComponent(id)}/messages`,
      {
        method: 'POST',
        body: JSON.stringify({
          content,
          encrypted,
          ...(imageIds?.length ? { image_ids: imageIds } : {}),
        }),
      },
    );
  },

  markRead(id: string) {
    return apiRequest<{ message: string }>(
      `/chat/conversations/${encodeURIComponent(id)}/read`,
      { method: 'PUT' },
    );
  },
};

export const messages = {
  edit(id: string, content: string) {
    return apiRequest<{ message: string }>(
      `/chat/messages/${encodeURIComponent(id)}`,
      { method: 'PUT', body: JSON.stringify({ content }) },
    );
  },

  delete(id: string) {
    return apiRequest<{ message: string }>(
      `/chat/messages/${encodeURIComponent(id)}`,
      { method: 'DELETE' },
    );
  },
};

export const chatImages = {
  upload(file: File | Blob, encrypted = false) {
    const formData = new FormData();
    formData.append('image', file);
    if (encrypted) formData.append('encrypted', 'true');
    return apiRequest<import('../types/chat').ChatImageUploadResponse>('/chat/images', {
      method: 'POST',
      body: formData,
    });
  },

  getUrl(imageId: string) {
    return `${API_BASE_URL}/chat/images/${imageId}`;
  },
};

export const chatUnread = {
  count() {
    return apiRequest<{ unread_count: number }>('/chat/unread-count');
  },
};

// ─── E2EE Keys ───────────────────────────────────────────────────────────────

export const e2eeKeys = {
  /** Upload public key and encrypted key backup after key generation. */
  upload(data: { public_key: string; encrypted_key_backup: string }) {
    return apiRequest<{ message: string }>('/users/me/e2ee-keys', {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  /** Fetch another user's public key for deriving the shared secret. */
  getPublicKey(userId: string) {
    return apiRequest<{ public_key: string }>(`/users/${encodeURIComponent(userId)}/public-key`);
  },

  /** Fetch the current user's encrypted key backup for cross-device restore. */
  getBackup() {
    return apiRequest<{ encrypted_key_backup: string }>('/users/me/key-backup');
  },
};

// ─── Video ───────────────────────────────────────────────────────────────────

export const video = {
  getToken(conversationId: string) {
    return apiRequest<{ token: string; url: string; room: string }>('/video/token', {
      method: 'POST',
      body: JSON.stringify({ conversation_id: conversationId }),
    });
  },
};

// ─── Exports ─────────────────────────────────────────────────────────────────

export const moderation = {
  block: (userId: string) =>
    apiRequest<{ message: string }>('/moderation/blocks', {
      method: 'POST',
      body: JSON.stringify({ user_id: userId }),
    }),
  unblock: (userId: string) =>
    apiRequest<{ message: string }>(`/moderation/blocks/${userId}`, { method: 'DELETE' }),
  listBlocks: () =>
    apiRequest<{ blocks: Array<{ blocked_id: string; created_at: string }> }>(
      '/moderation/blocks'
    ),
  createReport: (payload: {
    target_user_id: string;
    target_kind: 'user' | 'message' | 'swap';
    target_id?: string;
    reason: string;
  }) =>
    apiRequest<{ report_id: string }>('/moderation/reports', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),
  // Admin endpoints
  adminListReports: (params?: { status?: string; limit?: number; offset?: number }) => {
    const q = new URLSearchParams();
    if (params?.status) q.set('status', params.status);
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));
    const qs = q.toString();
    return apiRequest<{
      reports: Array<{
        report_id: string;
        reporter_id: string;
        target_user_id: string;
        target_kind: string;
        target_id: string | null;
        reason: string;
        status: string;
        created_at: string;
        resolved_at: string | null;
        resolver_id: string | null;
        resolution_note: string | null;
      }>;
      total: number;
      limit: number;
      offset: number;
    }>(`/admin/abuse-reports${qs ? `?${qs}` : ''}`);
  },
  adminResolveReport: (id: string, payload: { status: 'resolved' | 'dismissed' | 'reviewing'; note?: string }) =>
    apiRequest<{ message: string }>(`/admin/abuse-reports/${id}/resolve`, {
      method: 'POST',
      body: JSON.stringify(payload),
    }),
};

export const api = {
  auth,
  users,
  skills,
  swaps,
  ratings,
  notifications,
  availability,
  files,
  conversations,
  messages,
  chatImages,
  chatUnread,
  e2eeKeys,
  video,
  moderation,
};

export { ApiClientError, clearAuth, getAccessToken, setTokens, notifyAuthChange, updateStoredUser, decodeTokenPayload };
export default api;
