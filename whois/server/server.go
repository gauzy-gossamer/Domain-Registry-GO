package server

import (
    "whois/cache"
    "whois/whois_resp"
)

type WhoisStorage interface {
    GetDomain(domainname string) (whois_resp.Domain, error)
}

type Server struct {
    RGconf RegConfig
    Cache cache.LRUCache[string, whois_resp.Domain]
    Storage WhoisStorage
}
