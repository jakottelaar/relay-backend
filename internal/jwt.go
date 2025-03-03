package internal

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jakottelaar/relay-backend/config"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type JWTClaims struct {
	UserId string
	jwt.RegisteredClaims
}

type AuthPayload struct {
	AccessToken string
}

type AuthResponse struct {
	UserId  string
	Expired bool
}

func Authenticate(authPayload *AuthPayload, jwtSecret string) (*AuthResponse, error) {
	token, err := parseToken(authPayload.AccessToken, jwtSecret)

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

func GenerateJWT(userId string, jwtSecret string, jwtExpirationSecond int) (string, error) {
	expiresAt := time.Now().Add(time.Duration(jwtExpirationSecond) * time.Second)
	jwtClaims := &JWTClaims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	accessToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

func parseToken(accessToken string, jwtSecret string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(accessToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
}

func JWTAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		accessToken := ExtractTokenFromHeader(c.Request)
		if accessToken == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		authResult, err := Authenticate(&AuthPayload{
			AccessToken: accessToken,
		}, cfg.JwtSecret)

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

func ExtractTokenFromHeader(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}
