package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/jakottelaar/relay-backend/config"
	"github.com/jakottelaar/relay-backend/internal"
)

type UserHandler struct {
	service UserService
	cfg     config.Config
}

func NewUserHandler(service UserService, cfg config.Config) *UserHandler {
	return &UserHandler{
		service: service,
		cfg:     cfg,
	}
}

func (h *UserHandler) RegisterUser(c *gin.Context) {

	var req *RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validate := validator.New()

	if err := validate.Struct(req); err != nil {
		_ = c.Error(internal.NewUnprocessableEntityError("Invalid input: " + err.Error()))
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), &User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		_ = c.Error(err)
		return
	}

	token, err := internal.GenerateJWT(user.ID.String(), h.cfg.JwtSecret, h.cfg.JwtExpirationSecond)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": &RegisterResponse{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			AccessToken: token,
			CreatedAt:   user.CreatedAt,
		},
	})

}

func (h *UserHandler) Login(c *gin.Context) {

	var req *LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validate := validator.New()

	if err := validate.Struct(req); err != nil {
		_ = c.Error(internal.NewUnprocessableEntityError("Invalid input: " + err.Error()))
		return
	}

	resp, err := h.service.LoginUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": resp,
	})

}

func (h *UserHandler) GetProfile(c *gin.Context) {

	id, ok := c.Get("user_id")
	if !ok {
		_ = c.Error(internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	user, err := h.service.GetUserByID(c.Request.Context(), id.(string))

	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": &ProfileResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	})
}
