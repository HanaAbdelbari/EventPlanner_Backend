# EventPlanner Backend

Go + Gin + GORM + MySQL REST API for the EventPlanner application.

Handles user authentication, event CRUD, registrations, and JWT-based authorization.
## Project Structure
```
EventPlanner_Backend/
├── config/
│   └── database.go           # GORM connection & configuration
├── controllers/
│   ├── auth_controller.go    # login, register, refresh token...
│   └── event_controller.go   # create/update/delete events, get events...
├── middleware/
│   └── auth.go               # JWT authentication middleware
├── models/
│   ├── user.go               # User model
│   └── event.go              # Event model + associations
├── routes/
│   └── routes.go             # Gin route definitions
├── utils/
│   ├── jwt/
│   │   └── jwt.go            # JWT generation + validation
│   └── password/
│       └── password.go       # password hashing (bcrypt)
├── .env                      # environment variables
├── Dockerfile                # backend Docker build
├── docker-compose.yml        # optional full stack compose (not used here)
├── drop.sql                  # optional cleanup script
├── go.mod
├── go.sum
└── main.go                   # application entry point
```

## Docker Deployment – Backend + Database

### 1. Create Networks (do once)

```bash
  # Network used by database ↔ backend
docker network create backend-net

# Network used by backend ↔ frontend (optional here)
docker network create frontend-net
```
### 2. Create Persistent Volume for MySQL
```bash
  docker volume create eventplanner-db-data
```
### 3. Build Images
```bash
  # Database image
   cd database
   docker build -t eventplanner-db:latest .
```
```bash
  # Backend image (from backend root)
  cd ..
  docker build -t eventplanner-backend:latest .
```
### 4. Run Database Container
```bash 
   docker run -d \
   --name eventplanner-db \
   --network backend-net \
   -e MYSQL_ROOT_PASSWORD= \
   -e MYSQL_DATABASE=eventplanner \
   -v eventplanner-db-data:/var/lib/mysql \
   -p 3307:3306 \
   eventplanner-db:latest
```
```bash
   sleep 25
```
### 5. Run Backend Container
```bash
   docker run -d \
   --name eventplanner-backend \
   --network backend-net \
   -e DB_HOST=eventplanner-db \
   -e DB_PORT=3306 \
   -e DB_USER= \
   -e DB_PASSWORD= \
   -e DB_NAME=eventplanner \
   -e JWT_SECRET= \
   -e PORT=8080 \
   -p 8080:8080 \
   eventplanner-backend:latest
```
### 6. Connect Backend to Frontend Network (if running frontend separately)
  ```bash
     docker network connect frontend-net eventplanner-backend
```
## Quick Restart / Cleanup Commands
```bash
 # Stop everything
 docker stop eventplanner-backend eventplanner-db

 # Remove containers
 docker rm   eventplanner-backend eventplanner-db

 # Remove volume (warning: deletes all data!)
 docker volume rm eventplanner-db-data
```

## Optional: Using docker-compose (Alternative to manual commands)

Although the project is set up to run with individual `docker run` commands (as shown above), a `docker-compose.yml` file is also included for convenience.

### Prerequisites
- Docker Compose v2+ installed (`docker compose` command)

### Useful docker-compose commands

Start the full stack (database + backend):

```bash
# From the backend project root (where docker-compose.yml is located)
docker compose up -d
```
Stop, remove containers and delete volumes (warning: wipes database data):
```bash
  docker compose down -v
```