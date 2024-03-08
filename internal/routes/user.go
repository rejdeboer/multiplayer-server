package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"unicode"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/db"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type UserCreate struct {
	Email    string
	Password string
}

func createUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := zerolog.Ctx(ctx)

	var user UserCreate
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error().Err(err).Msg("invalid body for create user")
		return
	}

	err = validateUserCreate(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error().Err(err).Msg("invalid payload")
		return
	}

	passhash, err := hashPassword(user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error().Err(err).Msg("invalid password")
		return
	}

	pool := ctx.Value("pool").(*pgxpool.Pool)

	q := db.New(pool)

	createdUser, err := q.CreateUser(ctx, db.CreateUserParams{
		Email:    user.Email,
		Passhash: passhash,
	})
	if err != nil {
		internalServerError(w)
		log.Error().Err(err).Msg("failed to push user to db")
		return
	}

	id, _ := createdUser.ID.Value()

	log.Info().Any("id", id).Msg("created new user")
	w.WriteHeader(http.StatusOK)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func validateUserCreate(userCreate UserCreate) error {
	_, err := mail.ParseAddress(userCreate.Email)
	if err != nil {
		return err
	}

	return validatePassword(userCreate.Password)
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password length must be at least 8 characters")
	}

	hasLowerCase := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLowerCase = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasLowerCase {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}
