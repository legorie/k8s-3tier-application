# 3 tier application deployment in K8S

The repository contains a step by step evolution of a three tier application from local development to K8S deployment. The application has a web component, an api server and a database.

Step 1: Web, API and DB running on localhost  
Step 2: Dockerize the web and api applications on localhost  
Step 3: Running the application on our local Kubernetes cluster  
Step 4: Deploy in the Kubernetes cluster using a declarative YAML  
Step 5: Use Ingress controller to manage the traffic  
Step 6: Converting the deployment to a Helm chart  
 
## Folder **Step_1**:

The application runs on localhost with user IDs, passwords and port numbers hardcoded in the code. This step is just to make sure the code is working fine.

*Web :* 
You will find a simple Golang web application running on port 8080. The web application calls the API server to access the DB. The usage of the CORS module helps us to run the application well in our local environment. 

```
> go run main.go
Starting server in port 8080
``` 

*Database :*
Instead of installing a DB on the localhost, a choice was made to run the DB as a docker container
```
$ docker run --name myPostgresDb -p 5455:5432 -e POSTGRES_USER=postgresUser -e POSTGRES_PASSWORD=postgresPW    -e POSTGRES_DB=postgresDB -d postgres

$ psql -h localhost -p 5455 -U postgresUser -d postgres 

$> CREATE DATABASE birds_db;

$> \c birds_db

$> CREATE TABLE birds (
  id SERIAL PRIMARY KEY,
  species VARCHAR(256),
  description VARCHAR(1024)
); 
```

*API :*
A simple API server using Gorilla MUX running on port 8090. There are 2 APIs:
localhost:8090/bird [GET]  ==> List all the birds  available in the database
localhost:8090/bird [POST] ==> Adds a new bird to the database

The API server connects to the PostgreSQL server exposed on 5455

```
> go run *.go

API server in port 8090
```

## Folder **Step_2**:

In this second step, we will dockerize the application.

*Web :* 

The Dockerfile in the web folder performs a multi stage build. A key learning in this docker build process is to understand the folder structure (local vs docker build image vs final docker image) and the execution of the ENTRYPONT.

> docker build -t local/web .
> docker run -d -p 8080:8080 local/web

This should have the web container running and the web app accessible from the browser of the dev machine(VM).

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

## Folder **Step_3**:

In the third step, let us try to deploy the application in the Kubernetes environment and test it. We'll be using *minikube* in our test K8S cluster.
Now, the images are in our local registry, inorder for the K8S environment to use our images, we need to send our images to a public registry or set up a private docker registry. Minikube comes with a handy private registry , more details here https://minikube.sigs.k8s.io/docs/handbook/registry/

Let tag and push the web & api images to the local registry.

> $ docker tag local/web localhost:5000/web
> $ docker tag local/api localhost:5000/api
> $ docker push localhost:5000/web
> $ docker push localhost:5000/api

Now we can run the webapp with a running pod or a deployment on the K8s ....
> $ k run web --image=localhost:5000/web

... but wait, how will the webapp talk to the api server. Our source code has localhost hardcoded in the code, time for some refactoring !

We will be refactoring the hardcoded items by the use of environment variables. The webapp and the API are modified to use a few environment variables, the port numbers are still hardcoded.

*Web :* 
The call to API server from the web app is done in the _index.html_ file. Using the html/template package, the API_HOST value is replaced in the _index.html_ file.

```
var apiHost = os.Getenv("API_HOST")

func index(w http.ResponseWriter, r *http.Request) {
	tpl, _ := template.ParseFiles("./assets/index.html")
	data := map[string]string{
		"API_HOST": apiHost,
	}
	w.WriteHeader(http.StatusOK)
	tpl.Execute(w, data)
}
```
*API :*

We updated the variables used in the connection string to get values from the ENV variables

```
func getEnv(key, fallback string) string {
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
)
```

Build the new images and push them to the private registry. Now, with the new images accessible to our minikube environment, let us get kubectl-ing.

