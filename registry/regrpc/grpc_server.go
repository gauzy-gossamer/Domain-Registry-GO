package regrpc

import (
    "fmt"
    "context"
    "net"
    "strconv"
    "registry/server"
    "registry/xml"
    "registry/epp"
    . "registry/epp/eppcom"
    "google.golang.org/grpc"
    "github.com/kpango/glg"
)

type registryServer struct {
    UnimplementedRegistryServer
    /* contains dbpool and config */
    mainServer *server.Server
}

func newServer(server *server.Server) *registryServer {
    s := &registryServer{}
    s.mainServer = server
    return s
}

func getSystemRegistrar(dbconn *server.DBConn) (*xml.EPPLogin, error) {
    login_cmd := xml.EPPLogin{Lang:LANG_EN}
    row := dbconn.QueryRow("SELECT handle, password, cert" +
                           " FROM registrar r INNER JOIN registraracl ra ON r.id=ra.id WHERE system")
    err := row.Scan(&login_cmd.Clid, &login_cmd.PW, &login_cmd.Fingerprint)
    if err != nil {
        return nil, err
    }

    return &login_cmd, nil
}

func (r *registryServer) LoginSystem(ctx context.Context, empty *Empty) (*Session, error) {
    glg.Trace("grpc LoginSystem")

    ret_msg := Session{}
    dbconn, err := server.AcquireConn(r.mainServer.Pool)
    if err != nil {
        glg.Error(err)
        return nil, err
    }
    defer dbconn.Close()
    login_cmd, err := getSystemRegistrar(dbconn)
    xml_cmd := xml.XMLCommand{SvTRID:"gRPCLogin", CmdType:EPP_LOGIN}
    xml_cmd.Content = login_cmd

    if err != nil {
        glg.Error(err)
        return nil, err
    }
    epp_result := epp.ExecuteEPPCommand(context.Background(), r.mainServer, &xml_cmd)
    if login_obj, ok := epp_result.Content.(*LoginResult); ok {
        ret_msg.Sessionid = strconv.FormatUint(login_obj.Sessionid,10)
    }

    return &ret_msg, nil
}

func (r *registryServer) LogoutSystem(ctx context.Context, session *Session) (*Status, error) {
    glg.Trace("grpc LogoutSystem")

    xml_cmd := xml.XMLCommand{SvTRID:"gRPCLogout", CmdType:EPP_LOGOUT}
    xml_cmd.Sessionid, _ = strconv.ParseUint(session.Sessionid, 10, 64)

    epp_result := epp.ExecuteEPPCommand(context.Background(), r.mainServer, &xml_cmd)

    ret_msg := Status{ReturnCode:0}

    if epp_result.RetCode != 1500 {
        ret_msg.ReturnCode = int32(epp_result.RetCode)
    }

    return &ret_msg, nil
}

func (r *registryServer) GetExpiredDomains(session *Session, stream Registry_GetExpiredDomainsServer) error {
    glg.Trace("grpc GetExpiredDomains")
    dbconn, err := server.AcquireConn(r.mainServer.Pool)
    if err != nil {
        glg.Error(err)
        return err
    }
    defer dbconn.Close()
    /* 0 - all objects */
    if err := epp.UpdateObjectStates(dbconn, uint64(0)) ; err != nil {
        return err
    }
    /* state_id = 17 - pendingDelete */
    expired_query := "SELECT o.name " +
                     "FROM object_state s JOIN object_registry o ON ( " +
                     " o.erdate ISNULL AND o.id=s.object_id " +
                     " AND s.state_id=17 AND s.valid_to ISNULL) " +
                     "LEFT JOIN domain d ON (d.id=o.id) " +
                     "WHERE o.type = 3"
    rows, err := dbconn.Query(expired_query)
    if err != nil {
        glg.Error(err)
        return err
    }
    defer rows.Close()
    for rows.Next() {
        var domain string
        err := rows.Scan(&domain)
        if err != nil {
            glg.Error(err)
            return err
        }
        domain_ret := Domain{Name:domain}
        if err := stream.Send(&domain_ret); err != nil {
            glg.Error(err)
            return err
        }
    }

    return nil
}

func (r *registryServer) DeleteDomain(ctx context.Context, domain *Domain) (*Status, error) {
    glg.Trace("grpc DeleteDomain", domain.Name)

    xml_cmd := xml.XMLCommand{SvTRID:"gRPCDelete", CmdType:EPP_DELETE_DOMAIN}
    xml_cmd.Sessionid, _ = strconv.ParseUint(domain.Sessionid, 10, 64)
    xml_cmd.Content = &xml.DeleteObject{Name:domain.Name}

    epp_result := epp.ExecuteEPPCommand(context.Background(), r.mainServer, &xml_cmd)

    ret_msg := Status{ReturnCode:0}

    if epp_result.RetCode != 1500 {
        ret_msg.ReturnCode = int32(epp_result.RetCode)
    }

    return &ret_msg, nil
}

func StartgRPCServer(serv *server.Server) {
    port := serv.RGconf.GrpcPort
    server_addr := fmt.Sprintf("localhost:%d", port)
    lis, err := net.Listen("tcp", server_addr)
    if err != nil {
        glg.Fatal("failed to start gRPC:", err)
    }
    var opts []grpc.ServerOption
    grpcServer := grpc.NewServer(opts...)
    RegisterRegistryServer(grpcServer, newServer(serv))
    glg.Info("running gRPC at ", server_addr)
    err = grpcServer.Serve(lis)
    if err != nil {
        glg.Fatal(err)
    }
}
