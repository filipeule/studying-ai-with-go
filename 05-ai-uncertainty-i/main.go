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
	playerTurn := true

	for !gameOver {
		// display the boards
		printBoards(human.GetBoard(), ai.GetBoard())

		// players take turn
		if playerTurn {
			fmt.Println("\n=== YOUR TURN ===")
			// let player take turn
			_, _ = human.TakeTurn(ai.GetBoard())

			// check for win condition
			if checkWinCondition(ai.GetBoard()) {
				gameOver = true
				printBoards(human.GetBoard(), ai.GetBoard())
				fmt.Println("\nYOU WIN! YOU SANK ALL ENEMY SHIPS!")
			}
		} else {
			// fmt.Println("Heat map:")
			// for i := range boardSize {
			// 	for j := range boardSize {
			// 		fmt.Printf("%2d  ", ai.heatMap[i][j])
			// 	}
			// 	fmt.Println()
			// }

			fmt.Println("\n=== ENEMY'S TURN ===")
			_, _ = ai.TakeTurn(human.GetBoard())

			// fmt.Println("Press enter to continue...")
			// reader.ReadString('\n')

			if checkWinCondition(human.GetBoard()) {
				gameOver = true
				printBoards(human.GetBoard(), ai.GetBoard())
				fmt.Println("\nYOU LOSE! ALL YOUR SHIPS WERE SUNKEN!")
			}
		}

		// switch turns
		playerTurn = !playerTurn
	}

	fmt.Println("\nThanks for playing Battleship! Press enter to exit...")
	reader.ReadString('\n')
}
