package database

import (
	"log"

	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Initialize establishes database connection and returns GORM DB instance
func Initialize(dbURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// Test the connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	log.Println("Database connected successfully")

	// Create PostgreSQL extensions and types if they don't exist
	if err := createPostgreSQLExtensions(db); err != nil {
		log.Printf("Warning: Could not create PostgreSQL extensions: %v", err)
	}

	return db, nil
}

// createPostgreSQLExtensions creates necessary PostgreSQL extensions and types
func createPostgreSQLExtensions(db *gorm.DB) error {
	// Enable UUID extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return err
	}

	// Create swap_status enum type
	if err := db.Exec("CREATE TYPE swap_status AS ENUM ('pending', 'accepted', 'rejected', 'cancelled')").Error; err != nil {
		// Ignore error if type already exists
		log.Printf("Info: swap_status type might already exist: %v", err)
	}

	log.Println("✓ PostgreSQL extensions and types created")
	return nil
}

// Migrate runs database migrations
func Migrate(db *gorm.DB) error {
	log.Println("Starting database migrations...")

	// Check if core tables exist first
	if hasAllTables(db) {
		log.Println("✓ All core tables already exist")

		// Run additional migrations
		if err := runAdditionalMigrations(db); err != nil {
			log.Printf("Warning: Could not run additional migrations: %v", err)
		}

		// Still try to seed default skills
		if err := seedDefaultSkills(db); err != nil {
			log.Printf("Warning: Could not seed default skills: %v", err)
		}

		log.Println("✅ Database schema is ready")
		return nil
	}

	// Execute SQL migration script
	if err := runSQLMigration(db); err != nil {
		log.Printf("Error running SQL migration: %v", err)
		return err
	}

	// Insert default skills if they don't exist
	if err := seedDefaultSkills(db); err != nil {
		log.Printf("Warning: Could not seed default skills: %v", err)
	}

	// Run additional migrations
	if err := runAdditionalMigrations(db); err != nil {
		log.Printf("Warning: Could not run additional migrations: %v", err)
	}

	log.Println("✅ Database migrations completed successfully")
	return nil
}

