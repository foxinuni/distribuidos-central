# Build the docker image
docker build -t dist-tools .

# Database URL environment variable
export DATABASE_URL=postgres://postgres:postgres@127.0.0.1/distribuidos_central?sslmode=disable

# Perform migrations
docker run --rm \
    --network host \
    -e DATABASE_URL=${DATABASE_URL} \
    dist-tools task migrate

# Populate the database
docker run --rm \
    --network host \
    dist-tools populate -classrooms 350 -laboratories 100 -database ${DATABASE_URL}

# Run central server
docker run --rm \
    --network host \
    dist-tools central -port 5555 -workers 20 -database ${DATABASE_URL}

# Run faculty
docker run --rm \
    --network host \
    -v ./logs/:/app/logs \
    dist-tools faculty -faculties 10 -address tcp://127.0.0.1:5555