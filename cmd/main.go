package main

import (
    "log"
    "net/http"
    "service/internal/config"
    "service/internal/handlers"
    "service/internal/repository"
	"service/internal/service"
    "service/internal/storage/cache"
    "service/internal/storage/db"
)

func main() {
    // Загрузка конфигурации
    cfg := config.Load()

    // Инициализация БД
    db, err := db.New(cfg.DBURL)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()

    // Инициализация схемы БД
    if err := db.Init(); err != nil {
        log.Fatal("Failed to initialize database schema:", err)
    }

    // Инициализация кеша
    cache := cache.New(cfg.CacheSize)

    // Создание репозитория
    repo := repository.New(db, cache)

    // Создание сервисного слоя
    linkService := service.NewLinkService(repo, cfg)

    // Создание обработчиков
    handlers := handlers.NewHandlers(linkService)

    // Настройка маршрутов
    http.HandleFunc("POST /links", handlers.CreateLinkHandler)
    http.HandleFunc("GET /links/", handlers.GetOriginalURLHandler)
    http.HandleFunc("GET /links", handlers.ListLinksHandler)
    http.HandleFunc("DELETE /links/", handlers.DeleteLinkHandler)
    http.HandleFunc("GET /links/{short_code}/stats", handlers.GetLinkStatsHandler)

    log.Printf("Server starting on port %s", cfg.Port)
    log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
