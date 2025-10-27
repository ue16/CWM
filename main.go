package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"oskar/cwm25/internal"
)

func main() {
	text := dbTest()
	internal.Serve(":80", text)
}

func dbTest() string {
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("DATABASE_URL env var not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := internal.NewDB(ctx, connString)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	// Create example board
	board, err := db.CreateBoard(ctx, "Demo Board")
	if err != nil {
		log.Fatalf("CreateBoard: %v", err)
	}
	fmt.Printf("Created board: %s\n", board)

	// Create example columns
	columnNames := []string{"To Do", "In Progress", "Done"}
	var columns []internal.Column
	for i, name := range columnNames {
		col, err := db.CreateColumn(ctx, internal.Column{
			BoardID:  board.ID,
			Name:     name,
			Position: i + 1,
		})
		if err != nil {
			log.Fatalf("CreateColumn %q: %v", name, err)
		}
		columns = append(columns, col)
		fmt.Printf(" Created column: %s\n", col)
	}

	// Create example cards
	cardSamples := []internal.Card{
		{ColumnID: columns[0].ID, Title: "Set up project", Description: "Initialize repo and tooling", Position: 1},
		{ColumnID: columns[1].ID, Title: "Develop feature X", Description: "Implement core functionality and tests", Position: 1},
		{ColumnID: columns[2].ID, Title: "Deploy to production", Description: "Run deployment pipeline", Position: 1},
	}

	for _, cs := range cardSamples {
		c, err := db.CreateCard(ctx, cs)
		if err != nil {
			log.Fatalf("CreateCard %q: %v", cs.Title, err)
		}
		fmt.Printf("  Created card: %s\n", c)
	}

	// Fetch and return all boards, columns and cards as a string
	out := ""
	out += "\nCurrent DB contents:\n"
	boards, err := db.ListBoards(ctx)
	if err != nil {
		log.Fatalf("ListBoards: %v", err)
	}
	for _, b := range boards {
		out += fmt.Sprintf("%v\n", b)
		cols, err := db.GetColumnsByBoardID(ctx, b.ID)
		if err != nil {
			log.Fatalf("GetColumnsByBoardID(%d): %v", b.ID, err)
		}
		for _, col := range cols {
			out += fmt.Sprintf("  %v\n", col)
			cards, err := db.GetCardsByColumnID(ctx, col.ID)
			if err != nil {
				log.Fatalf("GetCardsByColumnID(%d): %v", col.ID, err)
			}
			for _, card := range cards {
				out += fmt.Sprintf("    %v\n", card)
			}
		}
	}
	out += "\nDone.\n"
	return out
}
