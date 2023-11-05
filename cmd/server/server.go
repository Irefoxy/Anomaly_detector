package main

import (
	t "Anomaly_detector/pkg/grpc/transmitter"
	"errors"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"time"
)

type customTransmitterServer struct {
	t.UnimplementedTransmitterServer
}

func (s *customTransmitterServer) ListRequests(_ *t.Empty, stream t.Transmitter_ListRequestsServer) error {
	rd := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	mean := -10.0 + rand.Float64()*20
	sd := 0.3 + rand.Float64()*(1.5-0.3)
	id := uuid.New().String()
	for {
		select {
		case <-stream.Context().Done():
			return errors.New("context done")
		default:
			response := &t.Response{
				Uuid:      id,
				Frequency: rd.NormFloat64()*sd + mean,
				Timestamp: time.Now().Unix(),
			}
			err := stream.Send(response)
			log.Println(response, mean, sd)
			if err != nil {
				return err
			}
			time.Sleep(250 * time.Millisecond)
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", "localhost:3333")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	t.RegisterTransmitterServer(grpcServer, &customTransmitterServer{})
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
