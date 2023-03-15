pipeline {

  agent {
        node {
            label 'dev_v2'
        }
  }

  tools { go '1.20.2' }

  stages {
    stage ('Release') {
      environment {
        GITHUB_TOKEN = credentials('MYPURECLOUD_GITHUB_TOKEN')
        GPG_FINGERPINT="93CCF015F4ECD0AAACFEA0349E486A1367C54A5E"
      }

      steps {
        withCredentials([file(credentialsId: 'Terraform_GPG', variable: 'terraform_gpg_private_key'),
                 file(credentialsId: 'Terraform_GPG', variable: 'terraform_gpg_private_key')]) {
                    sh "cp \$terraform_gpg_private_key . & chmod 755 secret.asc"
                    sh "./addCredToConfig.sh "
        }

        sh './getgoreleaser.sh release --clean --release-notes=CHANGELOG.md --timeout 45m --parallelism 3'
      }
    }
  }
}