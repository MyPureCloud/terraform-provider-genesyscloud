pipeline {

  agent {
        node {
            label 'dev_v2'
        }
  }

  tools { go '1.19.5' }

  stages {
    stage ('Release') {
      environment {
        GITHUB_TOKEN = credentials('MYPURECLOUD_GITHUB_TOKEN')
        GPG_FINGERPINT="93CCF015F4ECD0AAACFEA0349E486A1367C54A5E"
      }

      steps {
        withCredentials([file(credentialsId: 'Terraform_GPG', variable: 'terraform_gpg_private_key'),
                 file(credentialsId: 'Terraform_GPG', variable: 'terraform_gpg_private_key')]) {
                    sh "cp \$terraform_gpg_private_key /tmp/terraform_gpg_secret.asc & chmod 755 /tmp/terraform_gpg_secret.asc"
                    sh "./addCredToConfig.sh"
                    sh "rm -f secret.asc"
        }

        sh './getgoreleaser.sh release --clean --release-notes=CHANGELOG.md --timeout 45m --parallelism 3'
      }
    }
  }
}