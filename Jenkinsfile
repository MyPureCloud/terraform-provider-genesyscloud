@Library('pipeline-library') _
pipeline {
    agent {
        node{
        label "dev_mesos_large_v2"
        }
    }

    environment {
        CREDENTIALS_ID  = "GENESYSCLOUD_OAUTHCLIENT_ID_AND_SECERET"
        GOPATH = "$HOME/go"
        //PATH = "$PATH:$GOPATH/bin"
		GENESYSCLOUD_REGION = "us-east-1"
        GENESYSCLOUD_SDK_DEBUG =  "true"
        GENESYSCLOUD_TOKEN_POOL_SIZE =  20
        GO120MODULE= 'on'
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
                sh 'ls'
                sh 'go version'
                sh 'go mod download'
                sh 'go build -v .'
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
