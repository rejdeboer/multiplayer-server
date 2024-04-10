package routes

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/db"
	"github.com/rejdeboer/multiplayer-server/pkg/httperrors"
	"github.com/rs/zerolog"
)

type DocumentCreate struct {
	Name string `json:"name"`
}

type DocumentResponse struct {
	ID         uuid.UUID   `json:"id"`
	Name       string      `json:"name"`
	OwnerID    uuid.UUID   `json:"owner_id"`
	SharedWith []uuid.UUID `json:"shared_with"`
	Content    []byte      `json:"content"`
}

type DocumentListItem struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

var createDocument = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	userID, err := uuid.Parse(ctx.Value("user_id").(string))
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("failed to parse uuid")
		return
	}

	createdDocument, err := q.CreateDocument(ctx, db.CreateDocumentParams{
		Name:    document.Name,
		OwnerID: userID,
	})
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("failed to push document to db")
		return
	}
	log.Info().Str("document_id", createdDocument.ID.String()).Msg("created new document")

	response, err := json.Marshal(DocumentResponse{
		ID:         createdDocument.ID,
		Name:       createdDocument.Name,
		OwnerID:    createdDocument.OwnerID,
		SharedWith: createdDocument.SharedWith,
	})
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("error marshalling response")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
})

var listDocuments = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := zerolog.Ctx(ctx)

	pool := ctx.Value("pool").(*pgxpool.Pool)
	q := db.New(pool)

	userID, err := uuid.Parse(ctx.Value("user_id").(string))
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("failed to parse uuid")
		return
	}

	dbDocuments, err := q.GetDocumentsByOwnerID(ctx, userID)
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("failed to fetch documents from db")
		return
	}

	var documents []DocumentListItem
	for _, document := range dbDocuments {
		documents = append(documents, DocumentListItem{
			ID:   document.ID,
			Name: document.Name,
		})
	}

	response, err := json.Marshal(documents)
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("error marshalling response")
		return
	}

	log.Info().Int("items", len(documents)).Msg("sending document list")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
})

var getDocument = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := zerolog.Ctx(ctx)

	pool := ctx.Value("pool").(*pgxpool.Pool)
	q := db.New(pool)

	docID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httperrors.Write(w, "invalid document id, please use uuid format", http.StatusBadRequest)
		log.Error().Err(err).Msg("user used invalid document id format")
		return
	}

	userID, err := uuid.Parse(ctx.Value("user_id").(string))
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("failed to parse uuid")
		return
	}

	document, err := q.GetDocumentByID(ctx, db.GetDocumentByIDParams{
		ID:      docID,
		OwnerID: userID,
	})
	if err != nil {
		httperrors.Write(w, "document not found", http.StatusNotFound)
		log.Error().Err(err).Str("document_id", docID.String()).Msg("document not found")
		return
	}

	response, err := json.Marshal(DocumentResponse{
		ID:         docID,
		OwnerID:    userID,
		Name:       document.Name,
		SharedWith: document.SharedWith,
		Content:    document.Content,
	})
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("error marshalling response")
		return
	}

	log.Info().Str("document_id", docID.String()).Msg("sending document")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
})
