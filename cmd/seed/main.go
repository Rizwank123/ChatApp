package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/Rizwank123/chatApp/internal/dependency"
	"github.com/Rizwank123/chatApp/internal/pkg/config"
	"github.com/Rizwank123/chatApp/internal/pkg/util"
)

type SeedUser struct {
	UserName    string
	DisplayName string
	Email       string
	FirstName   string
	LastName    string
	Mobile      string
	Avatar      string
	Bio         string
}

type SeedPost struct {
	AuthorUsername string
	Caption        string
	MediaURLs      []string
	Location       string
}

func main() {
	cfgOpt := config.Options{
		ConfigFileSource: config.SourceEnv,
		ConfigFile:       ".env",
	}
	cfg, err := dependency.NewConfig(cfgOpt)
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	db, err := dependency.NewDatabaseConfig(cfg)
	if err != nil {
		log.Fatalf("failed to create database pool: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("🌱 Seeding deterministic test data for ChatApp...")

	seedUsers := []SeedUser{
		{
			UserName:    "alice",
			DisplayName: "Alice Vance",
			Email:       "alice@example.com",
			FirstName:   "Alice",
			LastName:    "Vance",
			Mobile:      "+15551001001",
			Avatar:      "https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=200&fit=crop",
			Bio:         "Cybersecurity researcher & open-source enthusiast 🔐💻",
		},
		{
			UserName:    "bob",
			DisplayName: "Bob Smith",
			Email:       "bob@example.com",
			FirstName:   "Bob",
			LastName:    "Smith",
			Mobile:      "+15551001002",
			Avatar:      "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=200&fit=crop",
			Bio:         "Fullstack engineer building distributed Go backends ⚡☕",
		},
		{
			UserName:    "charlie",
			DisplayName: "Charlie Rose",
			Email:       "charlie@example.com",
			FirstName:   "Charlie",
			LastName:    "Rose",
			Mobile:      "+15551001003",
			Avatar:      "https://images.unsplash.com/photo-1570295999919-56ceb5ecca61?w=200&fit=crop",
			Bio:         "Designer crafting futuristic neon UI interfaces 🎨✨",
		},
		{
			UserName:    "david",
			DisplayName: "David Miller",
			Email:       "david@example.com",
			FirstName:   "David",
			LastName:    "Miller",
			Mobile:      "+15551001004",
			Avatar:      "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=200&fit=crop",
			Bio:         "Product strategist & tech explorer 🚀",
		},
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	userIDs := make(map[string]uuid.UUID)
	personnelIDs := make(map[string]uuid.UUID)

	for _, su := range seedUsers {
		normUsername := util.NormalizeUsername(su.UserName)

		var userID uuid.UUID
		err := db.QueryRow(ctx, "SELECT id FROM users WHERE LOWER(TRIM(user_name)) = $1", normUsername).Scan(&userID)
		if err != nil {
			userID = uuid.Must(uuid.NewV4())
			userQ := `INSERT INTO users (id, user_name, password, role, created_at, updated_at)
			          VALUES ($1, $2, $3, $4, NOW(), NOW())`
			_, err = db.Exec(ctx, userQ, userID, normUsername, string(hashedPassword), "USER")
			if err != nil {
				log.Fatalf("failed to insert user %s: %v", normUsername, err)
			}
			fmt.Printf("  ✓ Created user: @%s (ID: %s)\n", normUsername, userID)
		} else {
			// Update password to ensure it matches Password123!
			_, _ = db.Exec(ctx, "UPDATE users SET password = $1 WHERE id = $2", string(hashedPassword), userID)
			fmt.Printf("  • Found existing user: @%s (ID: %s)\n", normUsername, userID)
		}
		userIDs[normUsername] = userID

		var personnelID uuid.UUID
		err = db.QueryRow(ctx, "SELECT id FROM personnel WHERE user_id = $1", userID).Scan(&personnelID)
		if err != nil {
			personnelID = uuid.Must(uuid.NewV4())
			personnelQ := `INSERT INTO personnel (id, user_id, first_name, last_name, display_name, email, mobile, role, activation_status, avatar, bio, created_at, updated_at)
			               VALUES ($1, $2, $3, $4, $5, $6, $7, 'USER', 'ACTIVE', $8, $9, NOW(), NOW())`
			_, err = db.Exec(ctx, personnelQ, personnelID, userID, su.FirstName, su.LastName, su.DisplayName, su.Email, su.Mobile, su.Avatar, su.Bio)
			if err != nil {
				log.Fatalf("failed to insert personnel for %s: %v", normUsername, err)
			}
			fmt.Printf("  ✓ Created personnel profile for @%s (Personnel ID: %s)\n", normUsername, personnelID)
		} else {
			_, _ = db.Exec(ctx, `UPDATE personnel 
			                     SET display_name = $1, bio = $2, avatar = $3, mobile = $4, email = $5 
			                     WHERE id = $6`,
				su.DisplayName, su.Bio, su.Avatar, su.Mobile, su.Email, personnelID)
			fmt.Printf("  • Updated personnel profile for @%s (Personnel ID: %s)\n", normUsername, personnelID)
		}
		personnelIDs[normUsername] = personnelID

		// Ensure user_counters row exists
		_, _ = db.Exec(ctx, `INSERT INTO user_counters (personnel_id, follower_count, following_count, post_count, updated_at)
		                     VALUES ($1, 0, 0, 0, NOW())
		                     ON CONFLICT (personnel_id) DO NOTHING`, personnelID)
	}

	aliceID := personnelIDs["alice"]
	bobID := personnelIDs["bob"]
	charlieID := personnelIDs["charlie"]
	davidID := personnelIDs["david"]

	// Seed Follow Relationships:
	// 1. Alice <-> Bob : Mutual follow (Can chat ✅)
	// 2. Alice -> Charlie : 1-way follow (Cannot chat ⛔)
	// 3. Bob <-> Charlie : Mutual follow (Can chat ✅)
	// 4. David -> Alice : 1-way follow
	followPairs := [][2]uuid.UUID{
		{aliceID, bobID},     // Alice follows Bob
		{bobID, aliceID},     // Bob follows Alice
		{aliceID, charlieID}, // Alice follows Charlie
		{bobID, charlieID},   // Bob follows Charlie
		{charlieID, bobID},   // Charlie follows Bob
		{davidID, aliceID},   // David follows Alice
	}

	for _, pair := range followPairs {
		_, err := db.Exec(ctx, `INSERT INTO follows (follower_id, following_id, created_at, updated_at, deleted_at)
		                        VALUES ($1, $2, NOW(), NOW(), NULL)
		                        ON CONFLICT (follower_id, following_id) DO UPDATE SET deleted_at = NULL`, pair[0], pair[1])
		if err != nil {
			log.Printf("failed to seed follow relation: %v", err)
		}
	}

	// Seed sample Posts
	seedPosts := []SeedPost{
		{
			AuthorUsername: "alice",
			Caption:        "Neon nightfall in Tokyo 🌆✨ Loving the futuristic vibe here! #tokyo #vibely #cyberpunk",
			MediaURLs:      []string{"https://images.unsplash.com/photo-1503899036084-c55cdd92da26?w=800&fit=crop"},
			Location:       "Shinjuku, Tokyo",
		},
		{
			AuthorUsername: "alice",
			Caption:        "Zero-trust security and end-to-end encrypted messaging deep dive 🔐⚡ #security #privacy #golang",
			MediaURLs:      []string{"https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?w=800&fit=crop"},
			Location:       "San Francisco, CA",
		},
		{
			AuthorUsername: "bob",
			Caption:        "Late night distributed systems debugging with Go & Kafka ☕🚀 #backend #distributed #golang",
			MediaURLs:      []string{"https://images.unsplash.com/photo-1555066931-4365d14bab8c?w=800&fit=crop"},
			Location:       "Seattle, WA",
		},
		{
			AuthorUsername: "bob",
			Caption:        "Weekend mountain hike with crisp morning air 🏔️☀️ Nature resets the mind.",
			MediaURLs:      []string{"https://images.unsplash.com/photo-1464822759023-fed622ff2c3b?w=800&fit=crop"},
			Location:       "Mount Rainier, WA",
		},
		{
			AuthorUsername: "charlie",
			Caption:        "Neon glassmorphic UI system designed for Vibely! Sleek, modern, and dark-mode first 🎨✨ #uiux #design #vibely",
			MediaURLs:      []string{"https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=800&fit=crop"},
			Location:       "Berlin, Germany",
		},
	}

	postIDs := make([]uuid.UUID, 0)
	for _, sp := range seedPosts {
		authorID := personnelIDs[sp.AuthorUsername]
		var existingPostID uuid.UUID
		err := db.QueryRow(ctx, "SELECT id FROM posts WHERE personnel_id = $1 AND caption = $2 LIMIT 1", authorID, sp.Caption).Scan(&existingPostID)
		if err != nil {
			newPostID := uuid.Must(uuid.NewV4())
			postQ := `INSERT INTO posts (id, personnel_id, caption, media_url, location, visibility, status, is_archived, created_at, updated_at)
			          VALUES ($1, $2, $3, $4, $5, 'PUBLIC', 'PUBLISHED', FALSE, NOW() - INTERVAL '2 hours', NOW())`
			_, err = db.Exec(ctx, postQ, newPostID, authorID, sp.Caption, sp.MediaURLs, sp.Location)
			if err != nil {
				log.Printf("failed to insert post for %s: %v", sp.AuthorUsername, err)
			} else {
				postIDs = append(postIDs, newPostID)
				fmt.Printf("  ✓ Seeded post for @%s: \"%s\"\n", sp.AuthorUsername, sp.Caption[:30]+"...")
			}
		} else {
			postIDs = append(postIDs, existingPostID)
		}
	}

	// Seed Likes and Comments on posts
	if len(postIDs) > 0 {
		for i, pID := range postIDs {
			// Bob likes Alice's posts, Alice likes Bob's posts, Charlie likes all
			likers := []uuid.UUID{charlieID}
			if i%2 == 0 {
				likers = append(likers, bobID)
			} else {
				likers = append(likers, aliceID)
			}

			for _, liker := range likers {
				_, _ = db.Exec(ctx, `INSERT INTO likes (personnel_id, post_id, created_at, updated_at)
				                     VALUES ($1, $2, NOW(), NOW())
				                     ON CONFLICT (personnel_id, post_id) DO UPDATE SET deleted_at = NULL`, liker, pID)
			}

			// Seed comments
			commentID := uuid.Must(uuid.NewV4())
			commenter := bobID
			text := "Looks absolutely incredible! 🔥"
			if i%2 == 1 {
				commenter = aliceID
				text = "Super clean architecture! 🚀"
			}
			_, _ = db.Exec(ctx, `INSERT INTO comments (id, personnel_id, post_id, content, created_at, updated_at)
			                     VALUES ($1, $2, $3, $4, NOW(), NOW())
			                     ON CONFLICT DO NOTHING`, commentID, commenter, pID, text)
		}
	}

	// Update counter tallies accurately
	updateCountersQ := `
		UPDATE user_counters uc
		SET
			follower_count = (SELECT COUNT(*) FROM follows WHERE following_id = uc.personnel_id AND deleted_at IS NULL),
			following_count = (SELECT COUNT(*) FROM follows WHERE follower_id = uc.personnel_id AND deleted_at IS NULL),
			post_count = (SELECT COUNT(*) FROM posts WHERE personnel_id = uc.personnel_id AND deleted_at IS NULL),
			updated_at = NOW()
	`
	_, _ = db.Exec(ctx, updateCountersQ)

	// Seed Conversations & Messages:
	// Conversation 1: Alice & Bob (Mutual Follow)
	var conv1ID uuid.UUID
	checkConv1Q := `
		SELECT c.id FROM conversations c
		JOIN conversation_participants cp1 ON c.id = cp1.conversation_id AND cp1.personnel_id = $1
		JOIN conversation_participants cp2 ON c.id = cp2.conversation_id AND cp2.personnel_id = $2
		WHERE c.conversation_type = 'DIRECT'
		LIMIT 1`
	err = db.QueryRow(ctx, checkConv1Q, aliceID, bobID).Scan(&conv1ID)
	if err != nil {
		conv1ID = uuid.Must(uuid.NewV4())
		chatName := "Alice & Bob"
		_, err = db.Exec(ctx, "INSERT INTO conversations (id, name, conversation_type, created_at, updated_at) VALUES ($1, $2, 'DIRECT', NOW(), NOW())", conv1ID, chatName)
		if err != nil {
			log.Printf("failed to insert conversation between Alice and Bob: %v", err)
		}
	} else {
		fmt.Printf("  • Existing Alice & Bob conversation found (ID: %s)\n", conv1ID)
	}

	aliceUserID := userIDs["alice"]
	bobUserID := userIDs["bob"]
	_, _ = db.Exec(ctx, "INSERT INTO conversation_participants (conversation_id, personnel_id, user_id, joined_at) VALUES ($1, $2, $3, NOW()) ON CONFLICT DO NOTHING", conv1ID, aliceID, aliceUserID)
	_, _ = db.Exec(ctx, "INSERT INTO conversation_participants (conversation_id, personnel_id, user_id, joined_at) VALUES ($1, $2, $3, NOW()) ON CONFLICT DO NOTHING", conv1ID, bobID, bobUserID)

	var msgCount1 int
	_ = db.QueryRow(ctx, "SELECT COUNT(*) FROM messages WHERE conversation_id = $1", conv1ID).Scan(&msgCount1)
	if msgCount1 == 0 {
		msg1 := uuid.Must(uuid.NewV4())
		if _, err := db.Exec(ctx, `INSERT INTO messages (id, conversation_id, sender_id, message_type, content, created_at, updated_at)
		                     VALUES ($1, $2, $3, 'TEXT', 'Hey Bob! Excited for the launch party tonight 🚀', NOW() - INTERVAL '15 minutes', NOW())`, msg1, conv1ID, aliceID); err != nil {
			log.Printf("failed to insert msg1: %v", err)
		}

		msg2 := uuid.Must(uuid.NewV4())
		if _, err := db.Exec(ctx, `INSERT INTO messages (id, conversation_id, sender_id, message_type, content, created_at, updated_at)
		                     VALUES ($1, $2, $3, 'TEXT', 'Hey Alice! Everything is ready. Realtime WebSocket chat is flying! ⚡', NOW() - INTERVAL '10 minutes', NOW())`, msg2, conv1ID, bobID); err != nil {
			log.Printf("failed to insert msg2: %v", err)
		}

		msg3 := uuid.Must(uuid.NewV4())
		if _, err := db.Exec(ctx, `INSERT INTO messages (id, conversation_id, sender_id, message_type, content, created_at, updated_at)
		                     VALUES ($1, $2, $3, 'TEXT', 'Just checked the feeds too, the UI looks so smooth.', NOW() - INTERVAL '5 minutes', NOW())`, msg3, conv1ID, aliceID); err != nil {
			log.Printf("failed to insert msg3: %v", err)
		}
		fmt.Printf("  ✓ Seeded mutual-follow chat between Alice and Bob (Conv ID: %s)\n", conv1ID)
	}

	// Conversation 2: Bob & Charlie (Mutual Follow)
	var conv2ID uuid.UUID
	err = db.QueryRow(ctx, checkConv1Q, bobID, charlieID).Scan(&conv2ID)
	if err != nil {
		conv2ID = uuid.Must(uuid.NewV4())
		chatName := "Bob & Charlie"
		_, err = db.Exec(ctx, "INSERT INTO conversations (id, name, conversation_type, created_at, updated_at) VALUES ($1, $2, 'DIRECT', NOW(), NOW())", conv2ID, chatName)
		if err != nil {
			log.Printf("failed to insert conversation between Bob and Charlie: %v", err)
		}
	} else {
		fmt.Printf("  • Existing Bob & Charlie conversation found (ID: %s)\n", conv2ID)
	}

	charlieUserID := userIDs["charlie"]
	_, _ = db.Exec(ctx, "INSERT INTO conversation_participants (conversation_id, personnel_id, user_id, joined_at) VALUES ($1, $2, $3, NOW()) ON CONFLICT DO NOTHING", conv2ID, bobID, bobUserID)
	_, _ = db.Exec(ctx, "INSERT INTO conversation_participants (conversation_id, personnel_id, user_id, joined_at) VALUES ($1, $2, $3, NOW()) ON CONFLICT DO NOTHING", conv2ID, charlieID, charlieUserID)

	var msgCount2 int
	_ = db.QueryRow(ctx, "SELECT COUNT(*) FROM messages WHERE conversation_id = $1", conv2ID).Scan(&msgCount2)
	if msgCount2 == 0 {
		msg1 := uuid.Must(uuid.NewV4())
		if _, err := db.Exec(ctx, `INSERT INTO messages (id, conversation_id, sender_id, message_type, content, created_at, updated_at)
		                     VALUES ($1, $2, $3, 'TEXT', 'Hey Charlie, love the new neon cyber theme! 🎨', NOW() - INTERVAL '30 minutes', NOW())`, msg1, conv2ID, bobID); err != nil {
			log.Printf("failed to insert bob-charlie msg1: %v", err)
		}

		msg2 := uuid.Must(uuid.NewV4())
		if _, err := db.Exec(ctx, `INSERT INTO messages (id, conversation_id, sender_id, message_type, content, created_at, updated_at)
		                     VALUES ($1, $2, $3, 'TEXT', 'Thanks Bob! The contrast and glow look amazing on OLED displays.', NOW() - INTERVAL '20 minutes', NOW())`, msg2, conv2ID, charlieID); err != nil {
			log.Printf("failed to insert bob-charlie msg2: %v", err)
		}
		fmt.Printf("  ✓ Seeded mutual-follow chat between Bob and Charlie (Conv ID: %s)\n", conv2ID)
	}

	fmt.Println("\n✨ Demo seeding completed successfully!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("👥 Test User Accounts (Password for all: Password123!)")
	fmt.Println("  1. Username: alice    (Mutual with @bob, follows @charlie)")
	fmt.Println("  2. Username: bob      (Mutual with @alice and @charlie)")
	fmt.Println("  3. Username: charlie  (Mutual with @bob, followed by @alice)")
	fmt.Println("  4. Username: david    (Follows @alice, no mutuals)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("💬 Direct Chat Matrix:")
	fmt.Println("  • @alice <-> @bob      : MUTUAL FOLLOW -> CAN CHAT ✅ (Chat seeded with messages)")
	fmt.Println("  • @bob   <-> @charlie  : MUTUAL FOLLOW -> CAN CHAT ✅ (Chat seeded with messages)")
	fmt.Println("  • @alice ->  @charlie  : 1-WAY FOLLOW  -> CANNOT CHAT ⛔ (Charlie must follow back)")
	fmt.Println("  • @david vs  @bob      : NO FOLLOW     -> CANNOT CHAT ⛔")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}
