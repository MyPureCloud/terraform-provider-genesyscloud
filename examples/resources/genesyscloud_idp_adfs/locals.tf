locals {
  adfs_certificate = tls_self_signed_cert.example.cert_pem

  dependencies = {
    resource = [
      "../../common/certificates.tf"
    ]
  }
  working_dir = {
    idp_adfs = "."
  }

}
