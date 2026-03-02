package handler

import (
	"auth-service/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService domain.AuthService
}

func NewAuthHandler(as domain.AuthService) *AuthHandler {
	return &AuthHandler{authService: as}
}

// Struct untuk request body
type authRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type verifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Register handler
func (h *AuthHandler) Register(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input format"})
		return
	}

	if err := h.authService.Register(req.Email, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not register user"})
		return
	}

	// Setelah register, panggil SendOTP (opsional, bisa otomatis)
	_ = h.authService.SendOTP(req.Email)

	c.JSON(http.StatusCreated, gin.H{"message": "user registered, please check your email for OTP"})
}

// Login handler dengan metadata tracking
func (h *AuthHandler) Login(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	tokens, err := h.authService.Login(req.Email, req.Password, ua, ip)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// Realisasi VerifyOTP
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req verifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "otp and email are required"})
		return
	}

	if err := h.authService.VerifyOTP(req.Email, req.OTP); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "email verified successfully"})
}

// Realisasi RefreshToken (Rotation)
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}

	tokens, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	// userID diambil dari context yang diisi oleh AuthMiddleware
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	c.JSON(http.StatusOK, gin.H{
		"user_id":   userID,
		"user_role": userRole,
		"status":    "active",
	})
}
