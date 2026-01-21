package main

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
	ship     []Ship
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