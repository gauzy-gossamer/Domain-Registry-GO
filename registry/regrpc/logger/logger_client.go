package logger

import (
    "context"
    "time"
    "fmt"
    "registry/server"

    "google.golang.org/grpc"
)

const (
    // EPP
    SERVICE_ID =  3
)

type LoggerClient struct {
    conn *grpc.ClientConn
    logger server.Logger
}

func NewLoggerClient(host string, port int) *LoggerClient {
    client := &LoggerClient{}
    if host == "" {
        return client
    }

    host = fmt.Sprintf("%v:%d", host, port)
    client.Connect(host)
    return client
}

func (l *LoggerClient) Connect(host string) {
    tempConn, err := grpc.Dial(host, grpc.WithInsecure())
    if err != nil {
        l.logger.Error(err)
        return
    }

    l.conn = tempConn
}

func (l *LoggerClient) StartRequest(SourceIP string, RequestType uint32, SessionID uint64, UserID uint64) uint64 {
    if l.conn == nil {
        return 0
    }
    client := NewRegLoggerClient(l.conn)

    request := LogRequest{ServiceID:SERVICE_ID, SourceIP:SourceIP, RequestType:RequestType, SessionID:int64(SessionID)}

    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    ret, err := client.StartRequest(ctx, &request)

    if err != nil {
        l.logger.Error(err)
        return 0
    }

    return ret.LogID
}

func (l *LoggerClient) EndRequest(LogID uint64, ResponseCode uint32) {
    if l.conn == nil {
        return
    }
    client := NewRegLoggerClient(l.conn)

    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    _, err := client.EndRequest(ctx, &EndReq{LogID:LogID, ResponseCode:ResponseCode})

    if err != nil {
        l.logger.Error(err)
        return
    }
}
