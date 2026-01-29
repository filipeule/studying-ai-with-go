package main

import (
	"fmt"
	"time"
)

func AdvancedAIDecision(player Player, dealerUpCard Card, cardCounter *CardCounter) string {
	score := player.Score

	// always going to stand on 19 or higher
	if score >= 19 {
		return Stand
	}

	// calculate bust probability if we hit
	bustProbability := cardCounter.ChanceOfBusting(score)

	// calculate dealer bust probability
	dealerBustProbability := cardCounter.DealerChanceOfBusting(dealerUpCard)

	// display ai thinking
	fmt.Printf(
		"AI thinking: score %d, True Count %.1f, Bust probability %.1f%%, Dealer Bust probability %.1f%%\n",
		score,
		cardCounter.TrueCount,
		bustProbability * 100,
		dealerBustProbability * 100,
	)

	time.Sleep(500 * time.Millisecond)

	// decision logic incorporating card counting
	if score >= 17 {
		// with 17 or 18, consider the true count
		if cardCounter.TrueCount > 0 {
			// positive means more high cards left, so stand
			return Stand
		} else if dealerUpCard.Score >= 7 && cardCounter.TrueCount < -2 {
			// against a strong dealer, with a very negative count, so hit on 17
			if score == 17 {
				return Hit
			}
			return Stand
		}
		return Stand
	}

	// soft hands (have an ace counted as 11)
	hasAce := false
	for _, card := range player.Hand {
		if card.Value == "A" && score <= 21 {
			hasAce = true
			break
		}
	}

	if hasAce {
		// soft 18 or higher
		if score >= 18 {
			// stand unless dealer has a 9, 10 or Ace with negative count
			if (dealerUpCard.Score >= 9 || dealerUpCard.Value == "A") && cardCounter.TrueCount < -1 {
				return Hit
			} else {
				return Stand
			}
		}

		// soft 17 or lower, always hit
		return Hit
	}

	// 16 or lower with considerations
	if score <= 16 {
		// if low risk of busting and dealer is likely to bust, then stand
		if bustProbability < 0.3 && dealerBustProbability > 0.4 && score >= 13 {
			return Stand
		}

		// stand on 12-16 against dealer 2-6 (unless the true count is very negative)
		if score >= 12 && dealerUpCard.Score >= 2 && dealerUpCard.Score <= 6 && cardCounter.TrueCount > -3 {
			return Stand
		}

		// otherwise
		return Hit
	}

	// default to stand
	return Stand
}