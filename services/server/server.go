package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	pbu "github.com/dmitryzzz/grpc-example/proto/user"
)

func main() {
	conf := NewConfig()
	var wg sync.WaitGroup

	db := &DBManager{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		db.Connect(&conf.psql)
	}()

	redis := &RedisManager{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		redis.Connect(&conf.redis)
	}()

	kafka := &KafkaProducer{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		kafka.Connect(&conf.kafka)
	}()

	wg.Wait()

	if db.err == nil {
		log.Printf("Connected to DB %q (%s:%d)\n", conf.psql.db, conf.psql.host, conf.psql.port)
		defer db.Close()
	} else {
		log.Fatalf("ERROR: Failed DB connection (%s:%d) : %s", conf.psql.host, conf.psql.port, db.err.Error())
	}

	if redis.err == nil {
		log.Printf("Connected to the Redis (%s)\n", conf.redis.addr)
		defer redis.Close()
	} else {
		log.Fatalf("ERROR: Failed Redis connection %q (%s) : %s", conf.redis.db, conf.redis.addr, redis.err.Error())
	}

	if kafka.err == nil {
		log.Printf("Connected to the Kafka (%s:%s)\n", conf.kafka.server, conf.kafka.topic)
		defer kafka.Close()
	} else {
		log.Fatalf("ERROR: Failed Kafka connection (%s:%s): %s", conf.kafka.server, conf.kafka.topic, kafka.err.Error())
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", conf.grpc.port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pbu.RegisterUserRepoServer(grpcServer, newServer(db, redis, kafka))
	log.Printf("gRPC server started %s", lis.Addr().String())
	grpcServer.Serve(lis)
}
