package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
	"github.com/kivaplus/kivaplus-backend/internal/users/domain"
)

// PersonDocumentPostgresRepository implements the PersonDocumentRepository interface using PostgreSQL
type PersonDocumentPostgresRepository struct {
	db *database.Connection
}

// NewPersonDocumentPostgresRepository creates a new PostgreSQL person document repository
func NewPersonDocumentPostgresRepository(db *database.Connection) *PersonDocumentPostgresRepository {
	return &PersonDocumentPostgresRepository{
		db: db,
	}
}

// Create creates a new person document in the database
func (r *PersonDocumentPostgresRepository) Create(ctx context.Context, document *domain.PersonDocument) error {
	query := `
		INSERT INTO person_document (person_document_id, type, num, issuing_authority, issue_date, expiration_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	err := r.db.DB.QueryRowContext(
		ctx,
		query,
		document.PersonID,
		document.Type,
		document.Number,
		document.IssuingAuthority,
		document.IssueDate,
		document.ExpirationDate,
		document.CreatedAt,
		document.UpdatedAt,
	).Scan(&document.ID)

	if err != nil {
		return fmt.Errorf("failed to create person document: %w", err)
	}

	return nil
}

// GetByPersonID retrieves all documents for a person
func (r *PersonDocumentPostgresRepository) GetByPersonID(ctx context.Context, personID int64) ([]*domain.PersonDocument, error) {
	query := `
		SELECT id, person_document_id, type, num, issuing_authority, issue_date, expiration_date, created_at, updated_at
		FROM person_document
		WHERE person_document_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.DB.QueryContext(ctx, query, personID)
	if err != nil {
		return nil, fmt.Errorf("failed to query person documents: %w", err)
	}
	defer rows.Close()

	var documents []*domain.PersonDocument
	for rows.Next() {
		var doc domain.PersonDocument
		var issuingAuthority sql.NullString
		var issueDate, expirationDate sql.NullTime

		err := rows.Scan(
			&doc.ID,
			&doc.PersonID,
			&doc.Type,
			&doc.Number,
			&issuingAuthority,
			&issueDate,
			&expirationDate,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan person document: %w", err)
		}

		if issuingAuthority.Valid {
			doc.IssuingAuthority = issuingAuthority.String
		}
		if issueDate.Valid {
			doc.IssueDate = &issueDate.Time
		}
		if expirationDate.Valid {
			doc.ExpirationDate = &expirationDate.Time
		}

		documents = append(documents, &doc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate person documents: %w", err)
	}

	return documents, nil
}

// GetByPersonIDAndType retrieves a specific document type for a person
func (r *PersonDocumentPostgresRepository) GetByPersonIDAndType(ctx context.Context, personID int64, docType domain.DocumentType) (*domain.PersonDocument, error) {
	query := `
		SELECT id, person_document_id, type, num, issuing_authority, issue_date, expiration_date, created_at, updated_at
		FROM person_document
		WHERE person_document_id = $1 AND type = $2`

	var doc domain.PersonDocument
	var issuingAuthority sql.NullString
	var issueDate, expirationDate sql.NullTime

	err := r.db.DB.QueryRowContext(ctx, query, personID, docType).Scan(
		&doc.ID,
		&doc.PersonID,
		&doc.Type,
		&doc.Number,
		&issuingAuthority,
		&issueDate,
		&expirationDate,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document not found")
		}
		return nil, fmt.Errorf("failed to get person document: %w", err)
	}

	if issuingAuthority.Valid {
		doc.IssuingAuthority = issuingAuthority.String
	}
	if issueDate.Valid {
		doc.IssueDate = &issueDate.Time
	}
	if expirationDate.Valid {
		doc.ExpirationDate = &expirationDate.Time
	}

	return &doc, nil
}

// Update updates an existing person document
func (r *PersonDocumentPostgresRepository) Update(ctx context.Context, document *domain.PersonDocument) error {
	query := `
		UPDATE person_document
		SET num = $1, issuing_authority = $2, issue_date = $3, expiration_date = $4, updated_at = $5
		WHERE id = $6`

	result, err := r.db.DB.ExecContext(
		ctx,
		query,
		document.Number,
		document.IssuingAuthority,
		document.IssueDate,
		document.ExpirationDate,
		document.UpdatedAt,
		document.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update person document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("person document not found")
	}

	return nil
}

// Delete deletes a person document
func (r *PersonDocumentPostgresRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM person_document WHERE id = $1`

	result, err := r.db.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete person document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("person document not found")
	}

	return nil
}

// ExistsByNumberAndType checks if a document with the same number and type already exists
func (r *PersonDocumentPostgresRepository) ExistsByNumberAndType(ctx context.Context, number string, docType domain.DocumentType) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM person_document WHERE num = $1 AND type = $2)`

	var exists bool
	err := r.db.DB.QueryRowContext(ctx, query, number, docType).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check document existence: %w", err)
	}

	return exists, nil
}

// ExistsByNumberAndTypeForDifferentPerson checks if a document exists for a different person
func (r *PersonDocumentPostgresRepository) ExistsByNumberAndTypeForDifferentPerson(ctx context.Context, number string, docType domain.DocumentType, excludePersonID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM person_document WHERE num = $1 AND type = $2 AND person_document_id != $3)`

	var exists bool
	err := r.db.DB.QueryRowContext(ctx, query, number, docType, excludePersonID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check document existence for different person: %w", err)
	}

	return exists, nil
}
