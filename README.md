# Teresa
[![Build Status](https://travis-ci.org/luizalabs/teresa-api.svg)](https://travis-ci.org/luizalabs/teresa-api)

Teresa is an extremely simple platform as a service that runs on top of [Kubernetes](https://github.com/kubernetes/kubernetes).  
The API needs a database backend and access to Amazon S3 for storage.  
To have a full Teresa setup you'll need the API running as a POD and the CLI.

**Note**: the master branch is a development branch (unstable). Use the [**latest stable release**](https://github.com/luizalabs/teresa-api/releases/latest).

## API

### Database Backends

#### SQLite

This is the default. No configuration needed: a database file `teresa.sqlite` is automatically used.

#### MySQL

A non-empty TERESADB_HOSTNAME environment variable automatically selects this backend. The
other variables are:

* TERESADB_PORT
* TERESADB_USERNAME
* TERESADB_PASSWORD
* TERESADB_DATABASE

### Creating Users

For now you have to manually insert a row into the `users` table, for example on MySQL:

    insert into users
    (created_at, updated_at, name, email, password, is_admin)
    values
    (NOW(), NOW(), "myuser", "myuser@mydomain.com", "hashed_password", 0);

To generate a bcrypt hashed password:

    $ curl -d 'mypassword' http://hashpass.k8s-test.magazineluiza.com.br

### Running as a POD

First create and push a docker image:

    $ docker build -t <your-login>/teresa:latest .
    $ docker push <your-login>/teresa:latest

Create rsa keys and secrets for token signing:

    $ openssl genrsa -out teresa.rsa
    $ openssl rsa -in teresa.rsa -pubout > teresa.rsa.pub
    $ kubectl create namespace teresa
    $ kubectl create secret generic teresa-keys --from-file=./teresa.rsa --from-file=./teresa.rsa.pub -n teresa

At last create a deployment and expose it as a service:

    $ kubectl create -n teresa -f teresa.yml
    $ kubectl expose deployment teresa -n teresa --type=LoadBalancer --port=80 --target-port=8080

where a typical teresa.yml is:

```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: teresa
spec:
  replicas: 2
  template:
    metadata:
      labels:
        app: teresa
    spec:
      containers:
      - name: teresa
        image: <your-login>/teresa:latest
        ports:
        - containerPort: 8080
        env:
          - name: TERESAK8S_HOST
            value: KUBERNETES_API_ENDPOINT
          - name: TERESAK8S_USERNAME
            value: KUBERNETES_API_USERNAME
          - name: TERESAK8S_PASSWORD
            value: KUBERNETES_API_PASSWORD
          - name: TERESAK8S_INSECURE
            value: "true"
          - name: TERESAFILESTORAGE_TYPE
            value: s3
          - name: TERESAFILESTORAGE_AWS_KEY
            value: AWS_ACCESS_KEY
          - name: TERESAFILESTORAGE_AWS_SECRET
            value: AWS_SECRET_KEY
          - name: TERESAFILESTORAGE_AWS_REGION
            value: AWS_REGION
          - name: TERESAFILESTORAGE_AWS_BUCKET
            value: S3_BUCKET
          - name: TERESADEPLOY_REVISION_HISTORY_LIMIT
            value: "5"
          - name: TERESADEPLOY_START_TIMEOUT
            value: 10m
          - name: TERESADEPLOY_FINISH_TIMEOUT
            value: 30m
          - name: NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
```

The TERESADEPLOY environment variables are optional.

## View API Documentation

To view the auto-generated swagger API documentation, the following command will compile, run a webserver and open your browser on the swagger-ui:

    $ make swagger-docs

## CLI

Build the CLI:

    $ make build-client

Steps to setup a new cluster and deploy a new application, assuming you already have the API running:

    $ teresa config set-cluster mycluster --server <teresa-endpoint>
    $ teresa config use-cluster mycluster
    $ teresa login --user myuser@mydomain.com

Create a new team (optional, requires admin privileges):

    $ teresa team create myteam
    $ teresa team add-user --team myteam --user myuser@mydomain.com

Finally create and deploy the application:

    $ teresa app create myapp --team myteam
    $ teresa deploy /path/to/myapp --app myapp --description "release 1.0"

Make sure your application is ready for Teresa.  
Check out some examples [here](https://github.com/luizalabs/hello-teresa).
