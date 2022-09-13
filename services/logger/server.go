package main

import (
	"context"

	pbl "github.com/dmitryzzz/grpc-example/proto/logger"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type loggerServer struct {
	pbl.UnimplementedLoggerRepoServer
	ch *CH
}

func (s *loggerServer) GetTail(ctx context.Context, req *pbl.GetTailRequest) (*pbl.Log, error) {
	rows, err := s.ch.Read()
	if err != nil {
		return nil, err
	}
	result := make([]*pbl.LogRow, len(*rows))
	for i, r := range *rows {
		result[i] = &pbl.LogRow{Ts: timestamppb.New(r.Dt), Msg: r.Msg}
	}

	return &pbl.Log{Logs: result}, nil
}

func newServer(ch *CH) *loggerServer {
	s := &loggerServer{
		ch: ch,
	}
	return s
}
