// ── Conversation types ───────────────────────────────────────────────────────

export interface ConversationUser {
  user_id: string;
  name: string;
  has_photo: boolean;
}

export interface MessagePreview {
  content: string;
  sender_id: string;
  created_at: string;
}

export interface ConversationListItem {
  conversation_id: string;
  swap_id: string;
  other_user: ConversationUser;
  offered_skill: string;
  wanted_skill: string;
  last_message: MessagePreview | null;
  unread_count: number;
  created_at: string;
}

export interface ConversationDetail {
  conversation_id: string;
  swap_id: string;
  current_user: ConversationUser;
  other_user: ConversationUser;
  offered_skill: { skill_id: string; name: string };
  wanted_skill: { skill_id: string; name: string };
  swap_status: string;
  requester_completed: boolean;
  responder_completed: boolean;
  created_at: string;
}

// ── Message types ────────────────────────────────────────────────────────────

export interface MessageType {
  message_id: string;
  conversation_id: string;
  sender_id: string;
  content: string;
  encrypted?: boolean;
  has_images: boolean;
  is_edited: boolean;
  created_at: string;
  updated_at?: string;
  /** Client-only flag: set when an encrypted message could not be decrypted. */
  _decryptionFailed?: boolean;
}

export interface MessagesResponse {
  messages: MessageType[];
  has_more: boolean;
}

// ── Chat image types ─────────────────────────────────────────────────────────

export interface ChatImageUploadResponse {
  image_id: string;
  mime_type: string;
  file_size: number;
}

// ── Swap completion types ────────────────────────────────────────────────────

export interface SwapCompletionResponse {
  swap_id: string;
  status: string;
  requester_completed: boolean;
  responder_completed: boolean;
  requester: { user_id: string; name: string };
  responder: { user_id: string; name: string };
  offered_skill: { skill_id: string; name: string };
  wanted_skill: { skill_id: string; name: string };
}
