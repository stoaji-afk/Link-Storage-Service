package models

import "time"

type Link struct {
    ID         int       `json:"id"`           // уникальный идентификатор
    ShortCode  string    `json:"short_code"`   // короткий код ссылки
    OriginalURL string   `json:"original_url"` // исходный URL
    CreatedAt  time.Time `json:"created_at"`   // время создания
    Visits     int       `json:"visits"`       // количество переходов по ссылке
} 