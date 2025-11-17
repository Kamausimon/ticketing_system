package auth

import (
	"net/http"

	"gorm.io/gorm"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type RegisterReponse struct {
	Message string `json:"message"`
	UserId  uint   `json:"user_id"`
	Email   string `json:"email"`
}

func (H *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {

}

func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	// Login logic
}

func (h *AuthHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	// Logout logic
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	// Password reset logic
}
