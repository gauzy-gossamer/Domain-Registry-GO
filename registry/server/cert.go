package server

import (
    "fmt"
    "strings"
    "errors"
    "crypto/md5"
    "crypto/x509"
    "net/http"
    "encoding/pem"
)

func CalcMD5CertFingerprint(cert *x509.Certificate) string {
    cert_digest := md5.Sum(cert.Raw)

    var buf string
    for i, f := range cert_digest {
        if i > 0 {
            buf += ":"
        }
        buf += fmt.Sprintf("%02X", f)
    }

    return buf
}

/*reparse escaped certificate into format that pem.Decode understands */
func ReparseCert(cert string) string {
    splitter := "-----"
    parts := strings.Split(cert, splitter)
    if len(parts) != 5 {
        return ""
    }
    return splitter + parts[1] + splitter +
           strings.Replace(parts[2], " ", "\n", -1) +
           splitter + parts[3] + splitter
}

func DecodeCertificate(cert_info string) (*x509.Certificate, error) {
    new_cert := ReparseCert(cert_info)
    cert_block, _ := pem.Decode([]byte(new_cert))
    if cert_block == nil {
        return nil, errors.New("could not parse certificate")
    }
    cert, err := x509.ParseCertificate(cert_block.Bytes)
    if err != nil {
        return nil, errors.New("could not parse certificate")
    }
    return cert, nil
}

/* either get from X-SSL-CERT or directly from Request */
func GetCertificateFingerprint(serv *Server, req *http.Request) (string, error) {
    if serv.RGconf.HTTPConf.UseProxy {
        cert_info := req.Header.Get("X-SSL-CERT")
        cert, err := DecodeCertificate(cert_info)
        if err != nil {
            return "", err
        }
        return CalcMD5CertFingerprint(cert), nil

    } else {
        certs := req.TLS.PeerCertificates
        if len(certs) == 0 {
            return "", errors.New("could not parse certificate")
        }
        cert_fingerprint := CalcMD5CertFingerprint(certs[0])
        return cert_fingerprint, nil
    }
}