1) First, let us create the POSTGRES pod up and running

> $ k run postgres --image=postgres:latest --port=5432 --env="POSTGRES_USER=postgresUser" --env="POSTGRES_PASSWORD=postgresPW"    --env="POSTGRES_DB=postgrdesDB"

K8S port mapping --target-port=5455 is possible only at a service level and not at a pod level, hence we create a new service for the port mapping.

> $ k expose pod postgres --target-port=5432 --port=5455

```
$ k get svc
NAME         TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)        AGE
postgres     ClusterIP   10.105.45.86   <none>        5455/TCP       3m48s

$ k get po -o wide
NAME       READY   STATUS    RESTARTS   AGE   IP           NODE       NOMINATED NODE   READINESS GATES
postgres   1/1     Running   0          45s   172.17.0.5   minikube   <none>           <none>
```

Let us also prepare the database to be usable by the application.

> $ k exec -it postgres -- /bin/bash

```
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

```
 
2) Next, let us create the API POD.

> $ k run api --image=localhost:5000/api --env="DB_HOST=10.105.45.86" --env="DB_USER=postgresUser" --env="DB_PASSWORD=postgresPW" 
For the DB_HOST IP address, use the IP address of the postgres service IP.

Some sample error messages (1) Network not set up correctly (2) DB not setup correctly

```
> $ k logs  api
panic: dial tcp 172.17.0.5:5455: connect: connection refused

> $ k logs api
panic: pq: database "birds_db" does not exist

goroutine 1 [running]:
main.main()
	/go/src/api/main.go:51 +0x32f
```

3) Finally, let us run the webapp POD.

> $ k run web --image=localhost:5000/web --env="API_HOST=localhost"

```
$ k get po -o wide
NAME       READY   STATUS    RESTARTS   AGE    IP           NODE       NOMINATED NODE   READINESS GATES
api        1/1     Running   0          9m2s   172.17.0.6   minikube   <none>           <none>
postgres   1/1     Running   0          20m    172.17.0.5   minikube   <none>           <none>
web        1/1     Running   0          66s    172.17.0.7   minikube   <none>           <none>
```

To access the webapp from the local DEV machine, let us perform a port forwardding for the web POD.

```
 $ k port-forward pods/web 8080:8080
Forwarding from 127.0.0.1:8080 -> 8080
Forwarding from [::1]:8080 -> 8080
Handling connection for 8080
```

Also note that the API POD is also accessed by the local DEV machine from the browser. So, the API POD's port also has to be forwarded to be accessible
```
> $ k port-forward pods/api 8090:8090
Forwarding from 127.0.0.1:8090 -> 8090
Forwarding from [::1]:8090 -> 8090
Handling connection for 8090
```
Now, the application running on the minikube environment should be fully functional from the local DEV machine.

## Folder **Step_4**:

In this Step 4, we'll deploy our application using a declarative YAML. In the `deployment/3tierapp.yaml` file, we have :
1) The PostGreSQL deployment, the DB password saved as a secret and a service
2) The API deployment
3) The web deployment

> $ kubectl apply -f deployment/3tierapp.yaml

To test the application on your DEV machine, you'll still have to forward the ports 8080 & 8090 as we have seen previously.

## Folder **Step_5**:

In this step 5, we will try to use an Ingress Controller to route the traffic to the web and api applications.
The web app had to be modified for the way the new ingress is being setup (to the call the API), so we will have to re-build and push the web application (to the local registry) with new source code in the Step_5 folder.

```
$ docker build Step_5/web/. -t local/web
$ docker tag local/web localhost:5000/web
$ docker push localhost:5000/web
```

We can also recreate the application deployments & services as needed :
> $ kubectl apply -f deployment/3tierapp.yaml

As we are using minikube, on of the easier ways is deploy the ingress controller via minikube :
> $ minikube addons enable ingress

An alternative would be to deploy via the manifest of the nginx controller :
> $ kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.4.0/deploy/static/provider/cloud/deploy.yaml

