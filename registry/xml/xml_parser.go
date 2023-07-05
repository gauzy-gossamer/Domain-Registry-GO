package xml

import (
    "os"
    "fmt"
    "strings"
    "errors"
    "bufio"
    "encoding/json"

    . "registry/epp/eppcom"

    "github.com/kpango/glg"
    "github.com/lestrrat-go/libxml2"
    "github.com/lestrrat-go/libxml2/types"
    "github.com/lestrrat-go/libxml2/clib"
    "github.com/lestrrat-go/libxml2/xsd"
    "github.com/lestrrat-go/libxml2/xpath"
)

var (
    EPP_NS = "http://www.ripn.net/epp/ripn-epp-1.0"
    DOMAIN_NS = "http://www.ripn.net/epp/ripn-domain-1.0"
    CONTACT_NS = "http://www.ripn.net/epp/ripn-contact-1.0"
    HOST_NS = "http://www.ripn.net/epp/ripn-host-1.0"
    REGISTRAR_NS = "http://www.ripn.net/epp/ripn-registrar-1.0"
)

var namespaces = map[string]string{"epp":EPP_NS,
    "domain": DOMAIN_NS,
    "contact": CONTACT_NS,
    "host": HOST_NS,
    "registrar": REGISTRAR_NS,
}

type XMLParser struct {
    schema *xsd.Schema
}

type CommandError struct {
    RetCode int
    Msg string
}

func (e *CommandError) Error() string {
    return fmt.Sprintf("command error: %d; %s", e.RetCode, e.Msg)
}

/* set namespaces from config */
func (s *XMLParser) SetNamespaces(schema_ns string) error {
    if schema_ns == "" {
        return nil
    }
    type Namespaces struct {
        Epp string        `json:"epp"`
        EppCom string     `json:"eppcom"`
        Host string       `json:"host"`
        Domain string     `json:"domain"`
        Contact string    `json:"contact"`
        Registrar string  `json:"registrar"`
    }
    ns := Namespaces{}

    err := json.Unmarshal([]byte(schema_ns), &ns)
    if err != nil {
        return err
    }
    EPP_NS = ns.Epp
    DOMAIN_NS = ns.Domain
    HOST_NS = ns.Host
    CONTACT_NS = ns.Contact
    REGISTRAR_NS = ns.Registrar

    schemaLoc = ns.Epp + " ripn-epp-1.0.xsd"

    namespaces = map[string]string{"epp":EPP_NS,
        "domain": DOMAIN_NS,
        "contact": CONTACT_NS,
        "host": HOST_NS,
        "registrar": REGISTRAR_NS,
    }

    return nil
}

func (s *XMLParser) ReadSchema(schema_path string) {
    // change current directory so that libxml can access dependencies
    last_i := strings.LastIndex(schema_path, "/")
    filename := schema_path
    var cur_path string

    if last_i != -1 {
        change_path := schema_path[:last_i]
        filename = schema_path[last_i+1:]
        cur_path, _ = os.Getwd()
        err := os.Chdir(change_path)
        if err != nil {
            glg.Fatal(err)
        }
    }
    file, err := os.Open(filename)
    if err != nil {
        glg.Fatal(err)
    }
    defer func() {
        if err = file.Close(); err != nil {
            glg.Fatal(err)
        }
    }()

    var lines string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines += scanner.Text() + "\n"
    }

    schema, err := xsd.Parse([]byte(lines))
    if err != nil {
        glg.Fatal(err)
    }
    s.schema = schema

    if cur_path != "" {
        err = os.Chdir(cur_path)
        if err != nil {
            glg.Fatal(err)
        }
    }
}

/* get list of elements by xpath */
func getElementList(ctx *xpath.Context, xpath_val string) []string {
    var values []string
    nodes := xpath.NodeList(ctx.Find(xpath_val))
    for _,v := range nodes {
        if v.NodeType() != clib.ElementNode {
            continue
        }
        values = append(values, v.NodeValue())
    }
    return values
}

/* get list of statuses by xpath */
func getStatusList(ctx *xpath.Context, xpath_val string) []string {
    var values []string
    nodes := xpath.NodeList(ctx.Find(xpath_val))
    for _,v := range nodes {
        if v.NodeType() != clib.ElementNode {
            continue
        }
        node_ctx, err := xpath.NewContext(v)
        if err != nil {
            glg.Error(err)
            continue
        }
        status := xpath.String(node_ctx.Find("@s"))
        if status != "" {
            values = append(values, status)
        }
    }
    return values
}

