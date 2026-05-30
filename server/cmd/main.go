package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"server/internal/notifier"
	"server/pb"
	"syscall"
	"time"

	"server/internal/analyzer"
	"server/internal/api"
	"server/internal/collector"
	"server/internal/config"
	"server/internal/database"

	"golang.org/x/sync/errgroup"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbCtx, dbCancel := context.WithTimeout(ctx, 10*time.Second)
	defer dbCancel()

	db, err := database.InitDatabase(dbCtx, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	g, gCtx := errgroup.WithContext(ctx)
	eventChan := make(chan *pb.ARPEvent, 1024)
	notificationChan := make(chan *analyzer.Alert, 1024)

	grpcSrv := collector.NewGRPCServer(cfg.Server, db, eventChan)
	analyzerSrv := analyzer.NewAnalyzer(db, eventChan, notificationChan)
	notifierSrv := notifier.NewNotifier(db, notificationChan)
	apiSrv := api.NewAPIServer(cfg.API, db)

	g.Go(func() error {
		defer close(eventChan)
		return grpcSrv.Start(gCtx)
	})
	g.Go(func() error {
		return apiSrv.Start(gCtx)
	})
	g.Go(func() error {
		defer close(notificationChan)
		return analyzerSrv.Start(gCtx)
	})
	g.Go(func() error {
		return notifierSrv.Start(gCtx)
	})

	log.Println("Distributed ARP Watcher Server is running. Press Ctrl+C to stop.")

	if err := g.Wait(); err != nil && err != context.Canceled {
		log.Printf("Server exited with error: %v", err)
	} else {
		log.Println("Server shut down gracefully.")
	}
}
