pipeline {

  agent any
  
    tools {
        go 'go1.21.3'
    }
            environment {
        GENESYSCLOUD_OAUTHCLIENT_ID = credentials('GENESYSCLOUD_OAUTHCLIENT_ID')
        GENESYSCLOUD_OAUTHCLIENT_SECRET = credentials('GENESYSCLOUD_OAUTHCLIENT_SECRET')
        GOPATH = "${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}"
      }
  stages {
      




       stage('Pre Test') {
            steps {
                echo 'Installing dependencies'
                sh 'go version'
                sh 'go mod download'
                sh 'go build -v .'
            }
    }
  }
}
