package main

import (
	"bufio"
	"fmt"
	"os"
)

const (
	boardSize    = 10
	empty        = "."
	ship         = "O"
	hit          = "X"
	miss         = "~"
	hiddenShip   = "."
	headerRow    = "  A B C D E F G H I J"
	headerColumn = "0123456789"
)

func main() {
	// create players
	human := NewHumanPlayer()
	ai := NewAIPlayer()

	human.opponent = ai
	ai.opponent = human

	// welcome message
	fmt.Println("\n=== WELCOME TO BATTLESHIP ===")
	fmt.Println("Legend:")
	fmt.Printf("  %s - Empty water\n", empty)
	fmt.Printf("  %s - Your ship\n", ship)
	fmt.Printf("  %s - Hit\n", hit)
	fmt.Printf("  %s - Miss\n", miss)
	fmt.Println("\nPress enter to start the game...")

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	// place ships
	// ai places ships (player does not see this)
	ai.PlaceShips()

	// human player places ships
	human.PlaceShips()

	// main game loop
	gameOver := false
	// playerTurn := true

	for !gameOver {
		// display the boards

		// players take turn

		// switch turns

		// check win condition
	}

	fmt.Println("\nThanks for playing Battleship! Press enter to exit...")
	reader.ReadString('\n')
}
