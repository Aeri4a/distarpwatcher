package database

import (
	"context"
	"fmt"
	"log"
	"net"

	"server/pb"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database interface {
	SaveEvent(ctx context.Context, event *pb.ARPEvent) error
	Close()
}

type PostgresDatabase struct {
	pool *pgxpool.Pool
}

func InitDatabase(ctx context.Context, dsn string) (*PostgresDatabase, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection pool established.")

	if err := RunMigrations(dsn); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &PostgresDatabase{
		pool: pool,
	}, nil
}

func (db *PostgresDatabase) Close() {
	if db.pool != nil {
		db.pool.Close()
		log.Println("Database connection pool closed.")
	}
}

func (db *PostgresDatabase) SaveEvent(ctx context.Context, event *pb.ARPEvent) error {
	query := `
		INSERT INTO arp_events (
			agent_id, captured_at, opcode, target_ip, target_mac, sender_ip, sender_mac
		) VALUES (
			$1, to_timestamp($2 / 1000.0), $3, $4, $5, $6, $7
		)
	`

	targetIP := net.IP(event.TargetIp).String()
	targetMAC := net.HardwareAddr(event.TargetMac).String()
	senderIP := net.IP(event.SenderId).String()
	senderMAC := net.HardwareAddr(event.SenderMac).String()

	_, err := db.pool.Exec(ctx, query,
		event.AgentId,
		event.Timestamp, // Milliseconds
		event.Opcode,
		targetIP,
		targetMAC,
		senderIP,
		senderMAC,
	)

	if err != nil {
		return fmt.Errorf("failed to insert arp event: %w", err)
	}

	return nil
}
