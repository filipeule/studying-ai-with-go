package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
)

const (
	// heapmap weights
	baseProbability      = 1  // starting probability for valid cells
	checkerBoardBonus    = 1  // bonus for the checkerboard pattern (less likely adjacent placements)
	centerProximityBonus = 2  // bonus for cell closer to the center
	maxCenterDistance    = 3  // how far from the center qualifies for the bonus
	huntModeBoost        = 15 // significant bonus for cells adjacent to hits in hunt mode
	shipFitBonus         = 2  // base bonus multiplier for fitting a ship
)

type AIPlayer struct {
	board          Board
	heatMap        [boardSize][boardSize]int
	hits           []Position
	shipsSunk      int
	huntMode       bool
	potentialShips []struct {
		size    int
		sunk    bool
		hits    int
		shipPos []Position
	}
	ships    []Ship
	opponent *HumanPlayer
}

func NewAIPlayer() *AIPlayer {
	p := &AIPlayer{
		shipsSunk: 0,
		huntMode:  false,
	}

	for i := range boardSize {
		for j := range boardSize {
			p.board[i][j] = empty
		}
	}

	p.initializeHeatMap()

	// initialize potential ships tracking
	p.potentialShips = make([]struct {
		size    int
		sunk    bool
		hits    int
		shipPos []Position
	}, len(shipTypes))

	for i, shipType := range shipTypes {
		p.potentialShips[i].size = shipType.size
		p.potentialShips[i].sunk = false
		p.potentialShips[i].hits = 0
		p.potentialShips[i].shipPos = make([]Position, 0)
	}

	return p
}

func (p *AIPlayer) initializeHeatMap() {
	for i := range boardSize {
		for j := range boardSize {
			// start with base probability
			p.heatMap[i][j] = baseProbability

			// increase probability in a checkerboard pattern
			if (i+j)%2 == 0 {
				p.heatMap[i][j] += checkerBoardBonus
			}

			// higher probability in the center
			centerDistance := abs(i-boardSize/2) + abs(j-boardSize/2)
			if centerDistance <= maxCenterDistance {
				p.heatMap[i][j] += centerProximityBonus
			}
		}
	}
}

// updateHeatMap recalculates the probability heat map based on the current game state
// it considers potential ship placements and prioritizes targets during hunt mode
func (p *AIPlayer) updateHeatMap(opponentBoard *Board) {
	// reset hat map (clear previous probabilities)
	p.initializeHeatMap()

	// calculate base probabilities & ship fit probabilities
	for r := range boardSize {
		for c := range boardSize {
			// skip cells that have already been targeted
			if opponentBoard[r][c] == hit || opponentBoard[r][c] == miss {
				continue
			}

			// assign base probability for valid, untargeted cells
			p.heatMap[r][c] = baseProbability

			// iterate through opponents ships that have not been sunk yet
			for _, shipData := range p.potentialShips {
				if shipData.sunk {
					continue
				}

				shipSize := shipData.size

				// check horizontal fit: can an unsunk ship of this size fit horizontally starting here?
				if c+shipSize <= boardSize {
					canFitHorizontal := true
					for k := range shipSize {
						// check to see if any cell needed for the ship is already a miss or hit
						if opponentBoard[r][c+k] == miss || opponentBoard[r][c+k] == hit {
							canFitHorizontal = false
							break
						}
					}
					if canFitHorizontal {
						// increase probability based on shipsize if it fits
						for k := range shipSize {
							p.heatMap[r][c+k] += shipFitBonus
						}
					}
				}

				// check vertical fit
				if r+shipSize <= boardSize {
					canFitVertical := true
					for k := range shipSize {
						// check to see if any cell need for the ship is already a miss or hit
						if opponentBoard[r+k][c] == miss || opponentBoard[r+k][c] == hit {
							canFitVertical = false
							break
						}
					}
					if canFitVertical {
						for k := range shipSize {
							p.heatMap[r+k][c] += shipFitBonus
						}
					}
				}
			}
		}
	}

	// apply hunt mode boost if applicable
	if p.huntMode && len(p.hits) > 0 {
		p.applyHuntModeBoosts(opponentBoard)
	}
}

