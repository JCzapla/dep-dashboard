package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	httpadapter "github.com/JCzapla/dep-dashboard/internal/adapter/inbound/http"
	depsdev "github.com/JCzapla/dep-dashboard/internal/adapter/outbound/depsdev"
	sqliteadapter "github.com/JCzapla/dep-dashboard/internal/adapter/outbound/sqlite"
	"github.com/JCzapla/dep-dashboard/internal/domain"
	"github.com/JCzapla/dep-dashboard/internal/service"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./deps.db?busy_timeout=5000")
	if err != nil {
		log.Fatalf("Open db error: %v", err)
	}
	defer db.Close()

	repo, err := sqliteadapter.NewRepository(db)
	if err != nil {
		log.Fatalf("Repository init error: %v", err)
	}

	client := depsdev.NewClient(&http.Client{Timeout: 10 * time.Second})
	service := service.NewDependencyService(repo,client)
	config := httpadapter.Config{
		DefaultPackage: domain.PackageRef{
			Name: "express",
			Version: "5.2.1",
		},
	}
	router := httpadapter.NewRouter(service, config)
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server Failed: %v", err)
	}
}