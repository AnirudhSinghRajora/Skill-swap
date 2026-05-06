-- Migration 006: Create chat tables for real-time messaging
-- Conversations, Messages, Read Status, Chat Images

-- Conversations table (one per accepted swap)
CREATE TABLE IF NOT EXISTS conversations (
    conversation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    swap_id UUID NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_message_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT fk_conversations_swap_id
        FOREIGN KEY (swap_id) REFERENCES swap_requests(swap_id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_conversations_swap_id ON conversations(swap_id);
CREATE INDEX IF NOT EXISTS idx_conversations_deleted_at ON conversations(deleted_at);

-- Messages table
CREATE TABLE IF NOT EXISTS messages (
    message_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL,
    sender_id UUID NOT NULL,
    content TEXT NOT NULL,
    has_images BOOLEAN DEFAULT FALSE,
    is_edited BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT fk_messages_conversation_id
        FOREIGN KEY (conversation_id) REFERENCES conversations(conversation_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_messages_sender_id
        FOREIGN KEY (sender_id) REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(conversation_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_deleted_at ON messages(deleted_at);

-- Message read status (tracks last read time per user per conversation)
CREATE TABLE IF NOT EXISTS message_read_status (
    user_id UUID NOT NULL,
    conversation_id UUID NOT NULL,
    last_read_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (user_id, conversation_id),

    CONSTRAINT fk_message_read_status_user_id
        FOREIGN KEY (user_id) REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_message_read_status_conversation_id
        FOREIGN KEY (conversation_id) REFERENCES conversations(conversation_id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

-- Chat images table (binary storage, same pattern as profile photos)
CREATE TABLE IF NOT EXISTS chat_images (
    image_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID,
    uploader_id UUID NOT NULL,
    image_data BYTEA NOT NULL,
    mime_type VARCHAR(50) NOT NULL,
    file_size INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_chat_images_message_id
        FOREIGN KEY (message_id) REFERENCES messages(message_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_chat_images_uploader_id
        FOREIGN KEY (uploader_id) REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_chat_images_message_id ON chat_images(message_id);
CREATE INDEX IF NOT EXISTS idx_chat_images_uploader_id ON chat_images(uploader_id);

-- Add updated_at trigger for messages
DROP TRIGGER IF EXISTS update_messages_updated_at ON messages;
CREATE TRIGGER update_messages_updated_at
    BEFORE UPDATE ON messages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