func parseCheck(ctx *xpath.Context, node *types.Node) (*XMLCommand, error) {
    ctx.SetContextNode(*node)

    nodes := xpath.NodeList(ctx.Find("*"))
    if len(nodes) != 1 {
        return nil, errors.New("unknown command")
    }

    var cmd XMLCommand
    if err := ctx.SetContextNode(nodes[0]); err != nil {
        return nil, err
    }
    switch nodes[0].NodeName() {
        case "domain:check":
            var check_obj CheckObject

            check_obj.Names = getElementList(ctx, "domain:name")

            cmd.CmdType = EPP_CHECK_DOMAIN
            cmd.Content = &check_obj

            return &cmd, nil
        case "host:check":
            var check_obj CheckObject

            check_obj.Names = getElementList(ctx, "host:name")

            cmd.CmdType = EPP_CHECK_HOST
            cmd.Content = &check_obj

            return &cmd, nil
        case "contact:check":
            var check_obj CheckObject

            check_obj.Names = getElementList(ctx, "contact:id")

            cmd.CmdType = EPP_CHECK_CONTACT
            cmd.Content = &check_obj

            return &cmd, nil
        default:
            return nil, errors.New("unknown command")
    }
}

func parseInfo(ctx *xpath.Context, node *types.Node) (*XMLCommand, error) {
    ctx.SetContextNode(*node)

    nodes := xpath.NodeList(ctx.Find("*"))
    if len(nodes) != 1 {
        return nil, errors.New("unknown command")
    }

    var cmd XMLCommand
    if err := ctx.SetContextNode(nodes[0]); err != nil {
        return nil, err
    }
    switch nodes[0].NodeName() {
        case "domain:info":
            var info_domain InfoDomain

            name := xpath.String(ctx.Find("domain:name"))
            pwd := xpath.String(ctx.Find("domain:authInfo/domain:pw"))

            info_domain.Name = name
            info_domain.AuthInfo = pwd

            cmd.CmdType = EPP_INFO_DOMAIN
            cmd.Content = &info_domain

            return &cmd, nil
        case "host:info":
            var info_host InfoObject

            info_host.Name = xpath.String(ctx.Find("host:name"))

            cmd.CmdType = EPP_INFO_HOST
            cmd.Content = &info_host

            return &cmd, nil
        case "contact:info":
            var info_contact InfoObject

            info_contact.Name = xpath.String(ctx.Find("contact:id"))

            cmd.CmdType = EPP_INFO_CONTACT
            cmd.Content = &info_contact

            return &cmd, nil
        case "registrar:info":
            var info_registrar InfoObject

            info_registrar.Name = xpath.String(ctx.Find("registrar:id"))

            cmd.CmdType = EPP_INFO_REGISTRAR
            cmd.Content = &info_registrar

            return &cmd, nil
        default:
            return nil, errors.New("unknown command")
    }
}

func parseCreate(ctx *xpath.Context, node *types.Node) (*XMLCommand, error) {
    ctx.SetContextNode(*node)

    nodes := xpath.NodeList(ctx.Find("*"))
    if len(nodes) != 1 {
        return nil, errors.New("unknown command")
    }

    var cmd XMLCommand
    ctx.SetContextNode(nodes[0])
    switch nodes[0].NodeName() {
        case "domain:create":
            var create_domain CreateDomain

            name := xpath.String(ctx.Find("domain:name"))
            registrant := xpath.String(ctx.Find("domain:registrant"))

            create_domain.Name = name
            create_domain.Registrant = registrant

            create_domain.Description = getElementList(ctx, "domain:description")
            create_domain.Hosts = getElementList(ctx, "domain:ns/domain:hostObj")

            cmd.CmdType = EPP_CREATE_DOMAIN
            cmd.Content = &create_domain

            return &cmd, nil
        case "host:create":
            var create_host CreateHost

            name := xpath.String(ctx.Find("host:name"))

            create_host.Name = name

            create_host.Addr = getElementList(ctx, "host:addr")

            cmd.CmdType = EPP_CREATE_HOST
            cmd.Content = &create_host

            return &cmd, nil
        case "contact:create":
            var create_contact CreateContact

            name := xpath.String(ctx.Find("contact:id"))

            create_contact.Fields.ContactId = name
            parseContactFields(ctx, &create_contact.Fields)

            cmd.CmdType = EPP_CREATE_CONTACT
            cmd.Content = &create_contact

            return &cmd, nil
        default:
            return nil, errors.New("unknown command")
    }
}

