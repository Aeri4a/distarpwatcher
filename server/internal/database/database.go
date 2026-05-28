package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"server/pb"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BindingStatus string

const (
	BIND_TRUSTED  BindingStatus = "TRUSTED"
	BIND_CONFLICT BindingStatus = "CONFLICT"
)

type IPMACBinding struct {
	IPAddress  string
	MACAddress string
	LastSeen   time.Time
	Status     BindingStatus
}

type Database interface {
	SaveEvent(ctx context.Context, event *pb.ARPEvent) error

	// State Management for Analyzer
	GetIPMACBinding(ctx context.Context, ip string) (*IPMACBinding, error)
	CreateIPMACBinding(ctx context.Context, ip string, mac string, timestamp uint64) error
	UpdateLastSeen(ctx context.Context, ip string, timestamp uint64) error
	UpdateStatus(ctx context.Context, ip string, status BindingStatus) error
	UpdateMAC(ctx context.Context, ip string, mac string, timestamp uint64) error
	GetAgentsForMAC(ctx context.Context, mac string, window time.Duration) ([]string, error)

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

func (db *PostgresDatabase) GetIPMACBinding(ctx context.Context, ip string) (*IPMACBinding, error) {
	query := `
		SELECT ip_address::text, mac_address::text, last_seen, status 
		FROM ip_mac_bindings 
		WHERE ip_address = $1
	`

	var b IPMACBinding
	err := db.pool.QueryRow(ctx, query, ip).Scan(
		&b.IPAddress,
		&b.MACAddress,
		&b.LastSeen,
		&b.Status,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found is not an error
		}
		return nil, fmt.Errorf("failed to query ip_mac_bindings: %w", err)
	}

	return &b, nil
}

func (db *PostgresDatabase) CreateIPMACBinding(ctx context.Context, ip string, mac string, timestamp uint64) error {
	query := `
		INSERT INTO ip_mac_bindings (
			ip_address, mac_address, last_seen, status
		) VALUES (
			$1, $2, to_timestamp($3 / 1000.0), $4
		)
	`

	_, err := db.pool.Exec(ctx, query,
		ip,
		mac,
		timestamp,
		BIND_TRUSTED,
	)

	if err != nil {
		return fmt.Errorf("failed to create ip_mac_binding: %w", err)
	}

	return nil
}

func (db *PostgresDatabase) UpdateLastSeen(ctx context.Context, ip string, timestamp uint64) error {
	query := `
		UPDATE ip_mac_bindings 
		SET last_seen = to_timestamp($1 / 1000.0) 
		WHERE ip_address = $2
	`

	_, err := db.pool.Exec(ctx, query, timestamp, ip)
	if err != nil {
		return fmt.Errorf("failed to update last_seen: %w", err)
	}

	return nil
}

func (db *PostgresDatabase) UpdateStatus(ctx context.Context, ip string, status BindingStatus) error {
	query := `
		UPDATE ip_mac_bindings 
		SET status = $1 
		WHERE ip_address = $2
	`

	_, err := db.pool.Exec(ctx, query, status, ip)
	if err != nil {
		return fmt.Errorf("failed to update binding status: %w", err)
	}

	return nil
}

func (db *PostgresDatabase) UpdateMAC(ctx context.Context, ip string, mac string, timestamp uint64) error {
	query := `
		UPDATE ip_mac_bindings 
		SET mac_address = $1, last_seen = to_timestamp($2 / 1000.0), status = $3
		WHERE ip_address = $4
	`

	_, err := db.pool.Exec(ctx, query, mac, timestamp, BIND_TRUSTED, ip)
	if err != nil {
		return fmt.Errorf("failed to update mac address: %w", err)
	}

	return nil
}

func (db *PostgresDatabase) GetAgentsForMAC(ctx context.Context, mac string, window time.Duration) ([]string, error) {
	query := `
		SELECT array_agg(DISTINCT agent_id)
		FROM arp_events
		WHERE sender_mac = $1 AND captured_at > NOW() - $2::interval
	`

	intervalStr := fmt.Sprintf("%f seconds", window.Seconds())

	var agents []string
	err := db.pool.QueryRow(ctx, query, mac, intervalStr).Scan(&agents)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []string{}, nil
		}
		return []string{}, nil
	}

	return agents, nil
}
