package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"cosine/auth"
	"cosine/config"
	"cosine/database"

	"github.com/gin-gonic/gin"
)

// generateState generates a random state string for OAuth
func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// LinuxDoAuthURLHandler returns the OAuth authorization URL
// GET /api/auth/linuxdo/url
func LinuxDoAuthURLHandler(c *gin.Context) {
	cfg := config.GlobalConfig
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "config not loaded"})
		return
	}

	state := generateState()
	url := auth.LinuxDoAuthCodeURL(cfg, state)

	c.JSON(http.StatusOK, gin.H{
		"url":   url,
		"state": state,
	})
}

// LinuxDoCallbackHandler handles the OAuth callback
// GET /api/auth/linuxdo/callback
func LinuxDoCallbackHandler(c *gin.Context) {
	cfg := config.GlobalConfig
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "config not loaded"})
		return
	}

	code := c.Query("code")
	state := c.Query("state")

	cs := &auth.CodeState{
		Code:  code,
		State: state,
	}

	// Get user info from LinuxDo
	userInfo, err := auth.GetLinuxDoUser(c.Request.Context(), cfg, cs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create or update user in database
	dbUser, err := database.CreateOrUpdateLinuxDoUser(
		userInfo.Id,
		userInfo.Username,
		userInfo.Name,
		userInfo.TrustLevel,
		userInfo.Active,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save user: " + err.Error()})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(
		dbUser.ID,
		dbUser.Username,
		dbUser.Name,
		dbUser.LinuxDoID,
		dbUser.TrustLevel,
		cfg.JWT.Secret,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  dbUser,
	})
}

// DonateRequest represents the request body for donating auth credentials
type DonateRequest struct {
	Auth   string `json:"auth" binding:"required"`
	TeamID string `json:"team_id" binding:"required"`
}

// DonateHandler handles account donation
// POST /api/donate
// Requires: Authorization header with Bearer token
func DonateHandler(c *gin.Context) {
	// Get claims from context (set by auth middleware)
	claims, ok := auth.GetClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req DonateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Validate auth and team_id are not empty
	if req.Auth == "" || req.TeamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "auth and team_id are required"})
		return
	}

	// Create the account with the user's linuxdo_id
	account, err := database.CreateAccount(req.Auth, req.TeamID, claims.LinuxDoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "donation successful",
		"account": gin.H{
			"id":         account.ID,
			"team_id":    account.TeamID,
			"linuxdo_id": account.LinuxdoID,
			"is_active":  account.IsActive,
			"created_at": account.CreatedAt,
		},
	})
}
