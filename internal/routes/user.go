package routes

import (
	"encoding/json"
	"net/http"

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
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		log.Error().Err(err).Msg("failed to push user to db")
		return
	}

	id, _ := createdUser.ID.Value()

	log.Info().Any("id", id).Msg("created new user")
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
