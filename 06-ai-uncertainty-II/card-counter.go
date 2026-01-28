package main

const deckSize = 52.0

type CardCounter struct {
	SeenCards      map[string]int // map to keep track of cards by value
	RunningCount   int            // running count for hi-lo strategy
	TrueCount      float64        // true count (running count / cards remaining)
	DecksRemaining float64        // estimate of remaining cards
}

func NewCardCounter() *CardCounter {
	counter := &CardCounter{
		SeenCards:      make(map[string]int),
		RunningCount:   0,
		TrueCount:      0,
		DecksRemaining: 1,
	}

	// initialize counts for each card - value 0
	values := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	for _, value := range values {
		counter.SeenCards[value] = 0
	}

	return counter
}

func (cc *CardCounter) Reset() {
	cc.RunningCount = 0
	cc.TrueCount = 0
	cc.DecksRemaining = 1

	values := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	for _, value := range values {
		cc.SeenCards[value] = 0
	}
}

// TrackCard updates the card counter's state based on a newly seen card
func (cc *CardCounter) TrackCard(card Card) {
	// update seen cards count
	cc.SeenCards[card.Value]++

	// update running count using the hi-lo strategy: 1+ for 2-6, 0 for 7-9, -1 for 10-A
	switch card.Value {
	case "2", "3", "4", "5", "6":
		cc.RunningCount++
	case "10", "J", "Q", "K", "A":
		cc.RunningCount--
	}

	// update decks remaining estimation (52 cards in a deck)
	totalSeen := 0
	for _, count := range cc.SeenCards {
		totalSeen += count
	}

	cc.DecksRemaining = (52.0 - float64(totalSeen)) / deckSize
	if cc.DecksRemaining < 0.1 {
		cc.DecksRemaining = 0.1 // avoid division by too small numbers
	}

	// calculate the true count
	cc.TrueCount = float64(cc.RunningCount) / cc.DecksRemaining
}

func (cc *CardCounter) ChanceOfBusting(playerScore int) float64 {
	// if player has 21 or more, they'll bust on any hit
	if playerScore >= 21 {
		return 1.0
	}

	// calculate how many points until the player busts
	pointsUntilBust := 21 - playerScore

	// count unseen cards that would cause a bust
	bustCards := 0
	totalUnseenCards := 0

	// for each card value, check if it would cause a bust
	values := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	scores := []int{11, 2, 3, 4, 5, 6, 7, 8, 9, 10, 10, 10, 10}

	for i, value := range values {
		// each value appears 4 times
		totalInDeck := 4
		seen := cc.SeenCards[value]
		unseen := max(totalInDeck - seen, 0)

		totalUnseenCards += unseen

		// if this card would cause a bust, count it
		if scores[i] > pointsUntilBust {
			bustCards += unseen
		}
	}

	// avoid division by zero
	if totalUnseenCards == 0 {
		return 0.5 // default to 50% if we somehow have no unseen cards
	}

	return float64(bustCards) / float64(totalUnseenCards)
}

func (cc *CardCounter) DealerChanceOfBusting(dealerUpCard Card) float64 {
	// base probabilities of dealer busting based on up card
	bustProbabilities := map[string]float64{
		"A": 0.17,
		"2": 0.35,
		"3": 0.37,
		"4": 0.40,
		"5": 0.42,
		"6": 0.42,
		"7": 0.26,
		"8": 0.24,
		"9": 0.23,
		"10": 0.21,
		"J": 0.21,
		"Q": 0.21,
		"K": 0.21,
	}

	// we adjust base probability based on our card counting
	baseProbability := bustProbabilities[dealerUpCard.Value]

	// if the true count is positive (more low card has been seen)
	// then the dealer is more likely to bust
	adjustment := cc.TrueCount * 0.02

	return baseProbability + adjustment
}