### Coding Exercise
```Build a banking application consisting of an API and database backend```

### Requirements
* Docker

### Setup instructions
1. [Download](https://docs.docker.com/get-started/get-docker/) and Install docker, docker-compose if you don't have it installed
2. Clone project - 
    ```shell
   $ git clone https://github.com/adxgun/payter-bank 
   ```
3. navigate to project root and run 
    ```shell
   $ docker-compose up
    ```
docker-compose is used to run all the application components:
* Database - Postgresql: Used as the database for the application
* Redis: Used for async task queue for resilient and asynchronous audit log records
* API: Written in Golang. Exposes the application REST APIs
* frontend: Used caddy to serve the application frontend. Frontend code is written in Javascript using React frontend framework.

## Using the Frontend app and API
On startup, an admin user is automatically created with the credentials:
```
Email: admin@payterbank.app
Password: admin
```
You can change these values by setting these environment variables in `docker-compose.yml` before running the application:
```
ADMIN_EMAIL=
ADMIN_PASSWORD=
```
The application will pick up these values at startup and create a new admin account, you will then be able to log in with those details.

## Exploring the Application
I put together a basic admin frontend app that shows the basic capabilities of the app. Access it here: http://localhost/login, after you've run the `docker-compose up` command. The frontend app contains the below features:
* Creating a new account
* Viewing an account
* Viewing audit logs on an account
* Viewing account transactions
* Crediting/Debiting an account.
There are more functionalities supported by the API e.g `interest rates` but not on the frontend. You can explore these other functionalities via API. API Documentation is available via Swagger.

## Swagger Documentation
API documentation can be accessed via: https://localhost:2025/docs/index.html

## Implementation Notes

### Directory Structure

### 📁 Interactive Code Directory (Tree Viewer)

Explore the interactive file structure of this project:

👉 [View Project Structure on tree.nathanfriend.com](https://tree.nathanfriend.com/?s=(%27opt9s!(%27fancy8~fullPath!false~trailingSlash8~rootDot8)~B(%27B%27payt4-bank3featuresFccount2Fuditlog20transact920int4estrate23int4

# Project Structure Explanation

## `features/`
Hosts the core application logic, organized by domain (e.g., `account`). Each feature package typically includes:

- `api.go` – Defines HTTP handlers for that domain.
- `service.go` – Contains business logic and service layer operations.

---

## `internal/`
Contains shared internal packages that support the feature modules. These are foundational components such as:

- Database connections
- Token generation utilities
- Configuration management

> These packages are meant to be used only within the application and not exposed externally.

---

## `migrations/`
Stores all the SQL migration files. The application uses an automatic migration process to apply these schema changes to the database on startup.

---

## `server/`
Acts as the entry point and orchestration layer for the backend. Responsibilities include:

- Bootstrapping the server
- Registering routes and handlers
- Setting up authentication middleware
- Wiring services to HTTP endpoints
- Handling graceful shutdowns and application lifecycle management
