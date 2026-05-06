-- Drop the photo_url column from users table
-- The photo URL is now derived from the user ID: /api/v1/files/users/{user_id}/photo
ALTER TABLE users DROP COLUMN IF EXISTS photo_url;

-- Drop the index that referenced photo_url
DROP INDEX IF EXISTS idx_users_has_photo;
