#!/bin/bash

# Function to start services
start_services() {
    # Check if correct number of arguments provided
    if [ $# -ne 2 ]; then
        echo "Usage: $0 start <container_prefix> <port>"
        echo "Example: $0 start mycompany 8080"
        exit 1
    fi

    # Assign arguments to variables
    CONTAINER_PREFIX=$1
    PORT=$2

    # Validate container prefix is not reserved
    if [ "$CONTAINER_PREFIX" = "burncloud-aiapi" ] || [ "$CONTAINER_PREFIX" = "burncloud-enterprise" ]; then
        echo "Error: Container prefix cannot be 'burncloud-aiapi' or 'burncloud-enterprise'"
        exit 1
    fi

    # Validate port is a number
    if ! [[ "$PORT" =~ ^[0-9]+$ ]]; then
        echo "Error: Port must be a number"
        exit 1
    fi

    # Validate port is 3003 or higher
    if [ "$PORT" -lt 3003 ]; then
        echo "Error: Port must be 3003 or higher"
        exit 1
    fi

    # Create required directories with 777 permissions
    echo "Creating required directories..."
    mkdir -p data mysql_data logs
    chmod -R 777 data mysql_data logs

    # Check if nginx-network exists, create if not
    echo "Checking Docker network..."
    if ! docker network ls | grep -q "nginx-network"; then
        echo "Creating nginx-network..."
        docker network create nginx-network
    else
        echo "nginx-network already exists"
    fi

    # Create a docker-compose file with replaced values
    COMPOSE_FILE="docker-compose-${CONTAINER_PREFIX}.yml"
    echo "Generating ${COMPOSE_FILE}..."

    # Copy the original file to the new file
    cp docker-compose-supplier.yml ${COMPOSE_FILE}

    # Replace container names and service names
    sed -i "s/burncloud-example-/${CONTAINER_PREFIX}-/g" ${COMPOSE_FILE}

    # Replace the port mapping
    sed -i "s/- \"[0-9]*:3000\"/- \"${PORT}:3000\"/g" ${COMPOSE_FILE}

    # Update SQL_DSN to use the new container prefix
    sed -i "s/burncloud-example-mysql/${CONTAINER_PREFIX}-mysql/g" ${COMPOSE_FILE}

    # Update Redis connection string to use the new container prefix
    sed -i "s/burncloud-example-redis/${CONTAINER_PREFIX}-redis/g" ${COMPOSE_FILE}

    # Update depends_on section to use the new container prefix
    sed -i "s/burncloud-example-redis/${CONTAINER_PREFIX}-redis/g" ${COMPOSE_FILE}
    sed -i "s/burncloud-example-mysql/${CONTAINER_PREFIX}-mysql/g" ${COMPOSE_FILE}

    # Update networks section to use the new container prefix in healthcheck
    sed -i "s/burncloud-example-redis/${CONTAINER_PREFIX}-redis/g" ${COMPOSE_FILE}
    sed -i "s/burncloud-example-mysql/${CONTAINER_PREFIX}-mysql/g" ${COMPOSE_FILE}

    # Start the services with the new file
    echo "Starting services..."
    docker-compose -f ${COMPOSE_FILE} up -d

    echo "Services started successfully!"
    echo "AI API service is running on port ${PORT}"
    echo "Docker Compose file saved as: ${COMPOSE_FILE}"
}

# Function to stop services
stop_services() {
    # Check if correct number of arguments provided
    if [ $# -ne 1 ]; then
        echo "Usage: $0 stop <container_prefix>"
        echo "Example: $0 stop mycompany"
        exit 1
    fi

    # Assign argument to variable
    CONTAINER_PREFIX=$1

    # Define the compose file name
    COMPOSE_FILE="docker-compose-${CONTAINER_PREFIX}.yml"

    # Check if the compose file exists
    if [ ! -f ${COMPOSE_FILE} ]; then
        echo "Error: ${COMPOSE_FILE} not found"
        exit 1
    fi

    # Stop the services with the compose file
    echo "Stopping services..."
    docker-compose -f ${COMPOSE_FILE} down

    echo "Services stopped successfully!"
    echo "Docker Compose file used: ${COMPOSE_FILE}"
}

# Function to list available docker-compose files
list_compose_files() {
    echo "Available docker-compose files in current directory:"
    for file in docker-compose-*.yml; do
        if [ -f "$file" ]; then
            echo "  - $file"
        fi
    done
    echo ""
}

# Main script logic
if [ $# -lt 1 ]; then
    list_compose_files
    echo "Usage: $0 {start|stop|q|quit} [arguments...]"
    echo "Start: $0 start <container_prefix> <port>"
    echo "Stop: $0 stop <container_prefix>"
    echo "Quit: $0 {q|quit}"
    echo ""
    echo "Validation rules:"
    echo "  - Container prefix cannot be 'burncloud-aiapi' or 'burncloud-enterprise'"
    echo "  - Port must be a number and 3003 or higher"
    exit 1
fi

# Get the operation
OPERATION=$1
shift

# Perform the requested operation
case $OPERATION in
    start)
        start_services $@
        ;;
    stop)
        stop_services $@
        ;;
    q|quit)
        echo "Exiting script without running any operations."
        exit 0
        ;;
    *)
        echo "Invalid operation. Use 'start', 'stop', 'q', or 'quit'."
        exit 1
        ;;
esac