-- Idempotent mock data for local UI testing
-- Seed users all use password: Password123!

BEGIN;

-- Ensure baseline skills exist (safe if they already exist)
INSERT INTO skills (skill_id, name) VALUES
  ('10000000-0000-0000-0000-000000000001', 'Go'),
  ('10000000-0000-0000-0000-000000000002', 'React'),
  ('10000000-0000-0000-0000-000000000003', 'UI Design'),
  ('10000000-0000-0000-0000-000000000004', 'Public Speaking')
ON CONFLICT (name) DO NOTHING;

-- Seed users
INSERT INTO users (
  user_id, name, email, password_hash, location, is_public,
  is_admin, is_banned, auth_provider, email_verified, created_at, updated_at
) VALUES
  (
    '00000000-0000-0000-0000-000000000001',
    'Alice Mentor',
    'alice@skillswap.dev',
    '$2a$10$u55CdCUchFBwP4Z8JfjseeesKEtiUrWxq5pDBc1pvGfXQUbfMx6aq',
    'Bengaluru',
    true,
    false,
    false,
    'local',
    true,
    NOW() - INTERVAL '20 days',
    NOW() - INTERVAL '1 day'
  ),
  (
    '00000000-0000-0000-0000-000000000002',
    'Bob Builder',
    'bob@skillswap.dev',
    '$2a$10$u55CdCUchFBwP4Z8JfjseeesKEtiUrWxq5pDBc1pvGfXQUbfMx6aq',
    'Mumbai',
    true,
    false,
    false,
    'local',
    true,
    NOW() - INTERVAL '15 days',
    NOW() - INTERVAL '1 day'
  ),
  (
    '00000000-0000-0000-0000-000000000003',
    'Charlie Coach',
    'charlie@skillswap.dev',
    '$2a$10$u55CdCUchFBwP4Z8JfjseeesKEtiUrWxq5pDBc1pvGfXQUbfMx6aq',
    'Delhi',
    true,
    false,
    false,
    'local',
    true,
    NOW() - INTERVAL '10 days',
    NOW() - INTERVAL '12 hours'
  )
ON CONFLICT (email) DO UPDATE SET
  name = EXCLUDED.name,
  password_hash = EXCLUDED.password_hash,
  location = EXCLUDED.location,
  is_public = EXCLUDED.is_public,
  auth_provider = EXCLUDED.auth_provider,
  email_verified = EXCLUDED.email_verified,
  updated_at = NOW();

-- User offered skills
INSERT INTO user_skills_offered (user_id, skill_id)
SELECT '00000000-0000-0000-0000-000000000001', skill_id FROM skills WHERE name = 'Go'
ON CONFLICT DO NOTHING;
INSERT INTO user_skills_offered (user_id, skill_id)
SELECT '00000000-0000-0000-0000-000000000002', skill_id FROM skills WHERE name = 'React'
ON CONFLICT DO NOTHING;
INSERT INTO user_skills_offered (user_id, skill_id)
SELECT '00000000-0000-0000-0000-000000000003', skill_id FROM skills WHERE name = 'Public Speaking'
ON CONFLICT DO NOTHING;

-- User wanted skills
INSERT INTO user_skills_wanted (user_id, skill_id)
SELECT '00000000-0000-0000-0000-000000000001', skill_id FROM skills WHERE name = 'UI Design'
ON CONFLICT DO NOTHING;
INSERT INTO user_skills_wanted (user_id, skill_id)
SELECT '00000000-0000-0000-0000-000000000002', skill_id FROM skills WHERE name = 'Go'
ON CONFLICT DO NOTHING;
INSERT INTO user_skills_wanted (user_id, skill_id)
SELECT '00000000-0000-0000-0000-000000000003', skill_id FROM skills WHERE name = 'React'
ON CONFLICT DO NOTHING;

-- Availability slots
INSERT INTO availability_slots (
  slot_id, user_id, label, day_bitmask, start_time, end_time, created_at
) VALUES
  (
    '20000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'Weekday Evenings',
    62,
    '18:00',
    '21:00',
    NOW() - INTERVAL '9 days'
  ),
  (
    '20000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000002',
    'Weekend Mornings',
    96,
    '09:00',
    '12:00',
    NOW() - INTERVAL '8 days'
  )
ON CONFLICT (slot_id) DO NOTHING;

-- Seed one accepted swap between Alice and Bob
INSERT INTO swap_requests (
  swap_id,
  requester_id,
  responder_id,
  offered_skill_id,
  wanted_skill_id,
  status,
  created_at,
  updated_at
)
SELECT
  '30000000-0000-0000-0000-000000000001',
  '00000000-0000-0000-0000-000000000002',
  '00000000-0000-0000-0000-000000000001',
  s1.skill_id,
  s2.skill_id,
  'accepted'::swap_status,
  NOW() - INTERVAL '6 days',
  NOW() - INTERVAL '2 hours'
FROM skills s1, skills s2
WHERE s1.name = 'React' AND s2.name = 'Go'
ON CONFLICT (swap_id) DO UPDATE SET
  status = EXCLUDED.status,
  updated_at = NOW();

-- Chat conversation linked to the accepted swap
INSERT INTO conversations (conversation_id, swap_id, created_at, last_message_at)
VALUES (
  '40000000-0000-0000-0000-000000000001',
  '30000000-0000-0000-0000-000000000001',
  NOW() - INTERVAL '5 days',
  NOW() - INTERVAL '90 minutes'
)
ON CONFLICT (swap_id) DO UPDATE SET
  last_message_at = EXCLUDED.last_message_at;

-- Messages
INSERT INTO messages (
  message_id,
  conversation_id,
  sender_id,
  content,
  encrypted,
  has_images,
  is_edited,
  created_at,
  updated_at
) VALUES
  (
    '50000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000002',
    'Hey Alice! Ready for our React-for-Go swap this weekend?',
    false,
    false,
    false,
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days'
  ),
  (
    '50000000-0000-0000-0000-000000000002',
    '40000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'Absolutely. I can do Saturday at 10 AM. Let us start with hooks + state patterns.',
    false,
    false,
    false,
    NOW() - INTERVAL '90 minutes',
    NOW() - INTERVAL '90 minutes'
  )
ON CONFLICT (message_id) DO NOTHING;

-- Read status
INSERT INTO message_read_status (user_id, conversation_id, last_read_at) VALUES
  ('00000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000001', NOW() - INTERVAL '85 minutes'),
  ('00000000-0000-0000-0000-000000000002', '40000000-0000-0000-0000-000000000001', NOW() - INTERVAL '4 hours')
ON CONFLICT (user_id, conversation_id) DO UPDATE SET
  last_read_at = EXCLUDED.last_read_at;

-- Notifications
INSERT INTO notifications (
  notification_id,
  user_id,
  type,
  title,
  message,
  is_read,
  related_id,
  created_at
) VALUES
  (
    '60000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'swap_request',
    'New Swap Request',
    'Bob Builder requested a skill swap for React <-> Go.',
    false,
    '30000000-0000-0000-0000-000000000001',
    NOW() - INTERVAL '6 days'
  ),
  (
    '60000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000002',
    'swap_accepted',
    'Swap Request Accepted',
    'Alice Mentor accepted your swap request.',
    false,
    '30000000-0000-0000-0000-000000000001',
    NOW() - INTERVAL '5 days'
  )
ON CONFLICT (notification_id) DO NOTHING;

COMMIT;
