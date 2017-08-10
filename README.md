# Teresa
[![Build Status](https://travis-ci.org/luizalabs/teresa-api.svg)](https://travis-ci.org/luizalabs/teresa-api)
[![codecov](https://codecov.io/gh/luizalabs/teresa-api/branch/master/graph/badge.svg)](https://codecov.io/gh/luizalabs/teresa-api)

Teresa is an extremely simple platform as a service that runs on top of [Kubernetes](https://github.com/kubernetes/kubernetes).  
It uses a client-server model: the client sends high level commands (create application, deploy, etc.) to the server, which translates them to the Kubernetes API.

## Installation

Server requirements:

- Kubernetes cluster

- database backend to store users and teams (SQLite or MySQL)

- storage for build artifacts (Amazon S3 or minio)

- rsa keys for token signing

- (optional) TLS encription key and certificate

The recommended installation method uses the [helm](https://github.com/kubernetes/helm) package manager,
for instance to install using S3 and MySQL (recommended):

    $ openssl genrsa -out teresa.rsa
    $ export TERESA_RSA_PRIVATE=`cat teresa.rsa | base64`
    $ openssl rsa -in teresa.rsa -pubout > teresa.rsa.pub
    $ export TERESA_RSA_PUBLIC=`cat teresa.rsa.pub | base64`
    $ helm repo add luizalabs http://helm.k8s.magazineluiza.com
    $ helm install luizalabs/teresa \
        --set rsa.private=$TERESA_RSA_PRIVATE \
        --set rsa.public=$TERESA_RSA_PUBLIC \
        --set aws.key.access=xxxxxxxx \
        --set aws.key.secret=xxxxxxxx \
        --set aws.region=us-east-1 \
        --set aws.s3.bucket=teresa \
        --set db.name=teresa \
        --set db.hostname=dbhostname \
        --set db.username=teresa \
        --set db.password=xxxxxxxx

Look [here](./helm/README.md) for more information about helm options.

## QuickStart

Teresa has the concept of teams, which are just sets of users. An application
belongs to a team and all its users can perform all operations on it. There are
administrative users, which are just regular users with an admin flag set up and
only them can do user and team management.

To create an admin user you need access to the environment where the Teresa
server is running (often a Kubernetes POD):

    $ export POD_NAME=$(kubectl get pods --namespace teresa -l "app=teresa" -o jsonpath="{.items[0].metadata.name}")
    $ kubectl exec $POD_NAME -it teresa-server create-super-user --email admin_email --password xxxxxxxx -namespace teresa

Now you can start creating other users and teams. First, you need to get the
Teresa endpoint created by Kubernetes and configure the client (get it
[here](https://github.com/luizalabs/teresa-api/releases/latest)):

    $ teresa config set-cluster mycluster --server <teresa-endpoint>

This creates a new cluster called `mycluster` and sets it as the current one.
Log in and create another user and a team:

    $ teresa login --user admin_email
    $ teresa team create myteam --email myemail
    $ teresa create user --name myname --email myemail --password xxxxxxxx
    $ teresa team add-user --team myteam --user myemail

This new user is able to create and deploy applications on behalf of the team:

    $ teresa login --user myemail
    $ teresa app create myapp --team myteam
    $ teresa deploy /path/to/myapp --app myapp --description "release 1.0"

Teresa has an extensive help builtin, you can access it with:

    $ teresa --help

Check out some examples [here](https://github.com/luizalabs/hello-teresa) to
make sure that your application is ready for Teresa.
