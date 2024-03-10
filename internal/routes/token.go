package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/db"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type UserCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

func getToken(signingKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := zerolog.Ctx(ctx)

		var credentials UserCredentials
		err := json.NewDecoder(r.Body).Decode(&credentials)
		if err != nil {
			writeError(w, err.Error(), http.StatusBadRequest)
			log.Error().Err(err).Msg("invalid payload")
			return
		}

		pool := ctx.Value("pool").(*pgxpool.Pool)
		q := db.New(pool)

		user, err := q.GetUserByEmail(ctx, credentials.Email)
		if err != nil {
			writeError(w, "invalid email or password", http.StatusUnauthorized)
			log.Error().Err(err).Str("email", credentials.Email).Msg("user with email does not exist")
			return
		}

		if !checkPasswordHash(credentials.Password, user.Passhash) {
			writeError(w, "invalid email or password", http.StatusUnauthorized)
			log.Error().Err(err).Msg("user entered wrong password")
			return
		}

		token, err := getJwt(signingKey, user.Email)
		if err != nil {
			internalServerError(w)
			log.Error().Err(err).Msg("error signing jwt")
			return
		}

		userId, _ := user.ID.Value()

		response, err := json.Marshal(TokenResponse{
			Token: token,
		})
		if err != nil {
			internalServerError(w)
			log.Error().Err(err).Msg("error marshalling response")
			return
		}

		log.Info().Any("user_id", userId).Msg("successful authentication")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
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
	claims["exp"] = time.Now().Add(time.Hour * 4).Unix()

	tokenString, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
