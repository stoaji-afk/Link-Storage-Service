package repository

import (
	"context"
	"database/sql"
	"fmt"
    "time"
	"service/internal/models"
	"service/internal/storage/cache"
	"service/internal/storage/db"
)

type Repository struct {
	db    *db.DB
	cache *cache.Cache
}

func New(db *db.DB, cache *cache.Cache) *Repository {
	return &Repository{db: db, cache: cache}
}

// CreateLink создаёт новую запись ссылки в базе данных
func (r *Repository) CreateLink(ctx context.Context, originalURL, shortCode string) error {
	query := `INSERT INTO links (short_code, original_url) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, shortCode, originalURL)
	return err
}

// GetLink получает ссылку по short_code — сначала проверяет кеш, затем БД
func (r *Repository) GetLink(ctx context.Context, shortCode string) (*models.Link, error) {
	// Проверка кеша
	if link, ok := r.cache.Get(shortCode); ok {
		return link, nil
	}

	query := `SELECT id, short_code, original_url, created_at, visits FROM links WHERE short_code = $1`
	var link models.Link

	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(
		&link.ID, &link.ShortCode, &link.OriginalURL, &link.CreatedAt, &link.Visits,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("link with short_code %s not found", shortCode)
	}
	return nil, fmt.Errorf("database error: %w", err)
	}

	// Сохраняем в кеш на 5 минут
	r.cache.Set(shortCode, &link, 5*time.Minute)

	return &link, nil
}

// IncrementVisits увеличивает счётчик посещений для указанной ссылки
func (r *Repository) IncrementVisits(ctx context.Context, shortCode string) error {
	query := `UPDATE links SET visits = visits + 1 WHERE short_code = $1`
	result, err := r.db.ExecContext(ctx, query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to increment visits: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("link with short_code %s not found", shortCode)
	}

	return nil
}

// ListLinks получает список ссылок с пагинацией (limit, offset)
func (r *Repository) ListLinks(ctx context.Context, limit, offset int) ([]*models.Link, error) {
	query := `
	SELECT id, short_code, original_url, created_at, visits
	FROM links
	LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list links: %w", err)
	}
	defer rows.Close()

	var links []*models.Link
	for rows.Next() {
		var link models.Link
		err := rows.Scan(&link.ID, &link.ShortCode, &link.OriginalURL, &link.CreatedAt, &link.Visits)
		if err != nil {
			return nil, fmt.Errorf("failed to scan link row: %w", err)
		}
		links = append(links, &link)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iteration: %w", err)
	}

	return links, nil
}

// DeleteLink удаляет ссылку по short_code и очищает кеш
func (r *Repository) DeleteLink(ctx context.Context, shortCode string) error {
	query := `DELETE FROM links WHERE short_code = $1`
	result, err := r.db.ExecContext(ctx, query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("link with short_code %s not found", shortCode)
	}

	// Удаляем из кеша
	r.cache.Delete(shortCode)

	return nil
}

// GetLinkStats получает полную информацию о ссылке (используется для статистики)
func (r *Repository) GetLinkStats(ctx context.Context, shortCode string) (*models.Link, error) {
	link, err := r.GetLink(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	return link, nil
}
