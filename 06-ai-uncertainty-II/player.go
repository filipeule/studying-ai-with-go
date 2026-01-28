package main

import (
	"fmt"
	"strings"
	"time"
)

const (
	Hit            = "h"
	Stand          = "s"
	Quit           = "q"
	MinDealerStand = 17
)

type Player struct {
	Name   string
	Hand   []Card
	Score  int
	IsAI   bool
	IsBust bool
}

func NewPlayer(name string, isAI bool) Player {
	return Player{
		Name:   name,
		Hand:   []Card{},
		Score:  0,
		IsAI:   isAI,
		IsBust: false,
	}
}

func (p *Player) CalculateScore() int {
	// first, add up all the non ace cards
	nonAceScore := 0
	aces := 0

	// go through each card
	for _, card := range p.Hand {
		if card.Value == "A" {
			// count aces
			aces++
		} else {
			nonAceScore += card.Score
		}
	}

	// handle aces if any
	aceScore := 0
	for range aces {
		// for each ace, try to use eleven first
		// if that would make us go over 21, then use 1 instead
		if nonAceScore+aceScore+11 <= 21 {
			aceScore += 11
		} else {
			aceScore += 1
		}
	}

	return nonAceScore + aceScore
}

func (p *Player) AddCard(card Card, cardCounter *CardCounter) {
	// add the card to their hand
	p.Hand = append(p.Hand, card)

	// update their score
	p.Score = p.CalculateScore()

	// if we're keeping track of cards, update the counter
	if cardCounter != nil {
		cardCounter.TrackCard(card)
	}
}

func (p *Player) DisplayHand(hideSecondCard bool) {
	cards := []string{}

	// go through each card in the player's hand
	for i, card := range p.Hand {
		if hideSecondCard && i > 0 {
			// if we're hiding the second card, show ?? instead
			cards = append(cards, "??")
		} else {
			cards = append(cards, card.String())
		}
	}

	// print their name and all their cards
	fmt.Printf("%s's hand: %s", p.Name, strings.Join(cards, " "))

	// show their score (or ? if we're hiding cards)
	if hideSecondCard {
		fmt.Printf("(Score: ?)\n")
	} else {
		fmt.Printf("(Score: %d)\n", p.Score)
	}
}

func (p *Player) handleHit(deck *Deck, cardCounter *CardCounter) bool {
	// draw a card and add it to the player's hand
	card := deck.Draw()
	p.AddCard(card, cardCounter)

	// show the card that they got
	fmt.Printf("%s drew: %s\n", p.Name, card.String())
	p.DisplayHand(false)

	// check to see if they went over 21
	if p.Score > 21 {
		fmt.Printf("%s busts with a score over 21!\n", p.Name)
		p.IsBust = true
		return true
	}

	// if it is the AI's turn, add a small delay to make it easier to follow
	if p.IsAI {
		time.Sleep(1 * time.Second)
	}

	return false
}

func (p *Player) PlayTurn(deck *Deck, cardCounter *CardCounter, dealerUpCard Card) {
	if p.IsAI {

	} else {
		// if it's a human, let then choose what to do
		p.playHumanTurn(deck, cardCounter)
	}
}

func (p *Player) playHumanTurn(deck *Deck, cardCounter *CardCounter) {
	fmt.Printf("\n--- %s's turn ---\n", p.Name)

	// keep asking then what they want to do? (h)it, (s)tand, (q)uit
	for {
		fmt.Printf("What would you like to do? (h)it, (s)tand, (q)uit: ")
		var choice string
		fmt.Scanln(&choice)
		choice = strings.ToLower(choice)

		switch choice {
		case Quit:
			// they want to quit the game
			fmt.Println("Thanks for playing!")
			return
		case Hit:
			// player wants to hit
			if p.handleHit(deck, cardCounter) {
				return
			}
		case Stand:
			// they're happy with their cards
			fmt.Printf("%s chose to stand\n", p.Name)
			return
		default:
			fmt.Println("Invalid choices. Please try again")
		}
	}
}