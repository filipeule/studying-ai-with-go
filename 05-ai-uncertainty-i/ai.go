package main

import "math/rand"

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
			p.heatMap[i][j] = 1

			// increase probability in a checkerboard pattern
			if (i+j)%2 == 0 {
				p.heatMap[i][j] += 1
			}

			// higher probability in the center
			centerDistance := abs(i-boardSize/2) + abs(j-boardSize/2)
			if centerDistance <= 3 {
				p.heatMap[i][j] += 2
			}
		}
	}
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
						r, c = row, col + i
					} else {
						r, c = row + i, col
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
						EndPosition: positions[len(positions)-1],
						ShipName: shipType.name,
					}

					p.ships = append(p.ships, newShip)
					break
				}
			}
		}
	}
}