func parseRenew(ctx *xpath.Context, node *types.Node) (*XMLCommand, error) {
    ctx.SetContextNode(*node)

    nodes := xpath.NodeList(ctx.Find("*"))
    if len(nodes) != 1 {
        return nil, errors.New("unknown command")
    }

    switch nodes[0].NodeName() {
        case "domain:renew":
            ctx.SetContextNode(nodes[0])

            var renew_domain RenewDomain
            var cmd XMLCommand

            name := xpath.String(ctx.Find("domain:name"))
            exp_date := xpath.String(ctx.Find("domain:curExpDate"))
            period := xpath.String(ctx.Find("domain:period"))

            renew_domain.Name = name
            renew_domain.CurExpDate = exp_date
            renew_domain.Period = period

            cmd.CmdType = EPP_RENEW_DOMAIN
            cmd.Content = &renew_domain

            return &cmd, nil
        default:
            return nil, errors.New("unknown command")
    }
}

func parseContactFields(ctx *xpath.Context, contact *ContactFields) {
    verified_nodes := len(xpath.NodeList(ctx.Find("contact:verified")))
    if verified_nodes > 0 {
        contact.Verified.Set(true)
    }
    unverified_nodes := len(xpath.NodeList(ctx.Find("contact:unverified")))
    if unverified_nodes > 0 {
        contact.Verified.Set(false)
    }

    nodes := xpath.NodeList(ctx.Find("contact:organization"))
    if len(nodes) == 0 {
        nodes = xpath.NodeList(ctx.Find("contact:person"))
        if len(nodes) == 0 {
            return
        }
        contact.ContactType = CONTACT_PERSON
    } else {
        contact.ContactType = CONTACT_ORG
    }
    ctx.SetContextNode(nodes[0])

    if contact.ContactType == CONTACT_ORG {
        contact.Fax = getElementList(ctx, "contact:fax")
        contact.LegalAddress = getElementList(ctx, "contact:legalInfo/contact:address")

        contact.IntPostal = xpath.String(ctx.Find("contact:intPostalInfo/contact:org"))
        contact.LocPostal = xpath.String(ctx.Find("contact:locPostalInfo/contact:org"))
        contact.TaxNumbers = xpath.String(ctx.Find("contact:taxpayerNumbers"))

    } else {
        contact.IntPostal = xpath.String(ctx.Find("contact:intPostalInfo/contact:name"))
        contact.LocPostal = xpath.String(ctx.Find("contact:locPostalInfo/contact:name"))

        contact.Birthday = xpath.String(ctx.Find("contact:birthday"))
    }
    contact.IntAddress = getElementList(ctx, "contact:intPostalInfo/contact:address")
    contact.LocAddress = getElementList(ctx, "contact:locPostalInfo/contact:address")

    contact.Emails = getElementList(ctx, "contact:email")
    contact.Voice = getElementList(ctx, "contact:voice")
}

