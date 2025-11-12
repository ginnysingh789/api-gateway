package handler

import (
	"context"
	"net/http"
	"time"

	"api-gateway/internal/config"
	"api-gateway/internal/models"
	"api-gateway/pkg/logger"
	"api-gateway/pkg/storage"
	"api-gateway/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	mongo  *storage.MongoClient
	config *config.Config
	logger *logger.Logger
}

func NewAuthHandler(mongo *storage.MongoClient, cfg *config.Config, log *logger.Logger) *AuthHandler {
	return &AuthHandler{
		mongo:  mongo,
		config: cfg,
		logger: log,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	collection := h.mongo.Database.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user already exists
	var existingUser models.User
	err := collection.FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"username": req.Username},
			{"email": req.Email},
		},
	}).Decode(&existingUser)

	if err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Username or email already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Errorw("Failed to hash password", "error", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Create user
	user := models.User{
		ID:        primitive.NewObjectID(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      "user",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		h.logger.Errorw("Failed to create user", "error", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate JWT token
	token, expiresAt, err := utils.GenerateToken(&user, h.config.JWT.Secret, h.config.JWT.Expiry)
	if err != nil {
		h.logger.Errorw("Failed to generate token", "error", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	h.logger.Infow("User registered successfully", "username", user.Username, "email", user.Email)

	utils.SuccessResponse(c, http.StatusCreated, "User registered successfully", gin.H{
		"token":      token,
		"expires_at": expiresAt,
		"user": models.UserResponse{
			ID:       user.ID.Hex(),
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	collection := h.mongo.Database.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find user
	var user models.User
	err := collection.FindOne(ctx, bson.M{"username": req.Username}).Decode(&user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Check if user is active
	if !user.Active {
		utils.ErrorResponse(c, http.StatusForbidden, "Account is inactive")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate JWT token
	token, expiresAt, err := utils.GenerateToken(&user, h.config.JWT.Secret, h.config.JWT.Expiry)
	if err != nil {
		h.logger.Errorw("Failed to generate token", "error", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	h.logger.Infow("User logged in successfully", "username", user.Username)

	utils.SuccessResponse(c, http.StatusOK, "Login successful", gin.H{
		"token":      token,
		"expires_at": expiresAt,
		"user": models.UserResponse{
			ID:       user.ID.Hex(),
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate old token
	claims, err := utils.ValidateToken(req.Token, h.config.JWT.Secret)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Get user from database
	collection := h.mongo.Database.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, _ := primitive.ObjectIDFromHex(claims.UserID)
	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found")
		return
	}

	if !user.Active {
		utils.ErrorResponse(c, http.StatusForbidden, "Account is inactive")
		return
	}

	// Generate new token
	newToken, expiresAt, err := utils.GenerateToken(&user, h.config.JWT.Secret, h.config.JWT.Expiry)
	if err != nil {
		h.logger.Errorw("Failed to generate token", "error", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Token refreshed successfully", gin.H{
		"token":      newToken,
		"expires_at": expiresAt,
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	objID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	collection := h.mongo.Database.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profile retrieved successfully", models.UserResponse{
		ID:       user.ID.Hex(),
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	})
}