func (p *AIPlayer) applyHuntModeBoosts(opponentBoard *Board) {
	// determine the hit pattern: single hit, horizontal line, vertical line
	isSingleHit := len(p.hits) == 1
	isHorizontal := false
	isVertical := false

	if !isSingleHit {
		firstHit := p.hits[0]
		isHorizontal = true
		isVertical = true

		for i := 1; i < len(p.hits); i++ {
			if p.hits[i].row != firstHit.row {
				isHorizontal = false
			}
			if p.hits[i].col != firstHit.col {
				isVertical = false
			}
		}

		// if hits are not aligned horizontally or vertically, treat as multiple single points
		// for adjancent checks
		if !isHorizontal && !isVertical {
			isSingleHit = true // fallback to checking adjacent cells for all hits if not clearly aligned
		}

		boostCell := func(r, c int) {
			if r >= 0 && r < boardSize && c >= 0 && c < boardSize &&
				opponentBoard[r][c] != miss && opponentBoard[r][c] != hit {
				p.heatMap[r][c] += huntModeBoost
			}
		}

		if isSingleHit {
			// boost all valid neighbors of the hit(s)
			for _, hitPos := range p.hits {
				boostCell(hitPos.row-1, hitPos.col) // up
				boostCell(hitPos.row+1, hitPos.col) // down
				boostCell(hitPos.row, hitPos.col-1) // left
				boostCell(hitPos.row, hitPos.col+1) // right
			}
		} else if isHorizontal {
			// boost cells to the left and right of the horizontal line of hits
			row := p.hits[0].row
			minCol, maxCol := p.hits[0].col, p.hits[0].col
			for _, hitPos := range p.hits {
				if hitPos.col < minCol {
					minCol = hitPos.col
				}
				if hitPos.col > maxCol {
					maxCol = hitPos.col
				}
			}
			boostCell(row, minCol-1) // left of the line
			boostCell(row, maxCol+1) // right of the line
		} else if isVertical {
			// boost cells above and below of the vertical line of hits
			col := p.hits[0].col
			minRow, maxRow := p.hits[0].row, p.hits[0].row
			for _, hitPos := range p.hits {
				if hitPos.row < minRow {
					minRow = hitPos.row
				}
				if hitPos.row > maxRow {
					maxRow = hitPos.row
				}
			}
			boostCell(minRow-1, col) // above the line
			boostCell(maxRow+1, col) // below the line
		}
	}
}

