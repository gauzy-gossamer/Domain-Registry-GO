package grpc

import (
    "context"
    "log"
    "net"
    "io"
    "testing"

    "registry/server"
    pb "registry/regrpc/cmd"
    "registry/tests/epptests"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/test/bufconn"
)

func ensureAllowDate(dbconn *server.DBConn) {
    _, _ = dbconn.Exec("insert into domain_allow_removal_dates(allow_date) values(current_date);")
}

func initServer(ctx context.Context) (pb.RegistryClient, *server.Server, func()) {
    buffer := 101024 * 1024
    lis := bufconn.Listen(buffer)

    serv := epptests.PrepareServer("../../server.conf")
    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        log.Printf("error acquiring conn: %v", err)
    }

    ensureAllowDate(dbconn)
    serv.Sessions.InitSessions(dbconn)

    baseServer := grpc.NewServer()
    pb.RegisterRegistryServer(baseServer, pb.NewServer(serv))
    go func() {
        if err := baseServer.Serve(lis); err != nil {
            log.Printf("error serving server: %v", err)
        }
    }()

    conn, err := grpc.DialContext(ctx, "",
        grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
            return lis.Dial()
        }), grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Printf("error connecting to server: %v", err)
    }

    closer := func() {
        err := lis.Close()
        if err != nil {
            log.Printf("error closing listener: %v", err)
        }
        baseServer.Stop()
    }

    client := pb.NewRegistryClient(conn)

    return client, serv, closer
}

func TestRegistryServer(t *testing.T) {
    ctx := context.Background()

    client, serv, closer := initServer(ctx)
    defer closer()

    session, err := client.LoginSystem(ctx, &pb.Empty{})
    if err != nil {
        t.Error("login failed")
    }

    epptests.CreateExpiredDomain(t, serv)

    stream, err := client.GetExpiredDomains(ctx, session)
    if err != nil {
        t.Error("get expired domain failed")
    }

    domain, err := stream.Recv()
    if err != nil {
        if err != io.EOF {
            t.Errorf("recv failed %v", err)
        }
    } else {
        _, err = client.DeleteDomain(ctx, &pb.Domain{Sessionid:session.Sessionid, Name:domain.Name})
        if err != nil {
            t.Errorf("delete domain failed %v", err)
        }
    }

    _, err = client.LogoutSystem(ctx, session)
    if err != nil {
        t.Error("logout failed")
    }
}
