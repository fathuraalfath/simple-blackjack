package main

import (
	"testing"

	"simple-blackjack/models"
)

func TestCalculateHandScore(t *testing.T) {
	tests := []struct {
		name     string
		cards    []models.Card
		expected int
	}{
		{
			name: "Standard cards without Ace",
			cards: []models.Card{
				{Rank: "10", Value: 10},
				{Rank: "7", Value: 7},
			},
			expected: 17,
		},
		{
			name: "Face cards J and Q",
			cards: []models.Card{
				{Rank: "J", Value: 10},
				{Rank: "Q", Value: 10},
			},
			expected: 20,
		},
		{
			name: "Soft Ace hand (Ace counts as 11)",
			cards: []models.Card{
				{Rank: "A", Value: 11},
				{Rank: "9", Value: 9},
			},
			expected: 20,
		},
		{
			name: "Hard Ace hand (Ace drops to 1 to prevent bust)",
			cards: []models.Card{
				{Rank: "A", Value: 11},
				{Rank: "9", Value: 9},
				{Rank: "5", Value: 5},
			},
			expected: 15,
		},
		{
			name: "Multiple Aces (one 11, one 1)",
			cards: []models.Card{
				{Rank: "A", Value: 11},
				{Rank: "A", Value: 11},
				{Rank: "9", Value: 9},
			},
			expected: 21,
		},
		{
			name: "Natural Blackjack (Ace and King)",
			cards: []models.Card{
				{Rank: "A", Value: 11},
				{Rank: "K", Value: 10},
			},
			expected: 21,
		},
		{
			name: "Four Aces",
			cards: []models.Card{
				{Rank: "A", Value: 11},
				{Rank: "A", Value: 11},
				{Rank: "A", Value: 11},
				{Rank: "A", Value: 11},
			},
			expected: 14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := models.CalculateHandScore(tt.cards)
			if score != tt.expected {
				t.Errorf("expected score %d, got %d for cards %+v", tt.expected, score, tt.cards)
			}
		})
	}
}
