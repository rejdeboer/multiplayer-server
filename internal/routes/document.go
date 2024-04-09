package routes

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
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
	OwnerID    string   `json:"owner_id"`
	SharedWith []string `json:"shared_with"`
	Content    []byte   `json:"content"`
}

type DocumentListItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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

	var uuid pgtype.UUID
	userID := ctx.Value("user_id").(string)
	err = uuid.Scan(userID)
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("failed to parse pg uuid")
		return
	}

	createdDocument, err := q.CreateDocument(ctx, db.CreateDocumentParams{
		Name:    document.Name,
		OwnerID: uuid,
	})
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("failed to push document to db")
		return
	}
	documentId, _ := createdDocument.ID.Value()
	log.Info().Any("document_id", documentId).Msg("created new document")

	ownerID, _ := createdDocument.OwnerID.Value()

	response, err := json.Marshal(DocumentResponse{
		ID:         documentId.(string),
		Name:       createdDocument.Name,
		OwnerID:    ownerID.(string),
		SharedWith: []string{},
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

	var uuid pgtype.UUID
	userID := ctx.Value("user_id").(string)
	err := uuid.Scan(userID)
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("failed to parse pg uuid")
		return
	}

	dbDocuments, err := q.GetDocumentsByOwnerID(ctx, uuid)
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("failed to fetch documents from db")
		return
	}

	var documents []DocumentListItem
	for _, document := range dbDocuments {
		documentId, _ := document.ID.Value()
		documents = append(documents, DocumentListItem{
			ID:   documentId.(string),
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

	docID := r.PathValue("id")
	userID := ctx.Value("user_id").(string)

	pool := ctx.Value("pool").(*pgxpool.Pool)
	q := db.New(pool)

	var docUUID pgtype.UUID
	err := docUUID.Scan(docID)
	if err != nil {
		httperrors.Write(w, "invalid document id, please use uuid format", http.StatusBadRequest)
		log.Error().Err(err).Msg("user used invalid document id format")
		return
	}

	var userUUID pgtype.UUID
	err = userUUID.Scan(userID)
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("failed to parse pg uuid")
		return
	}

	document, err := q.GetDocumentByID(ctx, db.GetDocumentByIDParams{
		ID:      docUUID,
		OwnerID: userUUID,
	})
	if err != nil {
		httperrors.Write(w, "document not found", http.StatusNotFound)
		log.Error().Err(err).Str("document_id", docID).Msg("document not found")
		return
	}

	response, err := json.Marshal(DocumentResponse{
		ID:         docID,
		OwnerID:    userID,
		Name:       document.Name,
		SharedWith: []string{},
		Content:    document.Content,
	})
	if err != nil {
		httperrors.InternalServerError(w)
		log.Error().Err(err).Msg("error marshalling response")
		return
	}

	log.Info().Str("document_id", docID).Msg("sending document")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
})
