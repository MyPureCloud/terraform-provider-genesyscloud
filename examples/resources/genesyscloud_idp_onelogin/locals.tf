locals {
  onelogin_certificate = utils_certificates.certificates.cert1
  dependencies = {
    resource = [
      "../../common/certificates.tf"
    ]
  }
  working_dir = {
    idp_onelogin = "."
  }
}
