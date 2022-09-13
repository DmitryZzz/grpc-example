package main

import (
	"context"
	"flag"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pbl "github.com/dmitryzzz/grpc-example/proto/logger"
	pbu "github.com/dmitryzzz/grpc-example/proto/user"
)

var serverAddr = flag.String("serveraddr", "localhost:8001", "the address to connect to")
var loggerAddr = flag.String("loggeraddr", "localhost:8002", "the address to connect to")

func saveUser(c pbu.UserRepoClient, ctx context.Context, u *pbu.User) *pbu.User {
	log.Printf("Saving User: {%s, %s}\n", u.GetName(), u.GetEmail())
	res, err := c.SaveUser(ctx, &pbu.SaveUserRequest{Data: u})
	if err != nil {
		log.Fatalf("Could not save: %v", err)
	}
	log.Printf("Status: %s, ID: %d\n", res.GetStatus(), res.GetData().GetId())
	return res.GetData()
}

func deleteUser(c pbu.UserRepoClient, ctx context.Context, ID int64) {
	delRes, err := c.DeleteUser(ctx, &pbu.DeleteUserRequest{Id: ID})
	log.Printf("Deleting user with ID: %d", ID)
	if err != nil {
		log.Fatalf("Could not delete: %v", err)
	}
	log.Printf("Status: %s", delRes.GetStatus())
}

func getUsers(c pbu.UserRepoClient, ctx context.Context) {
	log.Printf("Getting users...")
	stream, err := c.GetUsers(ctx, &pbu.GetUsersRequest{})
	if err != nil {
		log.Fatalf("Could not get users: %v", err)
	}
	for {
		u, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Could not get users: %v", err)
		}
		log.Printf("User: %d, %s, %s", u.GetId(), u.GetName(), u.GetEmail())
	}
}

func getLogs(c pbl.LoggerRepoClient, ctx context.Context) {
	log.Printf("Getting log")
	logs, err := c.GetTail(ctx, &pbl.GetTailRequest{})
	if err != nil {
		log.Fatalf("Could not get log: %v", err)
	}
	for _, row := range logs.Logs {
		log.Printf("%s: %s", row.GetTs(), row.GetMsg())
	}
}


func main() {
	flag.Parse()
	time.Sleep(time.Second * 10)
	// Set up a connection to the server.
	sConn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer sConn.Close()
	c := pbu.NewUserRepoClient(sConn)
	log.Printf("Connected gRPC (user) %s", *serverAddr)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	u := &pbu.User{Name: "John", Email: "john@example.com"}
	u = saveUser(c, ctx, u)

	time.Sleep(time.Second)
	log.Print("------------------------------------------------")
	getUsers(c, ctx)
	// put User1 to cache

	// If last request was more than a minute ago,
	// we will recieve all users include last saved.
	time.Sleep(time.Second)
	log.Print("------------------------------------------------")
	
	u2 := &pbu.User{Name: "Barbara", Email: "barbara@example.com"}
	u2 = saveUser(c, ctx, u2)

	time.Sleep(time.Second)
	log.Print("------------------------------------------------")
	// Last user was created less then a minute ago,
	// so we will not recieve it
	getUsers(c, ctx)
	//less than 5 sec have passed => show only one User1 from Redis
	time.Sleep(time.Second*4)
	log.Print("------------------------------------------------")
	getUsers(c, ctx)
	//more than 5 sec have passed => show User1 & User2 from db

	// Test deleting used by gotten IDs
	deleteUser(c, ctx, u.Id)
	deleteUser(c, ctx, u2.Id)

	// Test deleting used by unknown ID
	deleteUser(c, ctx, 0)

	time.Sleep(time.Second)
	log.Print("--------")
	// Set up a connection to the logger.
	lConn, err := grpc.Dial(*loggerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer lConn.Close()
	l := pbl.NewLoggerRepoClient(lConn)
	log.Printf("Connected gRPC (logger) %s", *loggerAddr)
	// If last request was more than a minute ago,
	// we will recieve all users include last saved.
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()
	getLogs(l, ctx2)
}
