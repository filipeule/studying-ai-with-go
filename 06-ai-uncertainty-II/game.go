package main

import "fmt"

func PlayRound(deck *Deck, cardCounter *CardCounter) {
	fmt.Println("\n=== New Round of Go Blackjack ===")
	fmt.Printf("Cards remaining in deck: %d\n", len(*deck))

	// initialize players
	dealer := NewPlayer("Dealer", false)
	human := NewPlayer("Human", false)
	ai := NewPlayer("AI", true)

	fmt.Println("Players:", dealer.Name, human.Name, ai.Name)

	// initial deal: two cards per player
	for range 2 {
		human.AddCard(deck.Draw(), cardCounter)
		ai.AddCard(deck.Draw(), cardCounter)
		dealer.AddCard(deck.Draw(), cardCounter)
	}

	// show initial hands
	fmt.Println("\nInitial Deal:")
	dealer.DisplayHand(true) // hiding second card
	human.DisplayHand(false)
	ai.DisplayHand(false)

	// play each player's turn
	human.PlayTurn(deck, cardCounter, dealer.Hand[0])
	if !human.IsBust {
		// let ai player play
		ai.PlayTurn(deck, cardCounter, dealer.Hand[0])

		// let dealer play
		dealer.PlayDealerTurn(deck, cardCounter)
	}

	// show results
	fmt.Println("\n=== Results ===")
	fmt.Printf("Dealer: %d\n", dealer.Score)
	fmt.Printf("Human: %d\n", human.Score)
	fmt.Printf("AI: %d\n", ai.Score)

	// display results
	fmt.Println(human.DetermineResult(dealer))
	fmt.Println(ai.DetermineResult(dealer))

	// display card counting statistics
	displayCardCountingStats(deck, cardCounter)
}

func displayCardCountingStats(deck *Deck, cardCounter *CardCounter) {
	fmt.Println("\n=== Card Counting Statistics ===")
	fmt.Printf("Final Running Count: %d\n", cardCounter.RunningCount)
	fmt.Printf("Final True Count: %.2f\n", cardCounter.TrueCount)
	fmt.Printf("Cards remaining in deck: %d\n", len(*deck))

	fmt.Println("\nCard Distribuition Seen:")
	values := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	for _, value := range values {
		fmt.Printf("%s: %d   ", value, cardCounter.SeenCards[value])
		if value == "6" {
			fmt.Println() // put in a line braker for readability
		}
	}
	fmt.Println()
}
