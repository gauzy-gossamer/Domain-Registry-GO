package main

import (
    "fmt"
    "context"
    "net"
    "logger/server"
    "logger/logrpc"
    "logger/logging"
    "google.golang.org/grpc"
    "github.com/kpango/glg"
)

type loggerServer struct {
    logrpc.UnimplementedRegistryServer
    mainServer *server.Server
}

func newServer(server *server.Server) *loggerServer {
    s := &loggerServer{}
    s.mainServer = server
    return s
}

func (r *loggerServer) StartRequest(ctx context.Context, logreq *logrpc.LogRequest) (*logrpc.LogID, error) {
    r_ctx := logrpc.RequestContext{Logger:logging.NewLogger("")}
    r_ctx.Logger.Trace("grpc StartRequest")
    server.Queries.Inc()

    ret_msg := logrpc.LogID{}

    logid, err := r.mainServer.Storage.(logrpc.StorageModule).StartRequest(&r_ctx, logreq)
    if err != nil {
        r_ctx.Logger.Error(err)
        return &ret_msg, err
    }

    ret_msg.LogID = logid

    return &ret_msg, nil
}

func (r *loggerServer) EndRequest(ctx context.Context, endreq *logrpc.EndReq) (*logrpc.Status, error) {
    r_ctx := logrpc.RequestContext{Logger:logging.NewLogger("")}
    r_ctx.Logger.Trace("grpc EndRequest")

    ret_msg := logrpc.Status{}

    err := r.mainServer.Storage.(logrpc.StorageModule).EndRequest(&r_ctx, endreq.LogID, endreq.RequestCode)
    if err != nil {
        r_ctx.Logger.Error(err)
        return nil, err
    }

    return &ret_msg, nil
}

func StartgRPCServer(serv *server.Server) {
    server_addr := fmt.Sprintf("%v:%d", serv.RGconf.GrpcHost, serv.RGconf.GrpcPort)
    lis, err := net.Listen("tcp", server_addr)
    if err != nil {
        glg.Fatal("failed to start gRPC:", err)
    }
    var opts []grpc.ServerOption
    grpcServer := grpc.NewServer(opts...)
    logrpc.RegisterRegistryServer(grpcServer, newServer(serv))
    glg.Info("running gRPC at ", server_addr)
    err = grpcServer.Serve(lis)
    if err != nil {
        glg.Fatal(err)
    }
}
