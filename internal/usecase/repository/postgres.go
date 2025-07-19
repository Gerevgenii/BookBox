package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/project/library/internal/entity"
)

var _ BooksRepository = (*postgresRepository)(nil)
var _ AuthorRepository = (*postgresRepository)(nil)

type postgresRepository struct {
	logger *zap.Logger
	db     *pgxpool.Pool
}

func (p *postgresRepository) CreateAuthor(ctx context.Context, author entity.Author) (resAuthor entity.Author, txErr error) {
	return myExtractCtx(ctx, p.db, func(tx pgx.Tx) (entity.Author, error) {
		const request = `INSERT INTO author (name) VALUES ($1) RETURNING id`

		result := entity.Author{
			Name: author.Name,
		}

		if err := tx.QueryRow(ctx, request, author.Name).Scan(&result.ID); err != nil {
			return entity.Author{}, changeError(err, entity.ErrAuthorNotFound)
		}

		return result, nil
	},
	)
}

func (p *postgresRepository) UpdateAuthor(ctx context.Context, author entity.Author) (txErr error) {
	return myExtractCtxNoT(ctx, p.db, func(tx pgx.Tx) error {
		const request = `UPDATE author SET name = $1 where id = $2`
		_, err := tx.Exec(ctx, request, author.Name, author.ID)
		return err
	})
}

func getBook(ctx context.Context, tx pgx.Tx, id string) (entity.Book, error) {
	const query = `SELECT b.id, b.name, array_remove(array_agg(ab.author_id), NULL) AS author_ids, b.created_at, b.updated_at FROM book b LEFT JOIN author_book ab ON b.id = ab.book_id WHERE b.id = $1 GROUP BY b.id;`
	var book entity.Book

	if err := tx.QueryRow(ctx, query, id).Scan(&book.ID, &book.Name, &book.AuthorIDs, &book.CreatedAt, &book.UpdatedAt); err != nil {
		return entity.Book{}, changeError(err, entity.ErrBookNotFound)
	}

	return book, nil
}

func (p *postgresRepository) GetAuthorBooks(ctx context.Context, authorID string) (resBooks []entity.Book, txErr error) {
	return myExtractCtx(ctx, p.db, func(tx pgx.Tx) ([]entity.Book, error) {
		const request = `SELECT book_id FROM author_book WHERE author_id = $1`
		rows, err := tx.Query(ctx, request, authorID)
		if err != nil {
			return []entity.Book{}, err
		}
		booksIdes, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (string, error) {
			var answer string
			if erro := row.Scan(&answer); erro != nil {
				return "", erro
			}
			return answer, nil
		})
		if err != nil {
			return []entity.Book{}, err
		}
		var ans = make([]entity.Book, 0, len(booksIdes))
		for _, id := range booksIdes {
			book, err := getBook(ctx, tx, id)
			if err != nil {
				return []entity.Book{}, err
			}
			ans = append(ans, book)
		}
		return ans, nil
	})
}

func changeError(err error, custom error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return custom
	}
	return err
}

func changeUnknownError(err error) error {
	const ErrForeignKeyViolation = "23503"

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == ErrForeignKeyViolation {
		return entity.ErrAuthorNotFound
	}

	return err
}

func (p *postgresRepository) GetAuthorInfo(ctx context.Context, authorID string) (resAuthor entity.Author, txErr error) {
	return myExtractCtx(ctx, p.db, func(tx pgx.Tx) (entity.Author, error) {
		const request = `SELECT id, name, created_at, updated_at FROM author WHERE id = $1;`
		var author entity.Author
		if err := tx.QueryRow(ctx, request, authorID).Scan(&author.ID, &author.Name, &author.CreatedAt, &author.UpdatedAt); err != nil {
			return entity.Author{}, changeError(err, entity.ErrAuthorNotFound)
		}
		return author, nil
	})
}

func (p *postgresRepository) UpdateBook(ctx context.Context, book entity.Book) (txErr error) {
	return myExtractCtxNoT(ctx, p.db, func(tx pgx.Tx) error {
		const request = `UPDATE book SET name = $1 WHERE id = $2`
		_, err := tx.Exec(ctx, request, book.Name, book.ID)
		if err != nil {
			return err
		}

		const newRequest = `DELETE FROM author_book WHERE book_id = $1`
		_, err = tx.Exec(ctx, newRequest, book.ID)
		if err != nil {
			return err
		}

		insertedRows := make([][]any, 0, len(book.AuthorIDs))

		for _, id := range book.AuthorIDs {
			insertedRows = append(insertedRows, []any{id, book.ID})
		}

		_, err = tx.CopyFrom(ctx, pgx.Identifier{"author_book"}, []string{"author_id", "book_id"}, pgx.CopyFromRows(insertedRows))
		if err != nil {
			return changeUnknownError(err)
		}

		return nil
	})
}

func NewPostgresRepository(logger *zap.Logger, db *pgxpool.Pool) *postgresRepository {
	return &postgresRepository{
		logger: logger,
		db:     db,
	}
}

func (p *postgresRepository) CreateBook(ctx context.Context, book entity.Book) (resBook entity.Book, txErr error) {
	return myExtractCtx(ctx, p.db, func(tx pgx.Tx) (entity.Book, error) {
		const queryBook = `
INSERT INTO book (name)
VALUES ($1)
RETURNING id, created_at, updated_at
`
		result := entity.Book{
			Name:      book.Name,
			AuthorIDs: book.AuthorIDs,
		}

		if err := tx.QueryRow(ctx, queryBook, book.Name).Scan(&result.ID, &result.CreatedAt, &result.UpdatedAt); err != nil {
			return entity.Book{}, changeError(err, entity.ErrBookNotFound)
		}

		const queryAuthorBooks = `
INSERT INTO author_book
(author_id, book_id)
VALUES ($1, $2)
`
		for _, authorID := range book.AuthorIDs {
			_, err := tx.Exec(ctx, queryAuthorBooks, authorID, result.ID)

			if err != nil {
				return entity.Book{}, changeUnknownError(err)
			}
		}

		return result, nil
	})
}

// GetBook
// incorrect
func (p *postgresRepository) GetBook(ctx context.Context, bookID string) (resBook entity.Book, txErr error) {
	return myExtractCtx(ctx, p.db, func(tx pgx.Tx) (entity.Book, error) {
		ans, err := getBook(ctx, tx, bookID)
		if err != nil {
			return entity.Book{}, err
		}
		return ans, nil
	})
}
