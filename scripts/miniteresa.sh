#!/bin/sh

set -e


kubectl='kubectl --context=minikube'
basic_auth_file='/data/auth/basic_auth.csv'
config_dir="$HOME/.minikube/miniteresa"
image="${TERESA_IMAGE_TAG:-teresa:latest}"
kube_password='kube2000'
minio_access_key='AKIAIOSFODNN7EXAMPLE'
minio_secret_key='wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY'
minio_bucket='teresa'
default_user='admin@admin'
default_password='teresa123'

minio_deployment="
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
            value: \"$minio_access_key\"
          - name: MINIO_SECRET_KEY
            value: \"$minio_secret_key\"
        volumeMounts:
        - mountPath: /export
          name: export
      volumes:
      - name: export
        hostPath:
          path: /data/export
"

teresa_deployment="
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: teresa
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: teresa
    spec:
      containers:
      - name: teresa
        image: $image
        imagePullPolicy: Never
        ports:
        - containerPort: 8080
        env:
          - name: TERESAK8S_HOST
            value: \"https://10.0.2.15:8443\"
          - name: TERESAK8S_USERNAME
            value: \"admin\"
          - name: TERESAK8S_PASSWORD
            value: \"$kube_password\"
          - name: TERESAK8S_INSECURE
            value: \"true\"
          - name: TERESABUILDER_STORAGE_AWS_KEY
            value: \"$minio_access_key\"
          - name: TERESABUILDER_STORAGE_AWS_SECRET
            value: \"$minio_secret_key\"
          - name: TERESABUILDER_STORAGE_AWS_REGION
            value: \"us-east-1\"
          - name: TERESABUILDER_STORAGE_AWS_BUCKET
            value: \"$minio_bucket\"
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
    local max_retries=5

    for i in $(seq "$max_retries"); do
        "$@" && return 0 || sleep 30
    done

    echo Retry failed
    return 1
}


setup_service()
{
    local name="$1"
    local config="$2"

    retry $kubectl create namespace "$name"
    retry echo "$config" | $kubectl -n "$name" create -f -
    retry $kubectl expose deployment "$name" --type=NodePort -n "$name"
}


ssh_command() {
    local user_host="docker@$(minikube ip)"

    ssh -o StrictHostKeyChecking=no \
        -i "$HOME/.minikube/machines/minikube/id_rsa" \
        "$user_host" "$@" >/dev/null 2>&1
}


setup_basic_auth()
{
    local basic_auth_dir="$(dirname $basic_auth_file)"

    ssh_command \
        "sudo sh -c 'mkdir $basic_auth_dir && \
                     echo $kube_password,admin,admin > $basic_auth_file'"
}


push_image()
{
    docker save "$image" | ssh_command 'cat | docker load'
}


wait_ready()
{
    local cmd="kubectl get pods -n $1 --context=minikube"
    local max_retries=5

    echo Waiting for "$1"...

    for i in $(seq "$max_retries"); do
        $cmd | grep -qwi running && return 0 || sleep 30
    done

    echo Waiting for "$1" failed
    return 1
}


show_endpoint()
{
    local port="$(\
        kubectl get service teresa -n teresa \
            --output='jsonpath="{.spec.ports[0].NodePort}"' \
            --context=minikube | tr -d \"\
    )"

    echo
    echo Teresa endpoint: http://$(minikube ip):$port
    echo
    echo User: "$default_user"
    echo Password: "$default_password"
}


configure()
{
    minikube start >/dev/null 2>&1
    setup_basic_auth
    setup_service minio "$minio_deployment"
    push_image
    setup_service teresa "$teresa_deployment"
    minikube stop >/dev/null 2>&1
    sleep 5
}


start()
{
    [ -d "$config_dir" ] || configure

    minikube start --extra-config=apiserver.BasicAuthFile="$basic_auth_file"
    [ -d "$config_dir" ] || mkdir "$config_dir"

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


for cmd in minikube kubectl docker; do
    check_cmd "$cmd"
done

check_docker

case "$1" in
    start)
        start
        ;;
    delete)
        minikube delete
        rm -rf "$config_dir"
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