// runSQLMigration executes the SQL migration script
func runSQLMigration(db *gorm.DB) error {
	log.Println("Running SQL migration script...")

	// Read and execute the migration file
	sqlContent := `
-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create custom ENUM types
DO $$ BEGIN
    CREATE TYPE swap_status AS ENUM ('pending', 'accepted', 'rejected', 'cancelled');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    location TEXT,
    is_public BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Skills table
CREATE TABLE IF NOT EXISTS skills (
    skill_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Availability slots table
CREATE TABLE IF NOT EXISTS availability_slots (
    slot_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    label TEXT NOT NULL,
    day_bitmask INTEGER NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_availability_slots_user_id 
        FOREIGN KEY (user_id) REFERENCES users(user_id) 
        ON UPDATE CASCADE ON DELETE CASCADE
);

-- User skills offered junction table
CREATE TABLE IF NOT EXISTS user_skills_offered (
    user_id UUID NOT NULL,
    skill_id UUID NOT NULL,
    
    PRIMARY KEY (user_id, skill_id),
    
    CONSTRAINT fk_user_skills_offered_user_id 
        FOREIGN KEY (user_id) REFERENCES users(user_id) 
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_user_skills_offered_skill_id 
        FOREIGN KEY (skill_id) REFERENCES skills(skill_id) 
        ON UPDATE CASCADE ON DELETE CASCADE
);

-- User skills wanted junction table
CREATE TABLE IF NOT EXISTS user_skills_wanted (
    user_id UUID NOT NULL,
    skill_id UUID NOT NULL,
    
    PRIMARY KEY (user_id, skill_id),
    
    CONSTRAINT fk_user_skills_wanted_user_id 
        FOREIGN KEY (user_id) REFERENCES users(user_id) 
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_user_skills_wanted_skill_id 
        FOREIGN KEY (skill_id) REFERENCES skills(skill_id) 
        ON UPDATE CASCADE ON DELETE CASCADE
);

-- Swap requests table
CREATE TABLE IF NOT EXISTS swap_requests (
    swap_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    requester_id UUID NOT NULL,
    responder_id UUID NOT NULL,
    offered_skill_id UUID NOT NULL,
    wanted_skill_id UUID NOT NULL,
    status swap_status DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT fk_swap_requests_requester_id 
        FOREIGN KEY (requester_id) REFERENCES users(user_id) 
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_swap_requests_responder_id 
        FOREIGN KEY (responder_id) REFERENCES users(user_id) 
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_swap_requests_offered_skill_id 
        FOREIGN KEY (offered_skill_id) REFERENCES skills(skill_id) 
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_swap_requests_wanted_skill_id 
        FOREIGN KEY (wanted_skill_id) REFERENCES skills(skill_id) 
        ON UPDATE CASCADE ON DELETE RESTRICT
);

-- Swap ratings table
CREATE TABLE IF NOT EXISTS swap_ratings (
    rating_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    swap_id UUID NOT NULL,
    rater_id UUID NOT NULL,
    ratee_id UUID NOT NULL,
    score SMALLINT NOT NULL CHECK (score >= 1 AND score <= 5),
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_swap_ratings_swap_id 
        FOREIGN KEY (swap_id) REFERENCES swap_requests(swap_id) 
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_swap_ratings_rater_id 
        FOREIGN KEY (rater_id) REFERENCES users(user_id) 
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_swap_ratings_ratee_id 
        FOREIGN KEY (ratee_id) REFERENCES users(user_id) 
        ON UPDATE CASCADE ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_skills_name ON skills(name);
CREATE INDEX IF NOT EXISTS idx_availability_slots_user_id ON availability_slots(user_id);
CREATE INDEX IF NOT EXISTS idx_user_skills_offered_user_id ON user_skills_offered(user_id);
CREATE INDEX IF NOT EXISTS idx_user_skills_offered_skill_id ON user_skills_offered(skill_id);
CREATE INDEX IF NOT EXISTS idx_user_skills_wanted_user_id ON user_skills_wanted(user_id);
CREATE INDEX IF NOT EXISTS idx_user_skills_wanted_skill_id ON user_skills_wanted(skill_id);
CREATE INDEX IF NOT EXISTS idx_swap_requests_requester_id ON swap_requests(requester_id);
CREATE INDEX IF NOT EXISTS idx_swap_requests_responder_id ON swap_requests(responder_id);
CREATE INDEX IF NOT EXISTS idx_swap_requests_offered_skill_id ON swap_requests(offered_skill_id);
CREATE INDEX IF NOT EXISTS idx_swap_requests_wanted_skill_id ON swap_requests(wanted_skill_id);
CREATE INDEX IF NOT EXISTS idx_swap_requests_status ON swap_requests(status);
CREATE INDEX IF NOT EXISTS idx_swap_requests_deleted_at ON swap_requests(deleted_at);
CREATE INDEX IF NOT EXISTS idx_swap_ratings_swap_id ON swap_ratings(swap_id);
CREATE INDEX IF NOT EXISTS idx_swap_ratings_rater_id ON swap_ratings(rater_id);
CREATE INDEX IF NOT EXISTS idx_swap_ratings_ratee_id ON swap_ratings(ratee_id);
CREATE INDEX IF NOT EXISTS idx_swap_ratings_score ON swap_ratings(score);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at columns
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_swap_requests_updated_at ON swap_requests;
CREATE TRIGGER update_swap_requests_updated_at 
    BEFORE UPDATE ON swap_requests 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
`

	if err := db.Exec(sqlContent).Error; err != nil {
		return err
	}

	log.Println("✓ SQL migration completed successfully")
	return nil
}

