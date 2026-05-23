package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"server/internal/config"
	"server/internal/database"
	"server/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type server struct {
	pb.UnimplementedARPCollectorServer
	db database.Database
}

func (s *server) ARPStream(stream pb.ARPCollector_ARPStreamServer) error {
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

		fmt.Printf("--- New ARP Event (Agent: %s) ---\n", event.AgentId)
		fmt.Printf("  Timestamp: %d\n", event.Timestamp)
		fmt.Printf("  Opcode:    %d\n", event.Opcode)
		fmt.Printf("  Target IP: %v\n", net.IP(event.TargetIp).String())
		fmt.Printf("  Target MAC: %02x\n", event.TargetMac)
		fmt.Printf("  Sender IP: %v\n", net.IP(event.SenderId).String())
		fmt.Printf("  Sender MAC: %02x\n", event.SenderMac)
		fmt.Println("-----------------------------------")
	}
}

func getCredentials(cfg *config.Config) credentials.TransportCredentials {
	serverCert, err := tls.LoadX509KeyPair(cfg.Server.ServerPem, cfg.Server.ServerKey)
	if err != nil {
		log.Fatalf("Failed to load server certificate: %v", err)
	}

	caCert, err := os.ReadFile(cfg.Server.CaCert)
	if err != nil {
		log.Fatalf("Failed to load ca certificate: %v", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
	})

	return creds
}

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.InitDatabase(ctx, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	lis, err := net.Listen("tcp", cfg.Server.Port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	creds := getCredentials(cfg)

	s := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterARPCollectorServer(s, &server{db: db})

	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