func (p *AIPlayer) TakeTurn(opponentBoard *Board) (Position, bool) {
	fmt.Println("\nEnemy is taking it's turn...")
	if p.huntMode {
		fmt.Println("AI is in hunting mode")
	} else {
		fmt.Println("AI is in probability target mode")
	}

	// update heat map based on game state
	p.updateHeatMap(opponentBoard)

	// select a target based on strategy
	var targetRow, targetCol int

	if p.huntMode {
		// find the highest probability cell(s)
		maxProb := 0
		candidates := []Position{}

		for i := 0; i < boardSize; i++ {
			for j := 0; j < boardSize; j++ {
				if p.heatMap[i][j] > maxProb && opponentBoard[i][j] != hit && opponentBoard[i][j] != miss {
					maxProb = p.heatMap[i][j]
					candidates = []Position{{i, j}}
				} else if p.heatMap[i][j] == maxProb && opponentBoard[i][j] != hit && opponentBoard[i][j] != miss {
					candidates = append(candidates, Position{i, j})
				}
			}
		}

		// select a random target from highest probability cell
		if len(candidates) > 0 {
			selected := candidates[rand.Intn(len(candidates))]
			targetRow, targetCol = selected.row, selected.col
		} else {
			// if cant find one, find back to random targeting
			for {
				// fallback to random targeting
				targetRow = rand.Intn(boardSize)
				targetCol = rand.Intn(boardSize)
				if opponentBoard[targetRow][targetCol] != hit && opponentBoard[targetRow][targetCol] != miss {
					break
				}
			}
		}

	} else {
		// find the highest probability cell(s)
		maxProb := 0
		candidates := []Position{}

		for i := range boardSize {
			for j := range boardSize {
				if p.heatMap[i][j] > maxProb && opponentBoard[i][j] != miss && opponentBoard[i][j] != hit {
					maxProb = p.heatMap[i][j]
					candidates = []Position{{i, j}}
				} else if p.heatMap[i][j] == maxProb && opponentBoard[i][j] != miss && opponentBoard[i][j] != hit {
					candidates = append(candidates, Position{i, j})
				}
			}
		}

		// select a random target from cells
		if len(candidates) > 0 {
			selected := candidates[rand.Intn(len(candidates))]
			targetRow, targetCol = selected.row, selected.col
		} else {
			// if cant find one, find back to random targeting
			for {
				// fallback to random targeting
				targetRow = rand.Intn(boardSize)
				targetCol = rand.Intn(boardSize)
				if opponentBoard[targetRow][targetCol] != hit && opponentBoard[targetRow][targetCol] != miss {
					break
				}
			}
		}
	}

	// perform the attack
	isHit := opponentBoard[targetRow][targetCol] == ship

	// check to see if we hit
	if isHit {
		opponentBoard[targetRow][targetCol] = hit
		fmt.Printf("Enemy targets %c%d... HIT!\n", 'A'+targetCol, targetRow)

		// update hit tracking
		p.hits = append(p.hits, Position{targetRow, targetCol})

		// if we hit, enter hunt mode
		p.huntMode = true

		// check to see if ship was sunk
		sunk, shipName := isShipSunk(opponentBoard, targetRow, targetCol, p.opponent.ships)
		if sunk {
			fmt.Printf("Enemy sunk your %s!\n", shipName)
			p.shipsSunk++
			p.huntMode = false
			p.hits = []Position{}
		}
	} else {
		opponentBoard[targetRow][targetCol] = miss
		fmt.Printf("Enemy targets %c%d... MISS!\n", 'A'+targetCol, targetRow)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Press enter to continue...")
	reader.ReadString('\n')

	return Position{targetRow, targetCol}, isHit
}

func (p *AIPlayer) GetBoard() *Board {
	return &p.board
}

func (p *AIPlayer) PlaceShips() {
	// place ships in a mix and edge of center clusters
	// attempt to place larger ships near the edges
	for i, shipType := range shipTypes {
		placed := false
		attempts := 0

		for !placed && attempts < 100 {
			attempts++

			// decide on placement strategy based on ship size
			var row, col int
			horizontal := rand.Intn(2) == 0

			if shipType.size >= 4 {
				if horizontal {
					row = rand.Intn(boardSize)
					if rand.Intn(2) == 0 {
						// near the left edge
						col = rand.Intn(3)
					} else {
						// near right edge
						col = boardSize - shipType.size - rand.Intn(3)
					}
				} else {
					col = rand.Intn(boardSize)
					if rand.Intn(2) == 0 {
						// near the top edge
						row = rand.Intn(3)
					} else {
						// near the bottom edge
						row = boardSize - shipType.size - rand.Intn(3)
					}
				}
			} else {
				// place smaller ships in a more distributed pattern
				if horizontal {
					row = rand.Intn(boardSize)
					col = rand.Intn(boardSize - shipType.size + 1)
				} else {
					row = rand.Intn(boardSize - shipType.size + 1)
					col = rand.Intn(boardSize)
				}
			}

			// check if placement is valid
			positions := []Position{}
			valid := true

			for j := range shipType.size {
				var r, c int
				if horizontal {
					r, c = row, col+j
				} else {
					r, c = row+j, col
				}

				// check validity
				if r < 0 || r >= boardSize || c < 0 || c >= boardSize || p.board[r][c] == ship {
					valid = false
					break
				}

				// check surrounding cells to avoid placing ships adjacent to one another
				for dr := -1; dr <= 1; dr++ {
					for dc := -1; dc <= 1; dc++ {
						nr, nc := r+dr, c+dc
						if nr >= 0 && nr < boardSize && nc >= 0 && nc < boardSize && p.board[nr][nc] == ship && (dr != 0 || dc != 0) {
							// avoid placing ships diagonally or directly adjacent to another ship
							if i < 2 {
								valid = false
								break
							}
						}
					}
					if !valid {
						break
					}
				}
				if !valid {
					break
				}

				positions = append(positions, Position{r, c})
			}

			if valid {
				// place the ship
				for _, pos := range positions {
					p.board[pos.row][pos.col] = ship
				}

				// append this ship to the slice of ai ships
				newShip := Ship{
					StartPosition: positions[0],
					EndPosition:   positions[len(positions)-1],
					ShipName:      shipType.name,
				}

				p.ships = append(p.ships, newShip)
				placed = true
			}
		}

		// if we couldn't place the ship, fall back to rand placement
		if !placed {
			for {
				horizontal := rand.Intn(2) == 0
				var row, col int
				if horizontal {
					row = rand.Intn(boardSize)
					col = rand.Intn(boardSize - shipType.size + 1)
				} else {
					row = rand.Intn(boardSize - shipType.size + 1)
					col = rand.Intn(boardSize)
				}

				// check if placement is valid
				valid := true
				positions := []Position{}

				for i := range shipType.size {
					var r, c int
					if horizontal {
						r, c = row, col+i
					} else {
						r, c = row+i, col
					}

					if p.board[r][c] == ship {
						valid = false
						break
					}

					positions = append(positions, Position{r, c})
				}

				if valid {
					// place ship
					for _, pos := range positions {
						p.board[pos.row][pos.col] = ship
					}

					// append to slice of ai's ships
					newShip := Ship{
						StartPosition: positions[0],
						EndPosition:   positions[len(positions)-1],
						ShipName:      shipType.name,
					}

					p.ships = append(p.ships, newShip)
					break
				}
			}
		}
	}
}
