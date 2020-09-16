pipeline {
    agent any // Execute this Pipeline or any of its stages, on any available agent.
    stages {
        stage('Test') { // Defines the "Test" stage.
            steps {
                sh 'make test'
            }
        }
    }
}
