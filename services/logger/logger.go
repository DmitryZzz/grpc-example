package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"

	pbl "github.com/dmitryzzz/grpc-example/proto/logger"
)

var (
	port        = flag.Int("port", 8002, "The server port")
	kafkaServer = flag.String("kafkaserver", "localhost:9092", "The Kafka server")
	kafkaGroup  = flag.String("kafkagroup", "loggers", "The Kafka group id")
	CHAddr      = flag.String("chaddr", "127.0.0.1:9009", "The Clickhouse address")
	CHDB        = flag.String("chdb", "log", "The Clickhouse DB")
	CHUser      = flag.String("chuser", "default", "The Clickhouse user")
	CHPass      = flag.String("chpass", "", "The Clickhouse password")
)

func main() {
	flag.Parse()

	var wg sync.WaitGroup

	ch := &CH{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		ch.Connect(*CHAddr, *CHDB, *CHUser, *CHPass)
	}()

	kafka := NewKafkaConsumer(2)
	wg.Add(1)
	go func() {
		defer wg.Done()
		kafka.Connect(*kafkaServer, *kafkaGroup, ch, kafkaTopic)
	}()

	wg.Wait()

	if ch.err == nil {
		log.Println("[Logger] Connected to the Clickhouse")
		defer ch.Close()
	} else {
		log.Fatalf("ERROR: [Logger] Failed Clickhouse connection (%s): %s", *CHAddr, ch.err.Error())
	}

	if kafka.err == nil {
		log.Println("[Logger] Connected to the Kafka")
		defer kafka.Close()
		log.Print("[Logger] Starting consuming Kafka")
		go kafka.Listen()
	} else {
		log.Fatalf("ERROR: [Logger] Failed Kafka connection (%s): %s", *kafkaServer, kafka.err.Error())
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer lis.Close()
	grpcServer := grpc.NewServer()
	pbl.RegisterLoggerRepoServer(grpcServer, newServer(ch))
	log.Printf("gRPC logger server started %s ", lis.Addr().String())
	defer grpcServer.Stop()
	grpcServer.Serve(lis)

	// kafka, err := NewConsumer(*kafkaServer, *kafkaGroup, ch)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer kafka.Destroy()
	// log.Print("Start consuming Kafka: OK")
	// kafka.Subscribe(kafkaTopic)
}
