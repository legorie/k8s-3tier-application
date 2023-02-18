# 3 tier application deployment in K8S

The repository contains a step by step evolution of a three tier application from local development to K8S deployment.

Folder **Step_4**:

In this step 4, we'll deploy our application using a declarative YAML. In the `deployment/3tierapp.yaml` file, we have :
1) The PostGreSQL deployment, the DB password saved as a secret and a service
2) The API deployment
3) The web deployment

To test the application on your DEV machine, you'll have to forward the ports 8080 & 8090 as we have seen previously.


Folder **Step_3**:

In the third step, let us try to deploy the application in the Kubernetes environment and test it. I'm using minikube for my test K8S cluster.
Now, the images are in our local registry, inorder for the K8S environment to use our images, we need to send our images to a public registry or set up a private docker registry. Minikube comes with a handy private registry , more details here https://minikube.sigs.k8s.io/docs/handbook/registry/

Let tag and push the web & api images to the local registry.

> docker tag local/web localhost:5000/web
> docker tag local/api localhost:5000/api
> docker push localhost:5000/web
> docker push localhost:5000/api

Now we can run the webapp with a running pod or a deployment on the K8s ....
> k run web --image=localhost:5000/web

... but wait, how will the webapp talk to the api server. Our source code has localhost hardcoded in the code, time for some refactoring !

We will be refactoring the hardcoded items by the use of environment variables. The webapp and the API are modified to use a few environment variables, the port numbers are still hardcoded.

*Web :* 
The call to API server from the web app is done in the _index.html_ file. Using the html/template package, the API_HOST value is replaced in the _index.html_ file.

`var apiHost = os.Getenv("API_HOST")

func index(w http.ResponseWriter, r *http.Request) {
	tpl, _ := template.ParseFiles("./assets/index.html")
	data := map[string]string{
		"API_HOST": apiHost,
	}
	w.WriteHeader(http.StatusOK)
	tpl.Execute(w, data)
}
`
*API :*

We updated the variables used in the connection string to get values from the ENV variables


`func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var (
	host     = getEnv("DB_HOST", "localhost")
	port     = 5455
	user     = getEnv("DB_USER", "postgresUser")
	password = getEnv("DB_PASSWORD", "postgresPW")
	dbname   = "birds_db"
)`

Build the new images and push them to the private registry. Now, with the new images accessible to our minikube environment, let us get kubectl-ing.

1) First, let us create the POSTGRES pod up and running

> k run postgres --image=postgres:latest --port=5432 --env="POSTGRES_USER=postgresUser" --env="POSTGRES_PASSWORD=postgresPW"    --env="POSTGRES_DB=postgrdesDB"

K8s port mapping --target-port=5455 is possible only at a service level and not at a pod level, hence we create a new service for the port mapping.

> k expose pod postgres --target-port=5432 --port=5455

`
$ k get svc
NAME         TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)        AGE
postgres     ClusterIP   10.105.45.86   <none>        5455/TCP       3m48s

$ k get po -o wide
NAME       READY   STATUS    RESTARTS   AGE   IP           NODE       NOMINATED NODE   READINESS GATES
postgres   1/1     Running   0          45s   172.17.0.5   minikube   <none>           <none>
`

Let us also prepare the database to be usable by the application.

> k exec -it postgres -- /bin/bash

`
$ psql -h localhost -U postgresUser -d postgres
psql (15.2 (Debian 15.2-1.pgdg110+1))
Type "help" for help.

postgres=# CREATE DATABASE birds_db;
CREATE DATABASE
postgres=# \c birds_db
You are now connected to database "birds_db" as user "postgresUser".
birds_db=# CREATE TABLE birds (
  id SERIAL PRIMARY KEY,
  species VARCHAR(256),
  description VARCHAR(1024)
);
CREATE TABLE
birds_db=#

` 
 
2) Next, let us create the API POD.

> k run api --image=localhost:5000/api --env="DB_HOST=10.105.45.86" --env="DB_USER=postgresUser" --env="DB_PASSWORD=postgresPW" 
For the DB_HOST IP address, use the IP address of the postgres service IP.

Some sample error messages (1) Network not set up correctly (2) DB not setup correctly
`
>$ k logs  api
panic: dial tcp 172.17.0.5:5455: connect: connection refused

>$ k logs api
panic: pq: database "birds_db" does not exist

goroutine 1 [running]:
main.main()
	/go/src/api/main.go:51 +0x32f
`

3) Finally, let us run the webapp POD.

> k run web --image=localhost:5000/web --env="API_HOST=localhost"

`
$ k get po -o wide
NAME       READY   STATUS    RESTARTS   AGE    IP           NODE       NOMINATED NODE   READINESS GATES
api        1/1     Running   0          9m2s   172.17.0.6   minikube   <none>           <none>
postgres   1/1     Running   0          20m    172.17.0.5   minikube   <none>           <none>
web        1/1     Running   0          66s    172.17.0.7   minikube   <none>           <none>
`

To access the webapp from the local DEV machine, let us perform a port forwardding for the web POD.
> $ k port-forward pods/web 8080:8080
Forwarding from 127.0.0.1:8080 -> 8080
Forwarding from [::1]:8080 -> 8080
Handling connection for 8080

Also note that the API POD is also accessed by the local DEV machine from the browser. So, the API POD's port also has to be forwarded to be accessible

> $ k port-forward pods/api 8090:8090
Forwarding from 127.0.0.1:8090 -> 8090
Forwarding from [::1]:8090 -> 8090
Handling connection for 8090

Now, the application running on the minikube environment should be fully functional from the local DEV machine.

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