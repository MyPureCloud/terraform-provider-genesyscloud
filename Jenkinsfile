@Library('pipeline-library') _
pipeline {
    agent {
        node{
        label "dev_mesos_v2"
        customWorkspace "${JOB_NAME}-${currentBuild.number}"
        }
    }

    environment {
        CREDENTIALS_ID  = "GENESYSCLOUD_OAUTHCLIENT_ID_AND_SECERET"
        GOPATH = "$HOME/go"
		GENESYSCLOUD_REGION = "us-east-1"
        GENESYSCLOUD_SDK_DEBUG =  "true"
        GENESYSCLOUD_TOKEN_POOL_SIZE =  20
    }
    tools {
        go 'Go 1.20'
        terraform 'Terraform 1.0.10'
    }

    stages {
        
        stage('Load and Set Credentials') {
            steps {
                script{
                withCredentials([usernamePassword(credentialsId: CREDENTIALS_ID, usernameVariable: 'GENESYSCLOUD_OAUTHCLIENT_ID',passwordVariable:'GENESYSCLOUD_OAUTHCLIENT_SECRET')])
                {
                    echo 'Loading Genesys OAuth Credentials'
                }
                }
            }
        }
       stage('Install Dependencies') {
            steps {
                echo 'Installing dependencies'
                sh 'go version'
                sh 'go mod download'
                sh 'go build'
            }
	   }

       stage('Terraform Check') {
            steps {
                echo 'Check Terraform Installation'
                sh 'terraform -version'

            }
    }
        stage('Running Tests') {
            steps {
                echo 'Running Tests'
                sh 'go test -timeout 80m -v -cover ./genesyscloud/... -parallel 20 -coverprofile=coverage.out'

            }
    }


  }
}
