/*
go test -coverpkg=./... -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
*/
package tests

import (
    "testing"
    "registry/server"
)

var test_registrar_cert = "-----BEGIN CERTIFICATE----- MIIFLjCCAxYCFCZQKgIh9XiFqpuUrkqCNfQfl7hDMA0GCSqGSIb3DQEBCwUAMFQx CzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRl cm5ldCBXaWRnaXRzIFB0eSBMdGQxDTALBgNVBAMMBFJPT1QwH    hcNMjMwNDIyMTEx NDIwWhcNMjgwNDIwMTExNDIwWjBTMQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29t ZS1TdGF0ZTEhMB8GA1UECgwYSW50ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMQwwCgYD VQQDDANSRUcwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQC+ny6LPbh7 iRrqq+7KU    fMjuEEAHD7C+PnA1E9CL/p7OL54fKBjGK0n+FKPDWaBSgzdxPv9KwwU MHGu6gTRDtANcCFH6vFtJ8Z8ReKZxqHLK6Q485YeIEyGvCcZnI6twNFYYeRjrgdI DDTc4qI7JxQy1XzHMF2LAeP9JppNQV/TvroIrINcBMP67hZQxPswb5MmMUjk0oox Re5onf7+eBKwA0dfNiyUpvesGshnrOciks    iH92XO0cAkOSAOnM7Ef3x21CA0Xtbw WEhsbeUKhkciEnlfOxjtnr3H6bwX1mMT11gh3XkvqeqXXUxTUaBvwBK1tcZ7jbKv lBmc9OEX03NK5eSQ0IyEbwECyJ1CBkF9nDJGv+1QvsE7a7w5LnQeL+wNbikhx1Rf L1HwEOTkw6kC/QjFhFZlpZfEwv4vlgvKv/aWw/a13Ifw79U8anfWnoSsSgt    Znkny wx/3VAY2tvh0yJFwtYNjH1veFpYbhIoSsVJ+lNnfwrIHV2wrF+tF6b8OprsEOuW1 BcjpaC2Io7xUwtwgMnNTyKc2UnZ4B9M2IzaIvdwefYAiTfLpz7Gwi9HIFGadj+45 n2PXit42I18cZO9zY1Z1jRvZy7d/F+u8vsasIYc5qAKm0MWgSGe5AweNjT4gup47 DUtKXI13bXVMVz7FDZR    a/94PyoPfA8egxQIDAQABMA0GCSqGSIb3DQEBCwUAA4IC AQCQWmiyWamKT1f4bDkfI3yFyxQ2H6ubU+lS0hfaq49eicXFXxDbpGI8RRWRqaLu s/ouNwIzbc6xN7PP2z/qqnW/lUqe/hQs3XaG6a0dbT1V9wwYic8HyPfEBOYcI0lk 7wSYg3bKdiYk8iwETMQdF8uyyKba0pGLZljdFqsOacdJ    UA9pD/OTtz20dZJxCpQz vhzWnP71Ado5d7/PTuoIrRH7gRmjAFG60hwYKiWeQRujtYKXK5FNWZdDMqAgRXSv lvM4CaUel0QU/FJIcwanL6S2DpOcegUenc6sOY94rVPtwahjqRZ2mSBQZmp78LrW KWuQs2oQd0KWVNk9RZITRyJPpwmimR2VgVH4CcIc3VuiwcDevOz65fLchz4Kd+Jj FNP2    QnwbsIz1PPkxASC0a8kE1qiafLAnwSqx/c/6eb6ItUD2xSfRpa/BCwj10+ZY 7JKuJ5s49mEEOCHpv/v9oa8Yz4AnhPfVm09I2EYqp8rHv1LQkDplP2F9wiPnzhGt vaQUwswUf2rbsMHI9nt+4U+rKp2ZGeaKUA7OhCTzRow6ryilx/Yyf3TIm5McwYE7 qTeRNE5s8LY+pCAeFiZirAq6oaNPP    6nBOg/9HEjbbFinsX1pznbRyFlcxHkIlmSV ogocl6irWbZu2EkvRiJ0vNdSImPn0pFKy70VlHUaxrDmPQ== -----END CERTIFICATE-----"

func TestCert(t *testing.T) {
    cert := "-----BEGIN CERTIFICATE----- MIIFLjCCAxYCFCZQKgIh9XiFqpuUrkqCNfQfl7hDMA"
    reparsed_cert := server.ReparseCert(cert)
    if reparsed_cert != "" {
        t.Error("failed")
    }

    cert = test_registrar_cert
    reparsed_cert = server.ReparseCert(cert)
    if reparsed_cert == "" {
        t.Error("failed")
    }

    cert_decoded, err := server.DecodeCertificate(cert)
    if err != nil {
        t.Error("failed")
    }
    fingerprint := server.CalcMD5CertFingerprint(cert_decoded)
    if fingerprint != "A1:DD:46:43:35:51:EB:5F:42:8B:DF:A1:77:19:EA:DD" {
        t.Error("failed")
    }
}
