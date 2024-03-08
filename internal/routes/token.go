package routes

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/db"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type UserCredentials struct {
	email    string
	password string
}

func getToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := zerolog.Ctx(ctx)

	var credentials UserCredentials
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error().Err(err).Msg("invalid payload")
		return
	}

	pool := ctx.Value("pool").(*pgxpool.Pool)
	q := db.New(pool)

	user, err := q.GetUserByEmail(ctx, credentials.email)
	if err != nil {
		http.Error(w, "invalid email or password", http.StatusBadRequest)
		log.Error().Err(err).Str("email", credentials.email).Msg("user with email does not exist")
		return
	}

	if !checkPasswordHash(credentials.password, user.Passhash) {
		http.Error(w, "invalid email or password", http.StatusBadRequest)
		log.Error().Err(err).Msg("user entered wrong password")
		return
	}
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