> $ kubectl get all -n ingress-nginx

Now, let us create the ingress for our application. Our new ingress points to the web application via the _web-svc_ service and to the api via the _api-svc_
N.B. Do learn about the annotations - rewrite-path and the path regex

```
$ k apply -f deployment/ingress.yaml
ingress.networking.k8s.io/frontend-ingress created
```

We can get the IP address of the ingress using the below kubectl command :

```
$ kubectl get ing
NAME               CLASS   HOSTS         ADDRESS        PORTS   AGE
frontend-ingress   nginx   k8sapp.info   192.168.49.2   80      8m1s
```
Also, in the /etc/hosts file, let us add a mapping for the URL in our ingress to the IP address assigned to the ingress.
```
> $ cat /etc/hosts
127.0.0.1	localhost
192.168.49.2    k8sapp.info
```
Now, the web app is functional from the browser on the VM via the URL http://k8sapp.info/web

## Folder **Step_6**:

In this final Step 6, we'll try to convert the deployment YAML into a Helm chart. To crate a new helm chart, we use :

> $ helm create my3tierapp_temp

Now, in the new Helm folder created, we'll have to move the deployment code to the _template/deployment.yaml_ , services to _template/service.yaml_ and any variables to be parametarized in the _values.yaml_

There is also a tool called helmify[https://github.com/arttor/helmify] which helped in converting the Kubernetes YAML files (created in the previous step) to a Helm chart.

> $ cat deployment/3tierapp.yaml ingress.ymal | *helmify* my3tierapp

This creates a new helm chart called _my3tierapp_. Try to install the helm chart, you'll surely face some issues, try to fix them, it gives us a great learning opportunity. For reference, the working helm chart code is present in the Step_6 folder

```
$ helm install my3tierapp my3tierapp/
NAME: my3tierapp
LAST DEPLOYED: Sat Feb 25 19:29:59 2023
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
$ helm list
NAME      	NAMESPACE	REVISION	UPDATED                                STATUS  	CHART           	APP VERSION
my3tierapp	default  	1       	2023-02-25 19:29:59.214309728 +0100 CETdeployed	my3tierapp-0.1.0	0.1.0
```

This created the Kubernetes objects for our application :

```
$ k get ing
NAME                          CLASS   HOSTS         ADDRESS        PORTS   AGE
my3tierapp-frontend-ingress   nginx   k8sapp.info   192.168.49.2   80      4m46s
$ k get deploy
NAME                        READY   UP-TO-DATE   AVAILABLE   AGE
my3tierapp-api-deployment   1/1     1            1           4m53s
my3tierapp-postgres         1/1     1            1           4m53s
my3tierapp-web-deployment   1/1     1            1           4m53s
$ k get pods
NAME                                         READY   STATUS    RESTARTS        AGE
my3tierapp-api-deployment-86bfc7684-kcdsp    1/1     Running   1 (4m57s ago)   5m5s
my3tierapp-postgres-58fbdc79fc-c8w9g         1/1     Running   0               5m5s
my3tierapp-web-deployment-6f69c9f678-526t2   1/1     Running   0               5m5s
$ k get svc
NAME                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
kubernetes                    ClusterIP   10.96.0.1       <none>        443/TCP          111d
my3tierapp-api-svc            NodePort    10.103.247.31   <none>        8090:32393/TCP   5m8s
my3tierapp-postgres-service   ClusterIP   10.99.125.144   <none>        5455/TCP         5m8s
my3tierapp-web-svc            NodePort    10.111.54.43    <none>        8080:32362/TCP   5m8s
```

The webapp and APIs are functional and can be tested on the local VM as we did previously.


Thanks to : 
Soham Kamani (https://medium.com/gojekengineering/adding-a-database-to-a-go-web-application-b0e8e8b16fb9)
https://www.weave.works/blog/deploying-an-application-on-kubernetes-from-a-to-z
