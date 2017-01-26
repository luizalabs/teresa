# Teresa

Teresa is an extremely simple platform as a service that runs on top of [Kubernetes](https://github.com/kubernetes/kubernetes).  
The API needs a database backend and access to Amazon S3 for storage.  
To have a full Teresa setup you'll need the API running as a POD and the [CLI](https://github.com/luizalabs/teresa-cli).  

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

    $ python3 -c 'import bcrypt; print(bcrypt.hashpw("mypassword".encode(), bcrypt.gensalt()))'

### Running as a POD

First create and push a docker image:

    $ docker build -t <your-login>/teresa:latest .
    $ docker push <your-login>/teresa:latest

Create a deployment and expose it as a service:

    $ kubectl create namespace teresa
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
```

## CLI

Steps to setup a new cluster and deploy a new application, assuming you already have the API running:

    $ teresa config set-cluster mycluster --server <k8s-endpoint>
    $ teresa config use-cluster mycluster
    $ teresa login --user myuser@mydomain.com

Create a new team (optional, requires admin privileges):

    $ teresa team create myteam
    $ teresa team add-user --team myteam --user myuser@mydomain.com

Finally create and deploy the application:

    $ teresa app create myapp --team myteam
    $ teresa deploy /path/to/myapp --app myapp --description "release 1.0"

## View API Documentation

To view the auto-generated swagger API documentation, the following command will compile, run a webserver and open your browser on the swagger-ui:

    $ make swagger-docs

## Deploy

Some informations to up and run your application on Kubernetes with Teresa.

### Port
Don't use a fixed port to up your web application, instead,
read the environment variable `PORT`, for instance:

```go
port := os.Getenv("PORT")
if port == "" {
    port = "5000"
}
http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
```

The deploy process will set this environment variable

### Procfile
According to [Heroku's docs](https://devcenter.heroku.com/articles/procfile):

```
A Procfile is a mechanism for declaring what commands are run by your application’s
dynos on the Heroku platform.
```

Teresa follows the same principle.
As an example, a Python application might have the following command on Procfile:

    web: gunicorn -b 0.0.0.0:$PORT -w 3 --pythonpath src myapp.wsgi


### Language detection
When you make a deploy on Teresa you don't need to specify the language of your application,
because, Teresa has a _language detection_ step in it's deploy pipeline.

> This step it's based on [Heroku's build packs](https://devcenter.heroku.com/articles/buildpacks).

#### Golang
Teresa will detect your application as Golang if you're using one of theses depedencies managers:

- [govendor](https://github.com/kardianos/govendor)
- [glide](https://github.com/Masterminds/glide)
- [GB](https://getgb.io/)
- [Godep](https://github.com/tools/godep)

If you don't need to deal with third party libs you just need to drop a simple `vendor/vendor.json`
file in the root dir of your application, for instance:

```json
{
  "comment": "",
  "ignore": "test",
  "package": [],
  "rootPath": "github.com/luizalabs/hello-teresa"
}
```

#### Python
To deploy a Python application on Teresa a `requirements.txt` file must be present in the root dir
of your application.  
The version of Python runtime can be specified with a `runtime.txt` file in the root dir, for instance:

    $ cat runtime.txt
    python-3.6.0

#### NodeJS
Teresa will detect your application as NodeJS when the application has a `package.json` file in the root dir.  
If no _Procfile_ is present in the root directory of your application during the build step,
your web process will be started by running `npm start`, a script you can specify in _package.json_, for instance:

```json
  "scripts": {
    "start": "node server.js"
  },
```

### teresa.yaml
Some features can be configured in a file called `teresa.yaml` in the the root dir of application.

#### Health Check
Kubernetes has two types of [health checks](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/),
the `Readiness` and the `Liveness`.

- **Readiness**: Based on the time of "boot" of application, the Kubernetes uses this configuration to know when container is ready to start acception traffic.
- **Liveness**: Conventional health check, the Kubernetes uses this configuration to know when to restart a container.

You can set both (_readiness_ and _liveness_) for your application in section `healthCheck` of the _teresa.yaml_, for instance:

```yaml
healthCheck:
    liveness:
        path: /healthcheck/
        timeoutSeconds: 2
        initialDelaySeconds: 10
        periodSeconds: 5
        failureThreshold: 2
        successThreshold: 1
    readiness:
        path: /healthcheck/
        timeoutSeconds: 5
        initialDelaySeconds: 5
        periodSeconds: 5
        failureThreshold: 5
        successThreshold: 1
```

> Teresa only perform health check based on _HTTP GET request_.

- **path**: endpoint of application than health check should hit.
- **timeoutSeconds**: timeout to determine if the application is unhealthy.
- **initialDelaySeconds**: delay (in seconds) to start to perform the execution of health check.
- **periodSeconds**: delay between checks.
- **failureThreshold**: max failure tolerance before restart the container.
- **successThreshold**: min number of success to determina that container it's healthy.

Any code greater than or equeal to 200 and less than 400 indicates success.
Any other code indicates failure.

#### Rolling Update
Kubernetes has the [Rolling Update](https://kubernetes.io/docs/user-guide/deployments/#rolling-update-deployment) strategy to deal with deploys.
With this strategy you can specify the `max unavailable` and the `max surge` fields to control
the rolling update process.  
You can set both (_maxUnavailable_ and _maxSurge_) for the deploy of your application in section
`RollingUpdate` of the _teresa.yaml_, for instance:

```yaml
rollingUpdate:
    maxUnavailable: "30%"
    maxSurge: "2"
```

- **Max Unavailable**: Specifies the maximum number of pods can be unavailable during the update process.
- **Max Surge**: Specifies the maximyum number of pods can be created above the desired number of pods.

> This field can be an absolute number (e.g. "2") or a percentage (e.g. "30%").
