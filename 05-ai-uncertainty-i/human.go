package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Ship struct {
	ShipName      string
	StartPosition Position
	EndPosition   Position
}

type HumanPlayer struct {
	board    Board
	ships    []Ship
	opponent *AIPlayer
}

func NewHumanPlayer() *HumanPlayer {
	p := &HumanPlayer{}
	for i := range boardSize {
		for j := range boardSize {
			p.board[i][j] = empty
		}
	}
	return p
}

func (p *HumanPlayer) PlaceShips() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n== SHIP PLACEMENT ==")
	fmt.Println("Place your ships on the board")
	fmt.Println("Format: A0 H (A0 starting position, H=horizontal or V=vertical)")
	fmt.Println("Positions are given as letter (A-J) for column and number (0-9) for row")

	for _, shipType := range shipTypes {
		for {
			// display current board
			printBoards(&p.board, &Board{})

			fmt.Printf("\nPlace your %s (length %d): ", shipType.name, shipType.size)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToUpper(input))
			parts := strings.Fields(input)

			// sanity checks
			if len(parts) != 2 {
				fmt.Println("invalid format! use format like 'A0 H'")
				time.Sleep(2 * time.Second)
				continue
			}

			pos := parts[0]
			dir := parts[1]

			if len(pos) < 2 || (dir != "H" && dir != "V") {
				fmt.Println("invalid input! position must be like 'A0' and direction must be 'H' or 'V'")
				time.Sleep(2 * time.Second)
				continue
			}

			// extract column (letter)
			if pos[0] < 'A' || pos[0] > 'J' {
				fmt.Println("column must be between A and J")
				time.Sleep(2 * time.Second)
				continue
			}

			col := int(pos[0] - 'A')

			// extracting row (number)
			rowString := pos[1:]
			row, err := strconv.Atoi(rowString)
			if err != nil || row < 0 || row >= boardSize {
				fmt.Println("row must be between 0 and 9!")
				time.Sleep(2 * time.Second)
				continue
			}

			// check if placement is valid
			valid := true
			positions := []Position{}

			for i := range shipType.size {
				var r, c int
				if dir == "H" {
					r, c = row, col+i	// horizontal placement, so column increases
				} else {
					r, c = row+i, col	// vertical placement, so row increases
				}

				// check if ship would go off the board
				if r >= boardSize || col >= boardSize {
					valid = false
					fmt.Printf("ship would go off of the board! (attempted to place at position %c%d)\n", 'A'+c, r)
					time.Sleep(2 * time.Second)
					break
				}

				// check if position overlaps with another ship
				if p.board[r][c] == ship {
					valid = false
					fmt.Printf("ship would overlap with another ship at position %c%d\n!", 'A'+c, r)
					time.Sleep(2 * time.Second)
					break
				}

				positions = append(positions, Position{r, c})
			}

			if valid {
				// place the ship
				newShip := Ship{
					StartPosition: positions[0],
				}
				for _, pos := range positions {
					p.board[pos.row][pos.col] = ship
				}
				newShip.EndPosition = positions[len(positions)-1]
				newShip.ShipName = shipType.name
				p.ships = append(p.ships, newShip)
				break
			}
		}
	}

	// show final placement
	printBoards(&p.board, &Board{})
	fmt.Println("\nAll ships placed! press Enter to start the game...")
	reader.ReadString('\n')
}