// runAdditionalMigrations runs additional migration scripts
func runAdditionalMigrations(db *gorm.DB) error {
	log.Println("Running additional migrations...")

	// Check if admin fields already exist
	var hasAdminField bool
	err := db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='is_admin')").Scan(&hasAdminField).Error
	if err != nil {
		return err
	}

	if !hasAdminField {
		log.Println("Adding admin and ban fields to users table...")

		// Add admin and ban fields
		sql := `
			ALTER TABLE users 
			ADD COLUMN IF NOT EXISTS is_admin BOOLEAN DEFAULT FALSE,
			ADD COLUMN IF NOT EXISTS is_banned BOOLEAN DEFAULT FALSE;

			-- Create indexes for admin lookups
			CREATE INDEX IF NOT EXISTS idx_users_is_admin ON users(is_admin);
			CREATE INDEX IF NOT EXISTS idx_users_is_banned ON users(is_banned);
		`

		if err := db.Exec(sql).Error; err != nil {
			return err
		}

		log.Println("✓ Added admin and ban fields to users table")
	} else {
		log.Println("✓ Admin fields already exist")
	}

	// Check if notifications table exists
	var hasNotificationsTable bool
	err = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='notifications')").Scan(&hasNotificationsTable).Error
	if err != nil {
		return err
	}

	if !hasNotificationsTable {
		log.Println("Creating notifications table...")

		// Create notifications table
		sql := `
			CREATE TABLE IF NOT EXISTS notifications (
				notification_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				user_id UUID NOT NULL,
				type VARCHAR(50) NOT NULL,
				title VARCHAR(255) NOT NULL,
				message TEXT NOT NULL,
				is_read BOOLEAN DEFAULT FALSE,
				related_id UUID,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				deleted_at TIMESTAMP WITH TIME ZONE,
				
				FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
			);

			-- Create indexes for better performance
			CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
			CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);
			CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
			CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
			CREATE INDEX IF NOT EXISTS idx_notifications_deleted_at ON notifications(deleted_at);

			-- Create composite index for common queries
			CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications(user_id, is_read) WHERE deleted_at IS NULL;
		`

		if err := db.Exec(sql).Error; err != nil {
			return err
		}

		log.Println("✓ Created notifications table")
	} else {
		log.Println("✓ Notifications table already exists")
	}

	// Check if photo storage fields exist
	var hasPhotoData bool
	err = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='photo_data')").Scan(&hasPhotoData).Error
	if err != nil {
		return err
	}

	if !hasPhotoData {
		log.Println("Adding photo storage fields to users table...")

		// Add photo storage fields
		sql := `
			ALTER TABLE users 
			ADD COLUMN IF NOT EXISTS photo_data BYTEA,
			ADD COLUMN IF NOT EXISTS photo_mime_type VARCHAR(100);
		`

		if err := db.Exec(sql).Error; err != nil {
			return err
		}

		log.Println("✓ Added photo storage fields to users table")
	} else {
		log.Println("✓ Photo storage fields already exist")
	}

	// Check if conversations table exists (Migration 006)
	var hasConversationsTable bool
	err = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='conversations')").Scan(&hasConversationsTable).Error
	if err != nil {
		return err
	}

	if !hasConversationsTable {
		log.Println("Creating chat tables...")

		sql := `
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

			-- Message read status
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

			-- Chat images table
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
		`

		if err := db.Exec(sql).Error; err != nil {
			return err
		}

		log.Println("✓ Created chat tables")
	} else {
		log.Println("✓ Chat tables already exist")
	}

	// Check if swap completion columns exist (Migration 007)
	var hasRequesterCompleted bool
	err = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='swap_requests' AND column_name='requester_completed')").Scan(&hasRequesterCompleted).Error
	if err != nil {
		return err
	}

	if !hasRequesterCompleted {
		log.Println("Adding swap completion fields...")

		sql := `
			ALTER TABLE swap_requests
				ADD COLUMN IF NOT EXISTS requester_completed BOOLEAN DEFAULT FALSE,
				ADD COLUMN IF NOT EXISTS responder_completed BOOLEAN DEFAULT FALSE;
		`

		if err := db.Exec(sql).Error; err != nil {
			return err
		}

		// Add 'completed' to swap_status enum
		if err := db.Exec("DO $$ BEGIN ALTER TYPE swap_status ADD VALUE IF NOT EXISTS 'completed'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;").Error; err != nil {
			log.Printf("Info: swap_status 'completed' value might already exist: %v", err)
		}

		log.Println("✓ Added swap completion fields")
	} else {
		log.Println("✓ Swap completion fields already exist")
	}

	// Check if E2EE key fields exist (Migration 008)
	var hasPublicKey bool
	err = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='public_key')").Scan(&hasPublicKey).Error
	if err != nil {
		return err
	}

	if !hasPublicKey {
		log.Println("Adding E2EE key fields to users table...")

		sql := `
			ALTER TABLE users
			ADD COLUMN IF NOT EXISTS public_key TEXT,
			ADD COLUMN IF NOT EXISTS encrypted_key_backup TEXT;

			CREATE INDEX IF NOT EXISTS idx_users_has_public_key
				ON users (user_id) WHERE public_key IS NOT NULL;
		`

		if err := db.Exec(sql).Error; err != nil {
			return err
		}

		log.Println("✓ Added E2EE key fields to users table")
	} else {
		log.Println("✓ E2EE key fields already exist")
	}

	// Check if encrypted column exists on messages table (Migration 009)
	var hasEncrypted bool
	err = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='messages' AND column_name='encrypted')").Scan(&hasEncrypted).Error
	if err != nil {
		return err
	}

	if !hasEncrypted {
		log.Println("Adding encrypted field to messages table...")

		sql := `ALTER TABLE messages ADD COLUMN IF NOT EXISTS encrypted BOOLEAN NOT NULL DEFAULT false;`

		if err := db.Exec(sql).Error; err != nil {
			return err
		}

		log.Println("✓ Added encrypted field to messages table")
	} else {
		log.Println("✓ Messages encrypted field already exists")
	}

	// Check if OAuth/email columns exist on users table (Migration 010)
	var hasAuthProvider bool
	err = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='auth_provider')").Scan(&hasAuthProvider).Error
	if err != nil {
		return err
	}

	if !hasAuthProvider {
		log.Println("Adding OAuth/email fields to users table...")

		sql := `
			ALTER TABLE users
			ADD COLUMN IF NOT EXISTS auth_provider VARCHAR(20) NOT NULL DEFAULT 'local',
			ADD COLUMN IF NOT EXISTS google_id VARCHAR(255) UNIQUE,
			ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE;

			ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

			CREATE INDEX IF NOT EXISTS idx_users_auth_provider ON users(auth_provider);
			CREATE UNIQUE INDEX IF NOT EXISTS idx_users_google_id_unique ON users(google_id) WHERE google_id IS NOT NULL;
		`

		if err := db.Exec(sql).Error; err != nil {
			return err
		}

		log.Println("✓ Added OAuth/email fields to users table")
	} else {
		log.Println("✓ OAuth/email fields already exist")
	}

	// Ensure password reset tokens table exists
	var hasPasswordResetTable bool
	err = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='password_reset_tokens')").Scan(&hasPasswordResetTable).Error
	if err != nil {
		return err
	}

	if !hasPasswordResetTable {
		log.Println("Creating password reset tokens table...")

		sql := `
			CREATE TABLE IF NOT EXISTS password_reset_tokens (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
				token VARCHAR(255) NOT NULL UNIQUE,
				expires_at TIMESTAMP NOT NULL,
				used BOOLEAN NOT NULL DEFAULT FALSE,
				created_at TIMESTAMP DEFAULT NOW()
			);

			CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token ON password_reset_tokens(token);
			CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
		`

		if err := db.Exec(sql).Error; err != nil {
			return err
		}

		log.Println("✓ Created password reset tokens table")
	} else {
		log.Println("✓ Password reset tokens table already exists")
	}

	// Ensure email verification tokens table exists
	var hasEmailVerificationTable bool
	err = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='email_verification_tokens')").Scan(&hasEmailVerificationTable).Error
	if err != nil {
		return err
	}

	if !hasEmailVerificationTable {
		log.Println("Creating email verification tokens table...")

		sql := `
			CREATE TABLE IF NOT EXISTS email_verification_tokens (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
				token VARCHAR(255) NOT NULL UNIQUE,
				expires_at TIMESTAMP NOT NULL,
				created_at TIMESTAMP DEFAULT NOW()
			);

			CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_token ON email_verification_tokens(token);
			CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_user_id ON email_verification_tokens(user_id);
		`

		if err := db.Exec(sql).Error; err != nil {
			return err
		}

		log.Println("✓ Created email verification tokens table")
	} else {
		log.Println("✓ Email verification tokens table already exists")
	}

	// Migration 011: user_blocks + reports (Commit 2)
	if !db.Migrator().HasTable("user_blocks") {
		log.Println("Creating user_blocks table...")
		sql := `
			CREATE TABLE IF NOT EXISTS user_blocks (
				blocker_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
				blocked_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (blocker_id, blocked_id),
				CHECK (blocker_id <> blocked_id)
			);
			CREATE INDEX IF NOT EXISTS idx_user_blocks_blocked_id ON user_blocks(blocked_id);
		`
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
		log.Println("✓ Created user_blocks table")
	} else {
		log.Println("✓ user_blocks table already exists")
	}

	if !db.Migrator().HasTable("reports") {
		log.Println("Creating reports table...")
		sql := `
			CREATE TABLE IF NOT EXISTS reports (
				report_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				reporter_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
				target_user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
				target_kind VARCHAR(20) NOT NULL,
				target_id UUID,
				reason TEXT NOT NULL,
				status VARCHAR(20) NOT NULL DEFAULT 'open',
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				resolved_at TIMESTAMP WITH TIME ZONE,
				resolver_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
				resolution_note TEXT
			);
			CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status);
			CREATE INDEX IF NOT EXISTS idx_reports_target_user ON reports(target_user_id);
			CREATE INDEX IF NOT EXISTS idx_reports_reporter ON reports(reporter_id);
		`
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
		log.Println("✓ Created reports table")
	} else {
		log.Println("✓ reports table already exists")
	}

	// Migration 012: no-show / dispute columns on swap_requests (Commit 3)
	for col, ddl := range map[string]string{
		"no_show_flag":        "ALTER TABLE swap_requests ADD COLUMN no_show_flag BOOLEAN NOT NULL DEFAULT FALSE",
		"no_show_reporter_id": "ALTER TABLE swap_requests ADD COLUMN no_show_reporter_id UUID REFERENCES users(user_id) ON DELETE SET NULL",
		"no_show_reason":      "ALTER TABLE swap_requests ADD COLUMN no_show_reason TEXT",
		"dispute_reason":      "ALTER TABLE swap_requests ADD COLUMN dispute_reason TEXT",
	} {
		if !db.Migrator().HasColumn(&models.SwapRequest{}, col) {
			if err := db.Exec(ddl).Error; err != nil {
				return err
			}
			log.Printf("✓ Added swap_requests.%s", col)
		}
	}

	// Migration 013: users.slug (Commit 4)
	if !db.Migrator().HasColumn(&models.User{}, "slug") {
		if err := db.Exec("ALTER TABLE users ADD COLUMN slug VARCHAR(40)").Error; err != nil {
			return err
		}
		if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_users_slug ON users(slug) WHERE slug IS NOT NULL").Error; err != nil {
			return err
		}
		log.Println("✓ Added users.slug column + unique index")
	}

	return nil
}

