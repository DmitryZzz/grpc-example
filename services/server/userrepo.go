package main

import (
	"context"
	"fmt"
	"log"

	pbu "github.com/dmitryzzz/grpc-example/proto/user"
)

type userServer struct {
	pbu.UnimplementedUserRepoServer
	dbm   *DBManager
	redis *RedisManager
	p     *KafkaProducer
	ctx   context.Context
}

func newServer(db *DBManager, redis *RedisManager, p *KafkaProducer) *userServer {
	s := &userServer{
		dbm:   db,
		redis: redis,
		p:     p,
		ctx:   context.Background(),
	}
	return s
}

func (s *userServer) SaveUser(ctx context.Context, req *pbu.SaveUserRequest) (*pbu.SaveUserResponse, error) {
	user := req.GetData()
	email := user.GetEmail()
	name := user.GetName()
	log.Printf("Saving User: %v, %v", name, email)
	var id int64
	sqlStatement := `INSERT INTO users (email, name) VALUES ($1, $2) RETURNING id`
	err := s.dbm.db.QueryRow(sqlStatement, email, name).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("Failed user saving: %v", err)
	}
	log.Printf("Created User Id: %v", id)
	user.Id = id
	go s.p.Produce(fmt.Sprintf("Created user with id %d, emain %s, name %s", user.Id, user.Email, user.Name))
	return &pbu.SaveUserResponse{Status: pbu.Status_Success, Data: user}, nil
}

func (s *userServer) DeleteUser(ctx context.Context, req *pbu.DeleteUserRequest) (*pbu.DeleteUserResponse, error) {
	id := req.GetId()
	if id <= 0 {
		return &pbu.DeleteUserResponse{Status: pbu.Status_RequestError}, nil
	}
	log.Printf("Deleting User by id: %v", id)
	sqlStatement := `DELETE FROM users WHERE id = $1;`
	res, err := s.dbm.db.Exec(sqlStatement, id)
	if err != nil {
		return nil, err
	}
	c, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if c > 0 {
		return &pbu.DeleteUserResponse{Status: pbu.Status_Success}, nil
	}
	return &pbu.DeleteUserResponse{Status: pbu.Status_NothinToDo}, nil
}

func (s *userServer) GetUsers(req *pbu.GetUsersRequest, stream pbu.UserRepo_GetUsersServer) error {
	redisUsers, ok := s.redis.getUsers()
	if ok {
		log.Print("Retrieved users from Redis")
		for _, u := range *redisUsers {
			stream.Send(&pbu.User{Id: u.Id, Name: u.Name, Email: u.Email})
		}
		return nil
	}

	log.Print("Retrieved users from DB")
	rows, err := s.dbm.db.Query("SELECT id, email, name FROM users")
	if err != nil {
		return err
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	var users []StoredUser
	for rows.Next() {
		u := &pbu.User{}
		if err := rows.Scan(&u.Id, &u.Email, &u.Name); err != nil {
			return fmt.Errorf("Failed while scaning row: %s", err)
		}
		users = append(users, StoredUser{Id: u.Id, Email: u.Email, Name: u.Name})
		stream.Send(u)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("Failed getting users: %s", err)
	}

	log.Print("Storing users to Redis")
	s.redis.storeUsers(&users)

	return nil
}
