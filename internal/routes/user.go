package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"unicode"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/db"
	"github.com/rejdeboer/multiplayer-server/pkg/httperrors"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type UserCreate struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func createUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := zerolog.Ctx(ctx)

	var user UserCreate
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		httperrors.Write(w, err.Error(), http.StatusBadRequest)
		log.Error().Err(err).Msg("invalid body for create user")
		return
	}

	err = validateUserCreate(user)
	if err != nil {
		httperrors.Write(w, err.Error(), http.StatusBadRequest)
		log.Error().Err(err).Msg("invalid payload")
		return
	}

	passhash, err := hashPassword(user.Password)
	if err != nil {
		httperrors.Write(w, err.Error(), http.StatusBadRequest)
		log.Error().Err(err).Msg("invalid password")
		return
	}

	pool := ctx.Value("pool").(*pgxpool.Pool)

	q := db.New(pool)

	createdUser, err := q.CreateUser(ctx, db.CreateUserParams{
		Email:    user.Email,
		Username: user.Username,
		Passhash: passhash,
	})
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("failed to push user to db")
		return
	}
	userId, _ := createdUser.ID.Value()
	log.Info().Any("user_id", userId).Msg("created new user")

	blob_client := ctx.Value("azblob").(*azblob.Client)
	_, err = blob_client.CreateContainer(ctx, userId.(string), nil)
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Str("user_id", userId.(string)).Msg("failed to create blob container")
		return
	}
	log.Info().Msg("created new blob container")

	response, err := json.Marshal(UserResponse{
		ID:       userId.(string),
		Email:    user.Email,
		Username: user.Username,
	})
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("error marshalling response")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func validateUserCreate(userCreate UserCreate) error {
	_, err := mail.ParseAddress(userCreate.Email)
	if err != nil {
		return errors.New("invalid email address")
	}

	err = validateUsername(userCreate.Username)
	if err != nil {
		return err
	}

	return validatePassword(userCreate.Password)
}

func validateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("username length must be at least 3 characters")
	}

	hasSpecial := false

	for _, char := range username {
		switch {
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if hasSpecial {
		return fmt.Errorf("username can not contain any special characters")
	}

	return nil
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
