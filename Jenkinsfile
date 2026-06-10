pipeline {
    agent { label 'billing-agent' }

    environment {
        DOCKER_REGISTRY = 'registry.jeknis.local'
        IMAGE_TAG = "${env.BUILD_NUMBER}-${env.GIT_COMMIT.take(7)}"
        SONAR_HOST_URL = 'http://192.168.222.55:9000'
        SONAR_PROJECT_KEY = 'billing-revaktor'
        SKIP_DATA_IMPORT = 'true'
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('SonarQube Analysis') {
            steps {
                withSonarQubeEnv('SonarQube') {
                    sh '''
                        docker run --rm \
                            -e SONAR_HOST_URL="${SONAR_HOST_URL}" \
                            -v "${WORKSPACE}:/usr/src" \
                            sonarsource/sonar-scanner-cli \
                            -Dsonar.projectKey=${SONAR_PROJECT_KEY} \
                            -Dsonar.sources=backend/cmd,backend/internal,backend/pkg,frontend/src \
                            -Dsonar.host.url=${SONAR_HOST_URL} \
                            -Dsonar.login=${SONAR_TOKEN}
                    '''
                }
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
