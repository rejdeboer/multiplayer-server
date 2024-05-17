-- name: AddDocumentContributor :exec
UPDATE documents 
SET shared_with = $2
WHERE id = $1;
