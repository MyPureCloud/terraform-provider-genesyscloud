pipeline {
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
        sh 'curl -sfL https://goreleaser.com/static/run | bash'
      }
    }
  }
}