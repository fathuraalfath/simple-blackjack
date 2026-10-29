package controller

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"simple-blackjack/config"
	"simple-blackjack/models"

	"github.com/gin-gonic/gin"
)

var (
	suits = []string{"H", "D", "C", "S"}
	ranks = []struct {
		rank  string
		value int
	}{
		{"2", 2}, {"3", 3}, {"4", 4}, {"5", 5}, {"6", 6},
		{"7", 7}, {"8", 8}, {"9", 9}, {"10", 10},
		{"J", 10}, {"Q", 10}, {"K", 10}, {"A", 11},
	}
)

type StartGameInput struct {
	Bet int `json:"bet" binding:"required"`
}

func generateFullDeck() []models.Card {
	var deck []models.Card
	for _, s := range suits {
		for _, r := range ranks {
			deck = append(deck, models.Card{
				Suit:  s,
				Rank:  r.rank,
				Value: r.value,
			})
		}
	}
	return deck
}

func drawFromRemaining(usedCards []models.Card) models.Card {
	usedMap := make(map[string]bool)
	for _, c := range usedCards {
		usedMap[c.Suit+c.Rank] = true
	}

	var available []models.Card
	for _, c := range generateFullDeck() {
		if !usedMap[c.Suit+c.Rank] {
			available = append(available, c)
		}
	}

	if len(available) == 0 {
		available = generateFullDeck()
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return available[r.Intn(len(available))]
}

func StartGame(c *gin.Context) {
	currentUser := c.MustGet("currentUser").(models.User)

	var input StartGameInput
	if err := c.ShouldBindJSON(&input); err != nil || input.Bet <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bet must be greater than 0"})
		return
	}

	if err := config.DB.First(&currentUser, currentUser.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user"})
		return
	}

	if currentUser.Coin < input.Bet {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient coin balance"})
		return
	}

	// Deduct coin immediately upon starting game
	currentUser.Coin -= input.Bet
	if err := config.DB.Save(&currentUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deduct coin"})
		return
	}

	deck := generateFullDeck()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	playerCards := []models.Card{deck[0], deck[1]}
	dealerCards := []models.Card{deck[2], deck[3]}

	playerScore := models.CalculateHandScore(playerCards)
	dealerScore := models.CalculateHandScore(dealerCards)

	status := "IN_PROGRESS"

	// Natural blackjack check
	if playerScore == 21 {
		if dealerScore == 21 {
			status = "DRAW"
			currentUser.Coin += input.Bet
		} else {
			status = "PLAYER_WON"
			// Blackjack payout 2.5x (original bet + 1.5x profit)
			payout := int(float64(input.Bet) * 2.5)
			currentUser.Coin += payout
		}
		_ = config.DB.Save(&currentUser)
	}

	pBytes, _ := json.Marshal(playerCards)
	dBytes, _ := json.Marshal(dealerCards)

	game := models.Game{
		UserID:      currentUser.ID,
		BetAmount:   input.Bet,
		PlayerCards: string(pBytes),
		DealerCards: string(dBytes),
		PlayerScore: playerScore,
		DealerScore: dealerScore,
		Status:      status,
	}

	if err := config.DB.Create(&game).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start game"})
		return
	}

	c.JSON(http.StatusCreated, game.ToResponse(true))
}

func Hit(c *gin.Context) {
	currentUser := c.MustGet("currentUser").(models.User)
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}

	var game models.Game
	if err := config.DB.Where("id = ? AND user_id = ?", gameID, currentUser.ID).First(&game).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	if game.Status != "IN_PROGRESS" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game has already ended", "status": game.Status})
		return
	}

	var playerCards, dealerCards []models.Card
	_ = json.Unmarshal([]byte(game.PlayerCards), &playerCards)
	_ = json.Unmarshal([]byte(game.DealerCards), &dealerCards)

	var used []models.Card
	used = append(used, playerCards...)
	used = append(used, dealerCards...)

	newCard := drawFromRemaining(used)
	playerCards = append(playerCards, newCard)
	playerScore := models.CalculateHandScore(playerCards)

	game.PlayerScore = playerScore
	if playerScore > 21 {
		game.Status = "DEALER_WON" // Player bust
	}

	pBytes, _ := json.Marshal(playerCards)
	game.PlayerCards = string(pBytes)

	if err := config.DB.Save(&game).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update game"})
		return
	}

	c.JSON(http.StatusOK, game.ToResponse(game.Status == "IN_PROGRESS"))
}

func Stand(c *gin.Context) {
	currentUser := c.MustGet("currentUser").(models.User)
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}

	var game models.Game
	if err := config.DB.Where("id = ? AND user_id = ?", gameID, currentUser.ID).First(&game).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	if game.Status != "IN_PROGRESS" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game has already ended", "status": game.Status})
		return
	}

	var playerCards, dealerCards []models.Card
	_ = json.Unmarshal([]byte(game.PlayerCards), &playerCards)
	_ = json.Unmarshal([]byte(game.DealerCards), &dealerCards)

	var used []models.Card
	used = append(used, playerCards...)
	used = append(used, dealerCards...)

	// Dealer draws until score >= 17
	dealerScore := models.CalculateHandScore(dealerCards)
	for dealerScore < 17 {
		card := drawFromRemaining(used)
		dealerCards = append(dealerCards, card)
		used = append(used, card)
		dealerScore = models.CalculateHandScore(dealerCards)
	}

	game.DealerScore = dealerScore
	dBytes, _ := json.Marshal(dealerCards)
	game.DealerCards = string(dBytes)

	// Determine winner
	if dealerScore > 21 {
		// Dealer bust -> player wins 2x
		game.Status = "PLAYER_WON"
	} else if game.PlayerScore > dealerScore {
		// Player higher score -> player wins 2x
		game.Status = "PLAYER_WON"
	} else if game.PlayerScore == dealerScore {
		// Push / draw -> bet refunded
		game.Status = "DRAW"
	} else {
		game.Status = "DEALER_WON"
	}

	// Update payout if player won or draw
	if game.Status == "PLAYER_WON" || game.Status == "DRAW" {
		_ = config.DB.First(&currentUser, currentUser.ID)
		if game.Status == "PLAYER_WON" {
			currentUser.Coin += game.BetAmount * 2
		} else {
			currentUser.Coin += game.BetAmount
		}
		_ = config.DB.Save(&currentUser)
	}

	if err := config.DB.Save(&game).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to finalize game"})
		return
	}

	c.JSON(http.StatusOK, game.ToResponse(false))
}

func GetGame(c *gin.Context) {
	currentUser := c.MustGet("currentUser").(models.User)
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}

	var game models.Game
	if err := config.DB.Where("id = ? AND user_id = ?", gameID, currentUser.ID).First(&game).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	c.JSON(http.StatusOK, game.ToResponse(game.Status == "IN_PROGRESS"))
}

func ListGames(c *gin.Context) {
	currentUser := c.MustGet("currentUser").(models.User)

	var games []models.Game
	if err := config.DB.Where("user_id = ?", currentUser.ID).Order("id desc").Find(&games).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list games"})
		return
	}

	var res []models.GameResponse
	for _, g := range games {
		res = append(res, g.ToResponse(g.Status == "IN_PROGRESS"))
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(res),
		"games": res,
	})
}

func DeleteGame(c *gin.Context) {
	currentUser := c.MustGet("currentUser").(models.User)
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}

	var game models.Game
	if err := config.DB.Where("id = ? AND user_id = ?", gameID, currentUser.ID).First(&game).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	if err := config.DB.Delete(&game).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete game history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "game history deleted successfully",
		"game_id": gameID,
	})
}
