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
        GPG_FINGERPRINT="09AD081CB36489681F3C4136FC6E114B953764EE"
      }

      steps {
        withCredentials([file(credentialsId: 'TERRAFORM_GPG', variable: 'terraform_gpg_private_key'),
                 file(credentialsId: 'TERRAFORM_GPG', variable: 'terraform_gpg_private_key')]) {
                    sh "cp \$terraform_gpg_private_key /tmp/terraform_gpg_secret.asc & chmod 755 /tmp/terraform_gpg_secret.asc"
                    sh "./addCredToConfig.sh"
                    sh "rm -f secret.asc"
        }

        sh './getgoreleaser.sh release --clean --release-notes=CHANGELOG.md --timeout 45m --parallelism 3'
      }
    }
  }
}