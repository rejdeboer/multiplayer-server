package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/db"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type UserCredentials struct {
	Email    string
	Password string
}

func getToken(signingKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		user, err := q.GetUserByEmail(ctx, credentials.Email)
		if err != nil {
			http.Error(w, "invalid email or password", http.StatusUnauthorized)
			log.Error().Err(err).Str("email", credentials.Email).Msg("user with email does not exist")
			return
		}

		if !checkPasswordHash(credentials.Password, user.Passhash) {
			http.Error(w, "invalid email or password", http.StatusUnauthorized)
			log.Error().Err(err).Msg("user entered wrong password")
			return
		}

		token, err := getJwt(signingKey, user.Email)
		if err != nil {
			internalServerError(w)
			log.Error().Err(err).Msg("error signing jwt")
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, token)
	}
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func getJwt(signingKey string, username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
