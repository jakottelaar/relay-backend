package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jakottelaar/relay-backend/config"
)

type AuthService interface {
	Authenticate(accessToken string) (*AuthResponse, error)
	ExtractTokenFromHeader(header string) string
}

type AuthMiddlewareProvider interface {
	AuthMiddleware() gin.HandlerFunc
}

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type JWTClaims struct {
	UserId string `json:"sub"`
	jwt.RegisteredClaims
}

type AuthPayload struct {
	AccessToken string
}

type AuthResponse struct {
	UserId  string
	Expired bool
}

type SupabaseAuthService struct {
	JwtSecret string
}

func (s *SupabaseAuthService) Authenticate(accessToken string) (*AuthResponse, error) {
	token, err := s.parseToken(accessToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !(ok && token.Valid) {
		return nil, ErrInvalidToken
	}

	return &AuthResponse{
		UserId:  claims.UserId,
		Expired: false,
	}, nil
}

func (s *SupabaseAuthService) ExtractTokenFromHeader(header string) string {
	bearToken := header
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func (s *SupabaseAuthService) parseToken(accessToken string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(accessToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.JwtSecret), nil
	})
}

type SupabaseAuthMiddlewareProvider struct {
	Config      *config.Config
	AuthService AuthService
}

func (p *SupabaseAuthMiddlewareProvider) AuthMiddleware() gin.HandlerFunc {
	// If no auth service is provided, create one with the config
	if p.AuthService == nil {
		p.AuthService = &SupabaseAuthService{
			JwtSecret: p.Config.SupabaseJwtSecret,
		}
	}

	return func(c *gin.Context) {
		accessToken := p.AuthService.ExtractTokenFromHeader(c.Request.Header.Get("Authorization"))
		if accessToken == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		authResult, err := p.AuthService.Authenticate(accessToken)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if authResult.Expired {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "token expired",
			})
			return
		}

		c.Set("user_id", authResult.UserId)
		c.Next()
	}
}