func parseUpdate(ctx *xpath.Context, node *types.Node) (*XMLCommand, error) {
    ctx.SetContextNode(*node)

    nodes := xpath.NodeList(ctx.Find("*"))
    if len(nodes) != 1 {
        return nil, errors.New("unknown command")
    }

    var cmd XMLCommand
    ctx.SetContextNode(nodes[0])
    switch nodes[0].NodeName() {
        case "domain:update":
            var update_domain UpdateDomain

            update_domain.Name = xpath.String(ctx.Find("domain:name"))
            update_domain.Registrant = xpath.String(ctx.Find("domain:chg/domain:registrant"))
            update_domain.Description = getElementList(ctx, "domain:chg/domain:description")
            update_domain.AddHosts = getElementList(ctx, "domain:add/domain:ns/domain:hostObj")
            update_domain.RemHosts = getElementList(ctx, "domain:rem/domain:ns/domain:hostObj")

            update_domain.AddStatus = getStatusList(ctx, "domain:add/domain:status")
            update_domain.RemStatus = getStatusList(ctx, "domain:rem/domain:status")

            cmd.CmdType = EPP_UPDATE_DOMAIN
            cmd.Content = &update_domain

            return &cmd, nil
        case "host:update":
            var update_host UpdateHost

            update_host.Name = xpath.String(ctx.Find("host:name"))
            update_host.AddAddrs = getElementList(ctx, "host:add/host:addr")
            update_host.RemAddrs = getElementList(ctx, "host:rem/host:addr")

            update_host.AddStatus = getStatusList(ctx, "host:add/host:status")
            update_host.RemStatus = getStatusList(ctx, "host:rem/host:status")

            cmd.CmdType = EPP_UPDATE_HOST
            cmd.Content = &update_host

            return &cmd, nil
        case "contact:update":
            var update_contact UpdateContact

            update_contact.Fields.ContactId = xpath.String(ctx.Find("contact:id"))

            update_contact.AddStatus = getStatusList(ctx, "contact:add/contact:status")
            update_contact.RemStatus = getStatusList(ctx, "contact:rem/contact:status")

            nodes = xpath.NodeList(ctx.Find("contact:chg"))
            if len(nodes) > 0 {
                ctx.SetContextNode(nodes[0])
                parseContactFields(ctx, &update_contact.Fields)
            }

            cmd.CmdType = EPP_UPDATE_CONTACT
            cmd.Content = &update_contact

            return &cmd, nil
        case "registrar:update":
            var update_registrar UpdateRegistrar

            update_registrar.Name = xpath.String(ctx.Find("registrar:id"))

            update_registrar.AddAddrs = getElementList(ctx, "registrar:add/registrar:addr")
            update_registrar.RemAddrs = getElementList(ctx, "registrar:rem/registrar:addr")

            update_registrar.WWW = xpath.String(ctx.Find("registrar:chg/registrar:www"))
            update_registrar.Whois = xpath.String(ctx.Find("registrar:chg/registrar:whois"))

            cmd.CmdType = EPP_UPDATE_REGISTRAR
            cmd.Content = &update_registrar

            return &cmd, nil
        default:
            return nil, errors.New("unknown command")
    }
}

func deleteCmd(ctx *xpath.Context, obj string, CmdType int) (*XMLCommand, error) {
    var cmd XMLCommand

    var delete_obj DeleteObject

    field_name := ":name"
    if CmdType == EPP_DELETE_CONTACT {
        field_name = ":id"
    }

    name := xpath.String(ctx.Find(obj + field_name))

    delete_obj.Name = name

    cmd.CmdType = CmdType
    cmd.Content = &delete_obj

    return &cmd, nil
}

func parseDelete(ctx *xpath.Context, node *types.Node) (*XMLCommand, error) {
    ctx.SetContextNode(*node)

    nodes := xpath.NodeList(ctx.Find("*"))
    if len(nodes) != 1 {
        return nil, errors.New("unknown command")
    }

    ctx.SetContextNode(nodes[0])
    switch nodes[0].NodeName() {
        case "domain:delete":
            return deleteCmd(ctx, "domain", EPP_DELETE_DOMAIN)
        case "host:delete":
            return deleteCmd(ctx, "host", EPP_DELETE_HOST)
        case "contact:delete":
            return deleteCmd(ctx, "contact", EPP_DELETE_CONTACT)
        default:
            return nil, errors.New("unknown command")
    }
}

func parseTransfer(ctx *xpath.Context, node *types.Node) (*XMLCommand, error) {
    ctx.SetContextNode(*node)
    op := xpath.String(ctx.Find("@op"))

    var cmd XMLCommand
    var transfer TransferDomain

    op_code, ok := TransferOPMap[op]
    if !ok {
        return nil, &CommandError{RetCode:2500, Msg:"unsupported operation"}
    }

    transfer.OP = op_code

    transfer.Name = xpath.String(ctx.Find("domain:transfer/domain:name"))
    transfer.AcID = xpath.String(ctx.Find("domain:transfer/domain:acID"))
    transfer.ReID = xpath.String(ctx.Find("domain:transfer/domain:reID"))

    cmd.CmdType = EPP_TRANSFER_DOMAIN
    cmd.Content = &transfer

    return &cmd, nil
}

