package handler

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"

	"waxp/echo/internal/apierror"
	"waxp/echo/internal/db"
	"waxp/echo/internal/middleware"
)

type AuthHandler struct {
	queries   *db.Queries
	jwtSecret string
}

func NewAuthHandler(queries *db.Queries, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		queries:   queries,
		jwtSecret: jwtSecret,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

func (h *AuthHandler) Register(c *echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return apierror.JSON(c, http.StatusBadRequest, "invalid request body")
	}

	if req.Email == "" || req.Password == "" {
		return apierror.JSON(c, http.StatusBadRequest, "email and password are required")
	}

	if err := validateEmail(req.Email); err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	if len(req.Password) < 8 {
		return apierror.JSON(c, http.StatusBadRequest, "password must be at least 8 characters")
	}

	if len(req.Password) > 72 {
		return apierror.JSON(c, http.StatusBadRequest, "password must be at most 72 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return apierror.Internal(c, "failed to hash password", err)
	}

	user, err := h.queries.CreateUser(c.Request().Context(), db.CreateUserParams{
		Email:        req.Email,
		PasswordHash: string(hash),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apierror.JSON(c, http.StatusConflict, "email already exists")
		}
		return apierror.Internal(c, "failed to create user", err)
	}

	token, err := middleware.GenerateToken(h.jwtSecret, user.ID, user.Email)
	if err != nil {
		return apierror.Internal(c, "failed to generate token", err)
	}

	return c.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User:  UserResponse{ID: user.ID, Email: user.Email},
	})
}

func (h *AuthHandler) Login(c *echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return apierror.JSON(c, http.StatusBadRequest, "invalid request body")
	}

	if req.Email == "" || req.Password == "" {
		return apierror.JSON(c, http.StatusBadRequest, "email and password are required")
	}

	if err := validateEmail(req.Email); err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	user, err := h.queries.GetUserByEmail(c.Request().Context(), req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusUnauthorized, "invalid credentials")
		}
		return apierror.Internal(c, "failed to get user", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return apierror.JSON(c, http.StatusUnauthorized, "invalid credentials")
	}

	token, err := middleware.GenerateToken(h.jwtSecret, user.ID, user.Email)
	if err != nil {
		return apierror.Internal(c, "failed to generate token", err)
	}

	return c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  UserResponse{ID: user.ID, Email: user.Email},
	})
}

func (h *AuthHandler) Me(c *echo.Context) error {
	userID, ok := c.Get("user_id").(int64)
	if !ok {
		return apierror.JSON(c, http.StatusUnauthorized, "unauthorized")
	}
	email, ok := c.Get("email").(string)
	if !ok {
		return apierror.JSON(c, http.StatusUnauthorized, "unauthorized")
	}

	return c.JSON(http.StatusOK, UserResponse{
		ID:    userID,
		Email: email,
	})
}
