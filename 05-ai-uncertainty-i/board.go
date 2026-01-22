package main

import "fmt"

var shipTypes = []struct {
	name string
	size int
}{
	{"Carrier", 5},
	{"Battleship", 4},
	{"Cruiser", 3},
	{"Submarine", 3},
	{"Destroyer", 2},
}

type Board [boardSize][boardSize]string

type Position struct {
	row, col int
}

func printBoards(playerBoard, opponentBoardView *Board) {
	// clear the screen
	fmt.Print("\033[H\033[2J")

	fmt.Println("\n=== BATTLESHIP ===")
	fmt.Println()

	fmt.Println("  OPPONENT'S BOARD:")
	fmt.Println(headerRow)

	// print opponents's board
	// player should not see the enemys ships, only hits and misses
	for i := range boardSize {
		fmt.Printf("%d ", i)
		for j := range boardSize {
			cell := opponentBoardView[i][j]
			switch cell {
			case ship:
				// dont show ships on opponent's board
				fmt.Printf("%s ", hiddenShip)
			default:
				fmt.Printf("%s ", cell)
			}
		}
		fmt.Println()
	}

	fmt.Println("\n  YOUR BOARD:")
	fmt.Println(headerRow)

	// print player's board - all information
	for i := range boardSize {
		fmt.Printf("%d ", i)
		for j := range boardSize {
			fmt.Printf("%s ", playerBoard[i][j])
		}
		fmt.Println()
	}

	fmt.Println()
}