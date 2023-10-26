pipeline {

  agent any
  
    tools {
        go 'go1.21.3'
    }
            environment {
        GENESYSCLOUD_OAUTHCLIENT_ID = credentials('GENESYSCLOUD_OAUTHCLIENT_ID')
        GENESYSCLOUD_OAUTHCLIENT_SECRET = credentials('GENESYSCLOUD_OAUTHCLIENT_SECRET')
        GOPATH = "${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}"
		TF_ACC = "1"
        TF_LOG = "DEBUG"
        TF_LOG_PATH = "../test.log"
		GENESYSCLOUD_REGION = "us-east-1"
        GENESYSCLOUD_SDK_DEBUG =  "true"
        GENESYSCLOUD_TOKEN_POOL_SIZE =  20
      }
  stages {
      




       stage('Install Dependencies') {
            steps {
                echo 'Installing dependencies'
                sh 'go version'
                sh 'go mod download'
                sh 'go build -v .'
            }

       stage('Tests') {
            steps {
                echo 'Running Tests'
                sh ''

            }
    }
  }
}
