package main

import (
	"Anomaly_detector/internal/db"
	t "Anomaly_detector/pkg/grpc/transmitter"
	"context"
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
	"io"
	"log"
	"math"
	"sync"
)

func calcMean(elem float64, mean float64, count int) float64 {
	return (mean*float64(count-1) + elem) / float64(count)
}

func calcSd(mean float64, pool *sync.Pool) (float64, *sync.Pool) {
	sum := 0.0
	count := 0
	newPool := sync.Pool{}
	for elem := pool.Get(); elem != nil; elem = pool.Get() {
		sum += math.Pow(elem.(float64)-mean, 2.0)
		newPool.Put(elem)
		count++
	}
	return math.Sqrt(sum / float64(count)), &newPool
}

func detectAnomalies(stream t.Transmitter_ListRequestsClient, psql *gorm.DB, coefficient float64) error {
	mean := 0.0
	count := 0
	sd := 0.0
	anomalyCount := 0
	pool := new(sync.Pool)
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		count++
		if count < 150 {
			pool.Put(response.Frequency)
			mean = calcMean(response.Frequency, mean, count)
			sd, pool = calcSd(mean, pool)
			log.Printf("After %v responses mean is %v, sd is %v\n", count, mean, sd)
		}
		if count >= 150 {
			low := mean - (coefficient * sd)
			high := mean + (coefficient * sd)
			if !(response.Frequency >= low && response.Frequency <= high) {
				log.Printf("Anomaly detected in response %v; freq is %v; low %v, up %v\n", count, response.Frequency, low, high)
				anomalyCount++
				log.Println(anomalyCount, float64(anomalyCount)/float64(count)*100.0)
				psql.Create(&db.Record{
					Uuid:      response.Uuid,
					Frequency: response.Frequency,
					Timestamp: response.Timestamp,
				})
			}
		}
	}
}

func main() {
	coefficient := flag.Float64("k", 0.0, "Set an STD anomaly coefficient")
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial("localhost:3333", opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := t.NewTransmitterClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := client.ListRequests(ctx, &t.Empty{})
	if err != nil {
		log.Fatalf("error in client.ListRequests %v", err)
	}
	psql, err := db.Connect("./config/db.yaml")
	if err != nil {
		log.Fatalf("error in db connection %v", err)
	}
	if err = detectAnomalies(stream, psql, *coefficient); err != nil {
		log.Fatalf("error in recieving %v", err)
	}
}
