// Kanban MongoDB Initialization Script
// This script creates the MongoDB user, database, and seeds sample data
// Usage: mongosh < kanban.mongodb.js
// Or with credentials: mongosh -u admin -p <password> --authenticationDatabase admin < kanban.mongodb.js

// Switch to the kanban database
db = db.getSiblingDB("kanban");

// Drop existing collections (for repeatable runs)
db.users.drop();
db.boards.drop();
db.columns.drop();
db.cards.drop();
db.attachments.drop();
db.counters.drop();

print("✓ Dropped existing collections");

// Initialize the counters collection for auto-incrementing IDs
db.counters.insertMany([
	{ _id: "users", sequence: 0 },
	{ _id: "boards", sequence: 0 },
	{ _id: "columns", sequence: 0 },
	{ _id: "cards", sequence: 0 },
	{ _id: "attachments", sequence: 0 },
]);

print("✓ Initialized counters collection");

// Create indexes for better query performance
db.users.createIndex({ username: 1 }, { unique: true });
db.users.createIndex({ email: 1 }, { unique: true });
db.boards.createIndex({ user_id: 1 });
db.columns.createIndex({ board_id: 1 });
db.columns.createIndex({ board_id: 1, position: 1 });
db.cards.createIndex({ column_id: 1 });
db.cards.createIndex({ column_id: 1, position: 1 });
db.attachments.createIndex({ card_id: 1 });
db.attachments.createIndex({ card_id: 1, upload_done: 1 });

print("✓ Created database indexes");

// Insert sample user (password: password123)
// Hash: $2y$10$HovCNynh1dKtG3puJqbPWedlwlfuf8hpA0.Kw9skCBySohfGuuAoq
db.users.insertOne({
	id: 1,
	username: "testuser",
	email: "user@example.com",
	password_hash: "$2y$10$HovCNynh1dKtG3puJqbPWedlwlfuf8hpA0.Kw9skCBySohfGuuAoq",
	created_at: new Date(),
});

// Update counter after inserting user
db.counters.updateOne({ _id: "users" }, { $set: { sequence: 1 } });

print("✓ Inserted sample user (testuser)");

// Insert sample board
db.boards.insertOne({
	id: 1,
	name: "Project Alpha",
	user_id: 1,
});

// Update counter after inserting board
db.counters.updateOne({ _id: "boards" }, { $set: { sequence: 1 } });

print("✓ Inserted sample board");

// Insert sample columns for the board
db.columns.insertMany([
	{
		id: 1,
		board_id: 1,
		name: "To Do",
		position: 1,
		user_id: 1,
	},
	{
		id: 2,
		board_id: 1,
		name: "In Progress",
		position: 2,
		user_id: 1,
	},
	{
		id: 3,
		board_id: 1,
		name: "Done",
		position: 3,
		user_id: 1,
	},
]);

// Update counter after inserting columns
db.counters.updateOne({ _id: "columns" }, { $set: { sequence: 3 } });

print("✓ Inserted sample columns");

// Insert sample cards
db.cards.insertMany([
	{
		id: 1,
		column_id: 1,
		title: "Set up project",
		description: "Initialize repository and dependencies",
		position: 1,
		user_id: 1,
	},
	{
		id: 2,
		column_id: 2,
		title: "Develop feature X",
		description: "Implement the main feature",
		position: 1,
		user_id: 1,
	},
	{
		id: 3,
		column_id: 3,
		title: "Deploy to production",
		description: "Release the app",
		position: 1,
		user_id: 1,
	},
]);

// Update counter after inserting cards
db.counters.updateOne({ _id: "cards" }, { $set: { sequence: 3 } });
