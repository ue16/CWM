-- Kanban Seed File

-- Drop tables if they exist (for repeatable runs)
DROP TABLE IF EXISTS cards;
DROP TABLE IF EXISTS columns;
DROP TABLE IF EXISTS boards;

-- Create Boards table
CREATE TABLE boards (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

-- Create Columns table
CREATE TABLE columns (
    id SERIAL PRIMARY KEY,
    board_id INTEGER REFERENCES boards(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    position INTEGER NOT NULL
);

-- Create Cards table
CREATE TABLE cards (
    id SERIAL PRIMARY KEY,
    column_id INTEGER REFERENCES columns(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    position INTEGER NOT NULL
);

-- Insert sample board
INSERT INTO boards (name) VALUES ('Project Alpha');

-- Insert sample columns for the board
INSERT INTO columns (board_id, name, position) VALUES (1, 'To Do', 1);
INSERT INTO columns (board_id, name, position) VALUES (1, 'In Progress', 2);
INSERT INTO columns (board_id, name, position) VALUES (1, 'Done', 3);

-- Insert sample cards
INSERT INTO cards (column_id, title, description, position) VALUES (1, 'Set up project', 'Initialize repository and dependencies', 1);
INSERT INTO cards (column_id, title, description, position) VALUES (2, 'Develop feature X', 'Implement the main feature', 1);
INSERT INTO cards (column_id, title, description, position) VALUES (3, 'Deploy to production', 'Release the app', 1);
