package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"service/internal/models"
	"service/internal/repository"
	"service/internal/config"
)

type LinkService struct {
	repo *repository.Repository
	cfg  *config.Config
}

func NewLinkService(repo *repository.Repository, cfg *config.Config) *LinkService {
	return &LinkService{repo: repo, cfg: cfg}
}

// generateShortCode генерирует случайный короткий код заданной длины
func (s *LinkService) generateShortCode() (string, error) {
	bytes := make([]byte, (s.cfg.ShortCodeLength+1)/2) // Округляем вверх для чётного числа байт
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes)[:s.cfg.ShortCodeLength], nil
}

// CreateShortLink создаёт короткую ссылку для переданного URL
func (s *LinkService) CreateShortLink(ctx context.Context, originalURL string) (string, error) {
	// Генерируем уникальный short_code
	var shortCode string
	var err error
	maxAttempts := 10

	for i := 0; i < maxAttempts; i++ {
		shortCode, err = s.generateShortCode()
		if err != nil {
			return "", err
	}

		// Проверяем, не существует ли уже такой код
		_, err = s.repo.GetLink(ctx, shortCode)
		if errors.Is(err, sql.ErrNoRows) {
			// Код уникален — можно использовать
			break
	} else if err != nil {
			return "", fmt.Errorf("error checking short code uniqueness: %w", err)
	}
		// Если код уже существует, продолжаем попытки
	}

	if shortCode == "" {
		return "", errors.New("failed to generate unique short code after maximum attempts")
	}

	// Сохраняем ссылку в хранилище
	if err := s.repo.CreateLink(ctx, originalURL, shortCode); err != nil {
		return "", fmt.Errorf("failed to create link in repository: %w", err)
	}

	return shortCode, nil
}

// GetOriginalURL получает оригинальную ссылку по short_code и увеличивает счётчик посещений
func (s *LinkService) GetOriginalURL(ctx context.Context, shortCode string) (*models.Link, error) {
	// Сначала получаем ссылку
	link, err := s.repo.GetLink(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	// Увеличиваем счётчик посещений
	if err := s.repo.IncrementVisits(ctx, shortCode); err != nil {
		return nil, fmt.Errorf("failed to increment visits: %w", err)
	}

	return link, nil
}

// ListLinks получает список ссылок с пагинацией
func (s *LinkService) ListLinks(ctx context.Context, limit, offset int) ([]*models.Link, error) {
	links, err := s.repo.ListLinks(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list links: %w", err)
	}
	return links, nil
}

// DeleteLink удаляет ссылку по short_code
func (s *LinkService) DeleteLink(ctx context.Context, shortCode string) error {
	err := s.repo.DeleteLink(ctx, shortCode)
	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}
	return nil
}

// GetLinkStats получает полную статистику по ссылке
func (s *LinkService) GetLinkStats(ctx context.Context, shortCode string) (*models.Link, error) {
	link, err := s.repo.GetLinkStats(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	return link, nil
}
