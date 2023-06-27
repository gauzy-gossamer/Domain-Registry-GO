package filestorage

import (
    "fmt"
    "logger/logrpc"
//    "github.com/jackc/pgtype"
)

func (st *FileStorage) StartRequest(ctx *logrpc.RequestContext, logreq *logrpc.LogRequest) (uint64, error) {
    logline := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\n", logreq.SourceIP, logreq.RequestType, logreq.SessionID, logreq.UserID, logreq.IsMonitoring)

    err := st.WriteLine(logline)
    if err != nil {
        return 0, err
    }

    var request_id uint64

    return request_id, nil
}

func (st *FileStorage) EndRequest(ctx *logrpc.RequestContext, request_id uint64, result_code_id uint32) error {
    logline := fmt.Sprintf("%v\t%v\n", result_code_id, request_id)

    err := st.WriteLine(logline)
    if err != nil {
        return err
    }

    return nil
}
