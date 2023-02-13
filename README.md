# 3 tier application deployment in K8S

The repository contains a step by step evolution of a three tier application from local development to K8S deployment.


Folder **Step_2**:

In this second step, we will dockerize the application.

*Web :* 

The Dockerfile in the web folder performs a multi stage build. A key learning in this docker build process is to understand the folder structure (local vs docker build image vs final docker image) and the execution of the ENTRYPONT.

> docker build -t local/web .
> docker run -d -p 8080:8080 local/web

This should have the web container running and the web app accessible from the browser of the dev machine.

*API :*

The Dockerfile in the API folder also follows the multistage build. While running the *docker run* command,we use the network option as host, so that the API container can reach the DB container running on the local DEV machine.

> docker build -t local/api .
> docker run -d  --network=host local/api

The webapp is fully function on the local DEV machine and here is an example output of the __docker ps__ command

```
CONTAINER ID   IMAGE       COMMAND                  CREATED          STATUS          PORTS                                       NAMES
202f87a7f5b4   local/api   "./api"                  57 seconds ago   Up 56 seconds                                               hungry_meninsky
0eafae7b54b1   local/web   "./web"                  16 minutes ago   Up 16 minutes   0.0.0.0:8080->8080/tcp, :::8080->8080/tcp   friendly_allen
1bbaf0cc3891   postgres    "docker-entrypoint.s…"   2 weeks ago      Up 2 weeks      0.0.0.0:5455->5432/tcp, :::5455->5432/tcp   myPostgresDb
taku
```




Folder **Step_1**:

The application runs on localhost with user IDs, passwords and port numbers hardcoded in the code. This step is just to make sure the code is working fine.

*Web :* 
You will find a simple Golang web application running on port 8080. The web application calls the API server to access the DB. The usage of the CORS module helps us to run the application well in our local environment. 
`> go run main.go
Starting server in port 8080` 

*Database :*
Instead of installing a DB on the localhost, a choice was made to run the DB as a docker container
`
> docker run --name myPostgresDb -p 5455:5432 -e POSTGRES_USER=postgresUser -e POSTGRES_PASSWORD=postgresPW    -e POSTGRES_DB=postgresDB -d postgres

> psql -h localhost -p 5455 -U postgresUser -d postgres 

> CREATE DATABASE birds_db;

> \c birds_db

> CREATE TABLE birds (
  id SERIAL PRIMARY KEY,
  species VARCHAR(256),
  description VARCHAR(1024)
); 
` 

*API :*
A simple API server using Gorilla MUX running on port 8090. There are 2 APIs:
localhost:8090/bird [GET]  ==> List all the birds  available in the database
localhost:8090/bird [POST] ==> Adds a new bird to the database

The API server connects to the PostgreSQL server exposed on 5455

`
> go run *.go

API server in port 8090
`

Thanks to : Soham Kamani (https://medium.com/gojekengineering/adding-a-database-to-a-go-web-application-b0e8e8b16fb9)

https://www.weave.works/blog/deploying-an-application-on-kubernetes-from-a-to-z