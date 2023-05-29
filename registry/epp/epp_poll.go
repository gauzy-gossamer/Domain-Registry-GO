package epp

import (
//    "registry/xml"
    "strconv"
    "registry/epp/dbreg"
    . "registry/epp/eppcom"
)

func epp_poll_req_impl(ctx *EPPContext) *EPPResult {
    ctx.logger.Info("Poll req")

    count, err :=  dbreg.GetPollMessageCount(ctx.dbconn, ctx.session.Regid)
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    if count == 0 {
        return &EPPResult{RetCode:EPP_POLL_NO_MSG}
    }
    poll_msg, err := dbreg.GetFirstUnreadPollMessage(ctx.dbconn, ctx.session.Regid)
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }
    poll_msg.Count = count

    var res = EPPResult{RetCode:EPP_POLL_ACK_MSG}
    res.Content = poll_msg
    return &res
}

func epp_poll_ack_impl(ctx *EPPContext, msgid string) *EPPResult {
    ctx.logger.Info("Poll ack ", msgid)

    msgid_, err := strconv.ParseUint(msgid, 10, 32)
    if err != nil {
        return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{"incorrect msgID"}}
    }

    err = dbreg.MarkMessageRead(ctx.dbconn, ctx.session.Regid, msgid_)
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    var res = EPPResult{RetCode:EPP_OK}
//    res.Content = host_data
    return &res
}
