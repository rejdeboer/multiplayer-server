package routes

import (
	"net/http"
	"slices"

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

	contributorID, err := uuid.Parse(r.PathValue("user-id"))
	if err != nil {
		httperrors.Write(w, "Invalid user id, please use uuid format", http.StatusBadRequest)
		log.Error().Err(err).Msg("user used invalid user id format")
		return
	}

	docID, err := uuid.Parse(r.PathValue("document-id"))
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

	document, err := q.GetDocumentByID(ctx, docID)
	if err != nil {
		httperrors.Write(w, "Document not found", http.StatusNotFound)
		log.Error().Err(err).Msg("document not found")
		return
	}

	if document.OwnerID != userID {
		httperrors.Write(w, "Document not found", http.StatusNotFound)
		log.Error().Err(err).
			Msg("user is not allowed to add contributor")
		return
	}

	if slices.Contains(document.SharedWith, contributorID) {
		httperrors.Write(w, "User is already a contributor", http.StatusBadRequest)
		log.Error().Err(err).Msg("user is already a contributor")
		return
	}

	err = q.AddDocumentContributor(ctx, db.AddDocumentContributorParams{
		ID:         docID,
		SharedWith: append(document.SharedWith, contributorID),
	})
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("error adding contributor")
	}

	log.Info().Msg("added contributor")
	w.WriteHeader(http.StatusAccepted)
})
