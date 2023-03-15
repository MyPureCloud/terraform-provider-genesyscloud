pipeline {

  agent {
        node {
            label 'dev_v2'
        }
  }

  tools { go '1.20.2' }

   environment { 
      GPG_FINGERPINT= sh (returnStdout: true, script: 'echo aoeu').trim()
   }

  stages {
    // stage('Compile') {
    //   steps {
    //     sh 'make build'
    //   }
    // }

    // stage('Test') {
    //   steps {
    //     sh 'go test ./...'
    //   }
    // }

    stage ('Release') {
    //   when {
    //     buildingTag()
    //   }

      environment {
        GITHUB_TOKEN = credentials('MYPURECLOUD_GITHUB_TOKEN')
      }

      steps {
        withCredentials([file(credentialsId: 'Terraform_GPG', variable: 'terraform_gpg_private_key'),
                 file(credentialsId: 'Terraform_GPG', variable: 'terraform_gpg_private_key')]) {
                    sh "cp \$terraform_gpg_private_key . & chmod 755 secret.asc"
                    sh "./addCredToConfig.sh "
        }

        //sh './getgoreleaser.sh release --clean --release-notes=CHANGELOG.md --timeout 45m --parallelism 3'
      }
    }
  }
}