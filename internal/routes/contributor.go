package routes

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/db"
	"github.com/rejdeboer/multiplayer-server/pkg/httperrors"
	"github.com/rs/zerolog"
)

var addContributor = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := zerolog.Ctx(ctx)

	pool := ctx.Value("pool").(*pgxpool.Pool)
	q := db.New(pool)

	contributorID, err := uuid.Parse(r.PathValue("user_id"))
	if err != nil {
		httperrors.Write(w, "Invalid user id, please use uuid format", http.StatusBadRequest)
		log.Error().Err(err).Msg("user used invalid user id format")
		return
	}

	docID, err := uuid.Parse(r.PathValue("document_id"))
	if err != nil {
		httperrors.Write(w, "Invalid document id, please use uuid format", http.StatusBadRequest)
		log.Error().Err(err).Msg("user used invalid document id format")
		return
	}

	userID, err := uuid.Parse(ctx.Value("user_id").(string))
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("failed to parse uuid")
		return
	}

	*log = log.With().
		Str("document_id", docID.String()).
		Str("user_id", userID.String()).
		Str("contributor_id", contributorID.String()).
		Logger()

	_, err = getDocumentAsUser(ctx, docID, userID, q)
	if err != nil {
		log.Error().Err(err).Msg("error fetching document")
		httperrors.Write(w, "Document not found", http.StatusNotFound)
		return
	}

	err = q.CreateDocumentContributor(ctx, db.CreateDocumentContributorParams{
		DocumentID: docID,
		UserID:     contributorID,
	})
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("error adding contributor")
	}

	log.Info().Msg("added contributor")
	w.WriteHeader(http.StatusAccepted)
})
