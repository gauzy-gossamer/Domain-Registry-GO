package server

import (
    "registry/xml"
    "math/rand"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
    RGconf RegConfig
    Xml_parser xml.XMLParser
    Sessions EPPSessions
    Pool *pgxpool.Pool
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateRandString(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Int63() % int64(len(letterBytes))]
    }
    return string(b)
}

func GenerateTRID(n int) string {
    return "SV-" + GenerateRandString(n)
}

