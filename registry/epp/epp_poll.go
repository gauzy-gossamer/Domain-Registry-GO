package epp

import (
    "strconv"
    "registry/epp/dbreg"
    . "registry/epp/eppcom"
)

func get_poll_msg(ctx *EPPContext) (*PollMessage, error)  {
    var poll_msg *PollMessage

    ctx.logger.Trace(ctx.session.Regid)
    poll := dbreg.NewPollMsg(ctx.dbconn, ctx.session.Regid)
    count, err :=  poll.GetPollMessageCount()
    if err != nil {
        return poll_msg, err
    }

    if count == 0 {
        poll_msg = &PollMessage{}
        poll_msg.Count = count
        return poll_msg, nil
    }
    poll_msg, err = poll.SetExtended(true).GetFirstUnreadPollMessage()
    if err != nil {
        return poll_msg, err
    }
    poll_msg.Count = count

    return poll_msg, nil
}

func epp_poll_req_impl(ctx *EPPContext) *EPPResult {
    ctx.logger.Info("Poll req")

    poll_msg, err := get_poll_msg(ctx)
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    if poll_msg.Count == 0 {
        return &EPPResult{RetCode:EPP_POLL_NO_MSG}
    }

    var res = EPPResult{RetCode:EPP_POLL_ACK_MSG}
    res.Content = poll_msg.Content
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
    return &res
}
