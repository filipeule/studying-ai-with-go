package main

import "fmt"

func main() {
	// clear the screen
	clearScreen()

	// show a welcome message
	fmt.Println("=== Go Blackjack ===")
	fmt.Println("Welcome to Go Blackjack! You're playing against an AI opponent with card counting abilities.")

	// get a deck of cards
	deck := NewDeck().Shuffle()

	for _, c := range deck {
		fmt.Printf("%s%s\n", c.Value, c.Suit)
	}

	// create a card counter of the ai to use

	for {
		// check to see if we need to shuffle the deck

		// play a round

		// ask if the player wants to play another round

		// if not, quit the game

		// clear the screen
	}
}