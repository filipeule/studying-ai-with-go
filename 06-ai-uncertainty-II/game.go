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

	// show results

	// display results

	// display card counting statistics
}
