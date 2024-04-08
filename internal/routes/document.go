package routes

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/db"
	"github.com/rejdeboer/multiplayer-server/pkg/httperrors"
	"github.com/rs/zerolog"
)

type DocumentCreate struct {
	Name string `json:"name"`
}

type DocumentResponse struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	SharedWith []string `json:"shared_with"`
}

func createDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := zerolog.Ctx(ctx)

	var document DocumentCreate
	err := json.NewDecoder(r.Body).Decode(&document)
	if err != nil {
		httperrors.Write(w, err.Error(), http.StatusBadRequest)
		log.Error().Err(err).Msg("invalid body for create document")
		return
	}

	pool := ctx.Value("pool").(*pgxpool.Pool)

	q := db.New(pool)

	createdDocument, err := q.CreateDocument(ctx, document.Name)
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("failed to push document to db")
		return
	}
	documentId, _ := createdDocument.ID.Value()
	log.Info().Any("document_id", documentId).Msg("created new document")

	response, err := json.Marshal(DocumentResponse{
		ID:         documentId.(string),
		Name:       createdDocument.Name,
		SharedWith: []string{},
	})
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("error marshalling response")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