func parsePoll(ctx *xpath.Context, node *types.Node) (*XMLCommand, error) {
    ctx.SetContextNode(*node)
    op := xpath.String(ctx.Find("@op"))

    if op == "req" {
        return &XMLCommand{CmdType: EPP_POLL_REQ}, nil
    } else if op == "ack" {
        msgid := xpath.String(ctx.Find("@msgID"))
        return &XMLCommand{CmdType: EPP_POLL_ACK, Content:msgid}, nil
    }

    return nil, errors.New("unknown command")
}

func parseLogin(ctx *xpath.Context, node *types.Node) (*XMLCommand, error) {
    ctx.SetContextNode(*node)

    lang := xpath.String(ctx.Find("epp:options/epp:lang"))
    lang_code, ok := LanguageMap[lang]
    if !ok {
        return nil, &CommandError{RetCode:2500, Msg:"unsupported language"}
    }
    clid := xpath.String(ctx.Find("epp:clID"))
    pw := xpath.String(ctx.Find("epp:pw"))
    if clid == "" || pw == "" {
        return nil, &CommandError{RetCode:2500}
    }

    var epp_login = EPPLogin{PW:pw, Clid:clid, Lang:uint(lang_code)}
    var cmd = XMLCommand{CmdType:EPP_LOGIN, Content:&epp_login}

    return &cmd, nil
}

func parseCommand(ctx *xpath.Context, node *types.Node) (*XMLCommand, error) {
    ctx.SetContextNode(*node)

    nodes := xpath.NodeList(ctx.Find("epp:*[position()=1]"))
    if len(nodes) != 1 {
        glg.Error(len(nodes))
        return nil, errors.New("unknown command")
    }
    clTRID := xpath.String(ctx.Find("epp:clTRID"))

    var cmd *XMLCommand
    var err error
    switch nodes[0].NodeName() {
        case "check":
            cmd, err = parseCheck(ctx, &nodes[0])
        case "info":
            cmd, err = parseInfo(ctx, &nodes[0])
        case "create":
            cmd, err = parseCreate(ctx, &nodes[0])
        case "update":
            cmd, err = parseUpdate(ctx, &nodes[0])
        case "delete":
            cmd, err = parseDelete(ctx, &nodes[0])
        case "renew":
            cmd, err = parseRenew(ctx, &nodes[0])
        case "transfer":
            cmd, err = parseTransfer(ctx, &nodes[0])
        case "poll":
            cmd, err = parsePoll(ctx, &nodes[0])
        case "login":
            cmd, err = parseLogin(ctx, &nodes[0])
        case "logout":
            cmd = &XMLCommand{CmdType: EPP_LOGOUT}
        default:
            cmd = &XMLCommand{CmdType: EPP_UNKNOWN_ERR}
    }

    if err != nil {
        return nil, err
    }
    cmd.ClTRID = clTRID

    return cmd, err
}

func (s *XMLParser) ParseMessage(xml_message string) (*XMLCommand, error) {
    glg.Trace(xml_message)
    doc, err := libxml2.ParseString(xml_message)
    if err != nil {
        glg.Error(err)
        return nil, &CommandError{RetCode:EPP_SYNTAX_ERR, Msg:fmt.Sprint(err)}
    }
    defer doc.Free()

    valid, errs := s.validateMessage(&doc)
    if !valid {
        return nil, &CommandError{RetCode:EPP_SYNTAX_ERR, Msg:fmt.Sprint(errs)}
    }

    root ,err := doc.DocumentElement()
    if err != nil {
        glg.Error(err)
        return nil,err
    }
    ctx, err := xpath.NewContext(root)
    if err != nil {
        glg.Error(err)
        return nil,err
    }
    for v, k := range namespaces {
        ctx.RegisterNS(v, k)
    }
    nodes := xpath.NodeList(ctx.Find("/epp:epp/epp:*"))
    if len(nodes) != 1 {
        return nil, errors.New("incorrect command")
    }
    switch nodes[0].NodeName() {
        case "command":
            return parseCommand(ctx, &nodes[0])
        case "hello":
            var cmd XMLCommand
            cmd.CmdType = EPP_HELLO
            return &cmd, nil
        default:
            return nil, errors.New("unknown command")
    }
}

func (s *XMLParser) validateMessage(doc *types.Document) (bool, string) {
    if err := s.schema.Validate(*doc); err != nil {
        serr, _ := err.(xsd.SchemaValidationError)
        return_err := ""
        for _, e := range serr.Errors() {
            return_err += e.Error()
        }
        return false, return_err
    }

    return true, ""
}
