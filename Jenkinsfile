pipeline {

  agent {
        node {
            label 'dev_v2'
        }
  }

  stages {
    stage ('Release') {
      environment {
        GITHUB_TOKEN = credentials('MYPURECLOUD_GITHUB_TOKEN')
        GPG_FINGERPRINT="25D753B7C560659B057B714C970A8360B4BF5075"
      }

      steps {
         withCredentials([file(credentialsId: 'TERRAFORM_GPG', variable: 'terraform_gpg_private_key')]) {
                    sh "cp \$terraform_gpg_private_key /tmp/terraform_gpg_secret.asc & chmod 755 /tmp/terraform_gpg_secret.asc"
                    sh "./addCredToConfig.sh"
                    sh "rm -f secret.asc"
        }

        sh './getgoreleaser.sh release --clean --timeout 45m --parallelism 3'
      }
    }
  }
}
