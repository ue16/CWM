-- Kanban Seed File

-- Drop tables if they exist (for repeatable runs)
DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS cards;
DROP TABLE IF EXISTS columns;
DROP TABLE IF EXISTS boards;
DROP TABLE IF EXISTS users;

-- Create Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    username TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create Boards table
CREATE TABLE boards (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

-- Create Columns table
CREATE TABLE columns (
    id SERIAL PRIMARY KEY,
    board_id INTEGER REFERENCES boards(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    position INTEGER NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

-- Create Cards table
CREATE TABLE cards (
    id SERIAL PRIMARY KEY,
    column_id INTEGER REFERENCES columns(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    position INTEGER NOT NULL,
    attachments INTEGER ARRAY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

-- Create Attachments table
CREATE TABLE attachments (
    id SERIAL PRIMARY KEY,
    card_id INTEGER REFERENCES cards(id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    size INTEGER,
    type TEXT,
    object_id TEXT NOT NULL,
    upload_done BOOLEAN NOT NULL DEFAULT FALSE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

-- Insert sample user (password: password123)
INSERT INTO users (email, password_hash, username) VALUES ('user@example.com', '$2y$10$HovCNynh1dKtG3puJqbPWedlwlfuf8hpA0.Kw9skCBySohfGuuAoq', 'testuser');

-- Insert sample board
INSERT INTO boards (name, user_id) VALUES ('Project Alpha', 1);

-- Insert sample columns for the board
INSERT INTO columns (board_id, name, position, user_id) VALUES (1, 'To Do', 1, 1);
INSERT INTO columns (board_id, name, position, user_id) VALUES (1, 'In Progress', 2, 1);
INSERT INTO columns (board_id, name, position, user_id) VALUES (1, 'Done', 3, 1);

-- Insert sample cards
INSERT INTO cards (column_id, title, description, position, user_id) VALUES (1, 'Set up project', 'Initialize repository and dependencies', 1, 1);
INSERT INTO cards (column_id, title, description, position, user_id) VALUES (2, 'Develop feature X', 'Implement the main feature', 1, 1);
INSERT INTO cards (column_id, title, description, position, user_id) VALUES (3, 'Deploy to production', 'Release the app', 1, 1);
