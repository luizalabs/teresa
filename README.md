# Teresa

Teresa is an extremely simple platform as a service that runs on top of Kubernetes.  
To have a full Teresa setup you'll need the API running as a POD on Kubernetes and the CLI.  

## API

First create a docker image, for example:

```
make build
docker tag image_id login/teresa:latest
docker push login/teresa:latest
```

Create a Kubernetes deployment and expose it as a service:

```
kubectl create namespace teresa
kubectl create -n teresa -f teresa.yml
kubectl expose deployment teresa -n teresa --type=LoadBalancer --port=80 --target-port=8080
```

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
        image: login/teresa:latest
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
            value: true
          - name: TERESABUILDER_STORAGE_AWS_KEY
            value: AWS_ACCESS_KEY
          - name: TERESABUILDER_STORAGE_AWS_SECRET
            value: AWS_SECRET_KEY
          - name: TERESABUILDER_STORAGE_AWS_REGION
            value: AWS_REGION
          - name: TERESABUILDER_STORAGE_AWS_BUCKET
            value: S3_BUCKET
```

## CLI

Steps to setup a new cluster and deploy a new application, assuming you already have the API running:

    1. teresa config set-cluster cluster_name --server https://mycluster.com
    2. teresa config use-cluster cluster_name
    3. teresa login --user myuser@mydomain.com
    4. teresa create app myapp --team dev --scale 2
    5. teresa deploy /path/to/myapp --app myapp --team dev --description "release 1.2"


## View API documentation

  To view the auto-generated swagger API documentation, the following command will compile, run a webserver and open your browser on the swagger-ui:

    cd api; make swagger-docs
