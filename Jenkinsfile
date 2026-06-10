pipeline {
    agent { label 'billing-agent' }

    environment {
        SKIP_DATA_IMPORT = 'true'
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Environment Setup') {
            steps {
                sh '''
                    cp backend/.env.example backend/.env

                    # Generate valid Fernet ENCRYPTION_KEY
                    KEY=$(python3 -c "from cryptography.fernet import Fernet; print(Fernet.generate_key().decode())" 2>/dev/null || \\
                          dd if=/dev/urandom bs=1 count=32 2>/dev/null | base64 | tr '+/' '-_' | tr -d '=\\n')
                    if [ -n "$KEY" ]; then
                        sed -i "s|ENCRYPTION_KEY=\".*\"|ENCRYPTION_KEY=\"$KEY\"|" backend/.env
                        echo "ENCRYPTION_KEY generated successfully"
                    fi
                '''
            }
        }

        stage('Build & Deploy') {
            steps {
                sh '''
                    # Build images and start containers
                    docker compose up -d --build
                '''
            }
        }
    }

    post {
        always {
            sh 'docker image prune -f'
        }
        success {
            echo 'CI/CD Pipeline berhasil!'
        }
        failure {
            echo 'CI/CD Pipeline gagal!'
            sh 'docker compose logs --tail 20 mysql-master backend1 backend2'
        }
    }
}
