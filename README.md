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
