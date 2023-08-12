package server

import (
    "whois/cache"
    "whois/whois_resp"
)

type WhoisStorage interface {
    GetDomain(domainname string) (whois_resp.Domain, error)
    GetRegistrar(registrar string) (whois_resp.Registrar, error)
}

type Server struct {
    RGconf RegConfig
    Cache cache.LRUCache[string, whois_resp.Domain]
    Storage WhoisStorage
}
