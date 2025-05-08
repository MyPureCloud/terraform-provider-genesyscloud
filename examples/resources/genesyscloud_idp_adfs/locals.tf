locals {
  adfs_certificate = utils_certificates.certificates.cert1

  dependencies = [
    "../../common/certificates.tf"
  ]
  working_dir = {
    idp_adfs = "."
  }

}
