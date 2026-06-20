pipeline {
    agent { label 'billing-agent' }

    environment {
        SKIP_DATA_IMPORT = 'true'
        PERSISTENT_ENV = "${HOME}/.billing-env/backend.env"
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
                    if [ -f "$PERSISTENT_ENV" ]; then
                        cp "$PERSISTENT_ENV" backend/.env
                        echo "Using existing .env from $PERSISTENT_ENV"
                    else
                        cp backend/.env.example backend/.env

                        # Generate Fernet key once
                        KEY=$(python3 -c "from cryptography.fernet import Fernet; print(Fernet.generate_key().decode())" 2>/dev/null || \\
                              dd if=/dev/urandom bs=1 count=32 2>/dev/null | base64 | tr '+/' '-_' | tr -d '=\\n')
                        if [ -n "$KEY" ]; then
                            sed -i "s|ENCRYPTION_KEY=\".*\"|ENCRYPTION_KEY=\"$KEY\"|" backend/.env
                        fi

                        # Generate JWT SECRET_KEY once
                        JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || \\
                                     dd if=/dev/urandom bs=1 count=32 2>/dev/null | base64 | tr -d '=\\n')
                        if [ -n "$JWT_SECRET" ]; then
                            sed -i "s|SECRET_KEY=\".*\"|SECRET_KEY=\"$JWT_SECRET\"|" backend/.env
                        fi

                        # Save persistently
                        mkdir -p "$(dirname "$PERSISTENT_ENV")"
                        cp backend/.env "$PERSISTENT_ENV"
                        echo "New .env created and saved to $PERSISTENT_ENV"
                    fi
                '''
            }
        }

        stage('Build & Deploy') {
            steps {
                sh '''
                    docker compose up -d --build
                    docker compose restart nginx-proxy
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
