pipeline {
    agent { label 'billing-agent' }

    environment {
        DOCKER_REGISTRY = 'registry.jeknis.local'
        IMAGE_TAG = "${env.BUILD_NUMBER}-${env.GIT_COMMIT.take(7)}"
        SKIP_DATA_IMPORT = 'true'
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Build Backend') {
            steps {
                dir('backend') {
                    sh 'docker build -t ${DOCKER_REGISTRY}/billing-backend:${IMAGE_TAG} .'
                }
            }
        }

        stage('Build Frontend') {
            steps {
                dir('frontend') {
                    sh 'docker build --build-arg VITE_API_BASE_URL=/api/v1 -t ${DOCKER_REGISTRY}/billing-frontend:${IMAGE_TAG} .'
                }
            }
        }

        stage('Deploy to Testing') {
            steps {
                sh '''
                    # Export env variables untuk compose
                    export SKIP_DATA_IMPORT=true
                    export IMAGE_TAG=${IMAGE_TAG}
                    export DOCKER_REGISTRY=${DOCKER_REGISTRY}

                    # Deploy with override for testing (skip data import)
                    docker compose -f docker-compose.yml -f docker-compose.override.yml up -d
                '''
            }
        }
    }

    post {
        always {
            sh 'docker system prune -f'
        }
        success {
            echo 'CI/CD Pipeline berhasil!'
        }
        failure {
            echo 'CI/CD Pipeline gagal!'
        }
    }
}
