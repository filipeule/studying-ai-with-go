package main

import (
	"fmt"
	"strings"
)

func main() {
	// clear the screen
	clearScreen()

	// show a welcome message
	fmt.Println("=== Go Blackjack ===")
	fmt.Println("Welcome to Go Blackjack! You're playing against an AI opponent with card counting abilities.")

	// get a deck of cards
	deck := NewDeck().Shuffle()

	// create a card counter of the ai to use
	cardCounter := NewCardCounter()

	for {
		// check to see if we need to shuffle the deck
		if len(deck) < 10 {
			fmt.Println("\n=== Deck is running low. Reshuffling... ===")
			deck = NewDeck().Shuffle()
			cardCounter.Reset()
			fmt.Println("Deck reshuffled and card counter reset")
		}

		// play a round
		PlayRound(&deck, cardCounter)

		// ask if the player wants to play another round
		fmt.Printf("\nPlay another round? (y/n): ")
		var choice string
		fmt.Scanln(&choice)
		choice = strings.ToLower(choice)

		// if not, quit the game
		if choice == "n" {
			fmt.Println("Thanks for playing!")
			break
		}
	}
}