// hasAllTables checks if all required tables exist
func hasAllTables(db *gorm.DB) bool {
	tables := []interface{}{
		&models.User{},
		&models.Skill{},
		&models.AvailabilitySlot{},
		&models.UserSkillOffered{},
		&models.UserSkillWanted{},
		&models.SwapRequest{},
		&models.SwapRating{},
		&models.Notification{},
		&models.Conversation{},
		&models.Message{},
		&models.MessageReadStatus{},
		&models.ChatImage{},
	}

	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			return false
		}
	}
	return true
}

// seedDefaultSkills inserts default skills into the database if they don't exist
func seedDefaultSkills(db *gorm.DB) error {
	defaultSkills := []string{
		"JavaScript",
		"Python",
		"React",
		"Node.js",
		"PostgreSQL",
		"Docker",
		"AWS",
		"Machine Learning",
		"UI/UX Design",
		"Data Analysis",
		"Go",
		"TypeScript",
		"GraphQL",
		"Kubernetes",
		"DevOps",
		"HTML/CSS",
		"Java",
		"C++",
		"Swift",
		"Flutter",
		"Angular",
		"Vue.js",
		"MongoDB",
		"Redis",
		"Microservices",
		"System Design",
		"Cybersecurity",
		"Mobile Development",
		"Web Development",
		"Backend Development",
	}

	var existingSkills []models.Skill
	if err := db.Find(&existingSkills).Error; err != nil {
		return err
	}

	existingSkillsMap := make(map[string]bool)
	for _, skill := range existingSkills {
		existingSkillsMap[skill.Name] = true
	}

	var skillsToInsert []models.Skill
	for _, skillName := range defaultSkills {
		if !existingSkillsMap[skillName] {
			skillsToInsert = append(skillsToInsert, models.Skill{
				Name: skillName,
			})
		}
	}

	if len(skillsToInsert) > 0 {
		if err := db.Create(&skillsToInsert).Error; err != nil {
			return err
		}
		log.Printf("✓ Inserted %d default skills", len(skillsToInsert))
	} else {
		log.Println("✓ Default skills already exist")
	}

	return nil
}
