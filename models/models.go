package models

import (
	"encoding/json"
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Password  string    `gorm:"not null" json:"-"`
	Coin      int       `gorm:"not null;default:0" json:"coin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Card struct {
	Suit  string `json:"suit"`  // H (Hearts), D (Diamonds), C (Clubs), S (Spades)
	Rank  string `json:"rank"`  // 2, 3, ..., 10, J, Q, K, A
	Value int    `json:"value"` // Default face value
}

type Game struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	BetAmount   int       `gorm:"not null" json:"bet_amount"`
	PlayerCards string    `gorm:"type:text;not null" json:"player_cards"` // JSON string
	DealerCards string    `gorm:"type:text;not null" json:"dealer_cards"` // JSON string
	PlayerScore int       `gorm:"not null" json:"player_score"`
	DealerScore int       `gorm:"not null" json:"dealer_score"`
	Status      string    `gorm:"size:20;not null" json:"status"` // IN_PROGRESS, PLAYER_WON, DEALER_WON, DRAW
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GameResponse struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	BetAmount   int       `json:"bet_amount"`
	PlayerCards []Card    `json:"player_cards"`
	DealerCards []Card    `json:"dealer_cards"`
	PlayerScore int       `json:"player_score"`
	DealerScore int       `json:"dealer_score"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (g *Game) ToResponse(hideDealerHoleCard bool) GameResponse {
	var pCards, dCards []Card
	_ = json.Unmarshal([]byte(g.PlayerCards), &pCards)
	_ = json.Unmarshal([]byte(g.DealerCards), &dCards)

	displayedDealerCards := dCards
	displayedDealerScore := g.DealerScore

	if hideDealerHoleCard && g.Status == "IN_PROGRESS" && len(dCards) > 1 {
		displayedDealerCards = []Card{
			dCards[0],
			{Suit: "?", Rank: "?", Value: 0},
		}
		displayedDealerScore = CalculateHandScore([]Card{dCards[0]})
	}

	return GameResponse{
		ID:          g.ID,
		UserID:      g.UserID,
		BetAmount:   g.BetAmount,
		PlayerCards: pCards,
		DealerCards: displayedDealerCards,
		PlayerScore: g.PlayerScore,
		DealerScore: displayedDealerScore,
		Status:      g.Status,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}

func CalculateHandScore(cards []Card) int {
	total := 0
	aces := 0

	for _, c := range cards {
		switch c.Rank {
		case "A":
			aces++
			total += 11
		case "K", "Q", "J":
			total += 10
		default:
			total += c.Value
		}
	}

	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}

	return total
}
