pipeline {
  agent {
        node {
            label 'dev_v2'
        }
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
        sh 'getgoreleaser.sh'
      }
    }
  }
}