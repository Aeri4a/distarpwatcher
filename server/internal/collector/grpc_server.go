package collector

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"server/internal/config"
	"server/internal/database"
	"server/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type CollectorGRPCServer struct {
	pb.UnimplementedARPCollectorServer
	db     database.Database
	config config.ServerConfig
}

func NewGRPCServer(cfg config.ServerConfig, db database.Database) *CollectorGRPCServer {
	return &CollectorGRPCServer{
		db:     db,
		config: cfg,
	}
}

func (s *CollectorGRPCServer) Start(ctx context.Context) error {
	creds, err := s.getCredentials()
	if err != nil {
		return fmt.Errorf("failed to load gRPC credentials: %w", err)
	}

	grpcServer := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterARPCollectorServer(grpcServer, s)

	lis, err := net.Listen("tcp", s.config.Port)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.config.Port, err)
	}

	errChan := make(chan error, 1)
	go func() {
		log.Printf("gRPC Collector listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop()
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

func (s *CollectorGRPCServer) ARPStream(stream pb.ARPCollector_ARPStreamServer) error {
	log.Println("New agent connected to ARPStream.")
	eventsReceived := uint32(0)

	for {
		event, err := stream.Recv()
		if err == io.EOF {
			log.Println("Agent disconnected stream.")
			return stream.SendAndClose(&pb.ARPEventResponse{
				EventsReceived: eventsReceived,
				Success:        true,
			})
		}
		if err != nil {
			log.Printf("Error receiving from stream: %v", err)
			return err
		}

		eventsReceived++

		err = s.db.SaveEvent(context.Background(), event)
		if err != nil {
			log.Printf("Error saving event to database: %v", err)
		}

		log.Printf("[Event] Agent: %s, IP: %s, MAC: %02x",
			event.AgentId,
			net.IP(event.TargetIp).String(),
			event.TargetMac)
	}
}

func (s *CollectorGRPCServer) getCredentials() (credentials.TransportCredentials, error) {
	serverCert, err := tls.LoadX509KeyPair(s.config.ServerPem, s.config.ServerKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load server keypair: %w", err)
	}

	caCert, err := os.ReadFile(s.config.CaCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA cert to pool")
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
	})

	return creds, nil
}
