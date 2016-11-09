#!/bin/sh

set -e


random_string()
{
    local SIZE="$1"

    LC_ALL=C tr -dc '[:alnum:]' < /dev/urandom | head -c "$SIZE"
}


KUBECTL='kubectl --context=minikube'
BASIC_AUTH_FILE='/data/auth/basic_auth.csv'
CONFIG_DIR="$HOME/.minikube/miniteresa"
IMAGE="${TERESA_IMAGE_TAG:-teresa:latest}"
KUBE_PASSWORD="$(random_string 8)"
MINIO_ACCESS_KEY="$(random_string 20)"
MINIO_SECRET_KEY="$(random_string 40)"
MINIO_BUCKET='teresa'

MINIO_DEPLOYMENT="
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: minio
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: minio
    spec:
      containers:
      - name: minio
        image: minio/minio
        command: [\"go-wrapper\", \"run\", \"server\", \"/export\"]
        ports:
          - containerPort: 9000
        env:
          - name: MINIO_ACCESS_KEY
            value: \"$MINIO_ACCESS_KEY\"
          - name: MINIO_SECRET_KEY
            value: \"$MINIO_SECRET_KEY\"
        volumeMounts:
        - mountPath: /export
          name: export
      volumes:
      - name: export
        hostPath:
          path: /data/export
"

TERESA_DEPLOYMENT="
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
        image: $IMAGE
        imagePullPolicy: Never
        ports:
        - containerPort: 8080
        env:
          - name: TERESAK8S_HOST
            value: \"https://10.0.2.15:8443\"
          - name: TERESAK8S_USERNAME
            value: \"admin\"
          - name: TERESAK8S_PASSWORD
            value: \"$KUBE_PASSWORD\"
          - name: TERESAK8S_INSECURE
            value: \"true\"
          - name: TERESABUILDER_STORAGE_AWS_KEY
            value: \"$MINIO_ACCESS_KEY\"
          - name: TERESABUILDER_STORAGE_AWS_SECRET
            value: \"$MINIO_SECRET_KEY\"
          - name: TERESABUILDER_STORAGE_AWS_REGION
            value: \"us-east-1\"
          - name: TERESABUILDER_STORAGE_AWS_BUCKET
            value: \"$MINIO_BUCKET\"
          - name: TERESABUILDER_WAIT_POD_TIMEOUT
            value: \"2m\"
          - name: TERESABUILDER_WAIT_LB_TIMEOUT
            value: \"2m\"
"

usage()
{
cat << EOF

Miniteresa is a CLI tool that provisions Teresa on a single-node Kubernetes
cluster through Minikube. By default a local docker image called teresa:latest
is used but you can override it by using the env var TERESA_IMAGE_TAG.

Requirements: Linux or Mac with minikube, docker and kubectl.

Usage:
  miniteresa.sh [command]

Available Commands:
  delete           Deletes a local kubernetes cluster.
  start            Starts a local kubernetes cluster.
  status           Gets the status of a local kubernetes cluster.
  stop             Stops a running local kubernetes cluster.

To get help on minikube: minikube help

EOF
}


retry()
{
    local MAX_RETRIES=5

    for i in $(seq "$MAX_RETRIES"); do
        "$@" && return 0 || sleep 30
    done

    echo Retry failed
    return 1
}


setup_service()
{
    local NAME="$1"
    local CONFIG="$2"

    retry $KUBECTL create namespace "$NAME"
    retry echo "$CONFIG" | $KUBECTL -n "$NAME" create -f -
    retry $KUBECTL expose deployment "$NAME" --type=NodePort -n "$NAME"
}


ssh_command() {
    local USER_HOST="docker@$(minikube ip)"

    ssh -o StrictHostKeyChecking=no \
        -i "$HOME/.minikube/machines/minikube/id_rsa" \
        "$USER_HOST" "$@" >/dev/null 2>&1
}


setup_basic_auth()
{
    local BASIC_AUTH_DIR="$(dirname $BASIC_AUTH_FILE)"

    ssh_command \
        "sudo sh -c 'mkdir $BASIC_AUTH_DIR && \
                     echo $KUBE_PASSWORD,admin,admin > $BASIC_AUTH_FILE'"
}


push_image()
{
    docker save "$IMAGE" | ssh_command 'cat | docker load'
}


wait_ready()
{
    local CMD="$KUBECTL get pods -n $1"
    local MAX_RETRIES=5

    echo Waiting for "$1"...

    for i in $(seq "$MAX_RETRIES"); do
        $CMD | grep -qwi running && return 0 || sleep 30
    done

    echo Waiting for "$1" failed
    return 1
}


show_endpoint()
{
    local PORT="$(\
        $KUBECTL get service teresa -n teresa \
            --output='jsonpath="{.spec.ports[0].NodePort}"' | tr -d \"\
    )"

    echo
    echo Teresa endpoint: http://$(minikube ip):$PORT
    echo
}


configure()
{
    minikube start >/dev/null 2>&1
    setup_basic_auth
    setup_service minio "$MINIO_DEPLOYMENT"
    push_image
    setup_service teresa "$TERESA_DEPLOYMENT"
    minikube stop >/dev/null 2>&1
    sleep 5
}


start()
{
    [ -d "$CONFIG_DIR" ] || configure

    minikube start --extra-config=apiserver.BasicAuthFile="$BASIC_AUTH_FILE"
    [ -d "$CONFIG_DIR" ] || mkdir "$CONFIG_DIR"

    wait_ready minio
    wait_ready teresa
    show_endpoint
}


check_cmd()
{
    which "$1" >/dev/null 2>&1 || {
        echo Please install "$1" before running this script
        exit 1
    }
}


check_docker()
{
    docker version >/dev/null 2>&1 || {
        echo You must be able to call docker without sudo to run miniteresa
        exit 1
    }
}


for CMD in minikube kubectl docker; do
    check_cmd "$CMD"
done

check_docker

case "$1" in
    start)
        start
        ;;
    delete)
        minikube delete
        rm -rf "$CONFIG_DIR"
        ;;
    stop)
        minikube stop
        ;;
    status)
        minikube status
        ;;
    *)
        usage
esac
