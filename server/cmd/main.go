package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"server/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type server struct {
	pb.UnimplementedARPCollectorServer
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

func getCredentials() credentials.TransportCredentials {
	// adjust path later
	serverCert, err := tls.LoadX509KeyPair("../certs/server.pem", "../certs/server.key")
	if err != nil {
		log.Fatalf("Failed to load server certificate: %v", err)
	}

	caCert, err := os.ReadFile("../certs/ca.pem")
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
	port := ":50051"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	creds := getCredentials()

	s := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterARPCollectorServer(s, &server{})

	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
