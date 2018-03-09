# Teresa
[![Release](https://img.shields.io/github/release/luizalabs/teresa.svg?style=flat-square)](https://github.com/luizalabs/teresa/releases/latest)
[![Software License](https://img.shields.io/badge/license-apache-brightgreen.svg?style=flat-square)](/LICENSE.md)
[![Build Status](https://img.shields.io/travis/luizalabs/teresa/master.svg?style=flat-square)](https://travis-ci.org/luizalabs/teresa)
[![codecov](https://img.shields.io/codecov/c/github/luizalabs/teresa/master.svg?style=flat-square")](https://codecov.io/gh/luizalabs/teresa)
[![Go Report Card](https://goreportcard.com/badge/github.com/luizalabs/teresa?style=flat-square)](https://goreportcard.com/report/github.com/luizalabs/teresa)

Teresa is an extremely simple platform as a service that runs on top of [Kubernetes](https://github.com/kubernetes/kubernetes).
It uses a client-server model: the client sends high level commands (create application, deploy, etc.) to the server, which translates them to the Kubernetes API.

## Installation

Server requirements:

- Kubernetes cluster

- database backend to store users and teams (SQLite or MySQL)

- storage for build artifacts (AWS S3 or minio)

- rsa keys for token signing

- (optional) TLS encription key and certificate

The recommended installation method uses the [helm](https://github.com/kubernetes/helm) package manager,
for instance to install using S3 and MySQL (recommended):

    $ openssl genrsa -out teresa.rsa
    $ export TERESA_RSA_PRIVATE=`base64 teresa.rsa`  # use base64 -w0 on Linux
    $ openssl rsa -in teresa.rsa -pubout > teresa.rsa.pub
    $ export TERESA_RSA_PUBLIC=`base64 teresa.rsa.pub`
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
server is running (often a Kubernetes POD in namespace `teresa`):

    $ export POD_NAME=$(kubectl get pods --namespace teresa -l "app=teresa" -o jsonpath="{.items[0].metadata.name}")
    $ kubectl exec $POD_NAME -it teresa-server create-super-user --email admin_email --password xxxxxxxx --namespace teresa

Now you can start creating other users and teams. First, you need to get the
Teresa endpoint created by Kubernetes and configure the client (get it
[here](https://github.com/luizalabs/teresa/releases/latest)):

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

## FAQ

### Config

**Q: How to list the available clusters (aka environments)?**

    $ teresa config view

**Q: How to add/update a cluster?**

    $ teresa config set-cluster <cluster-name> --server <cluster-endpoint>

**Q: How to start using a cluster?**

    $ teresa config use-cluster <cluster-name>
    $ teresa login --user <email>

### App

**Q: How to create an app?**

    $ teresa app create <app-name> --team <team-name>

**Q: How to create an app without a load balancer (a worker for example) ?**

    $ teresa app create <app-name> --team <team-name> --process-type worker

You also have to adjust the `Procfile` to have a corresponding `worker` key.
There's nothing special with the name `worker`, it can be anything different
from `web` with a matching `Procfile` line.

**Q: How to create an internal app (without a external endpoint) ?**

    $ teresa app create <app-name> --team <team-name> --internal

The app is only visible inside the cluster. The flag `--internal` only works
with the `web` process type (the default).

**Q: How to get info about an app?**

    $ teresa app info <app-name>

**Q: How to get the list of apps I have access to?**

    $ teresa app list

**Q: How to get app logs?**

    $ teresa app logs <app-name>

**Q: How to set an environment variable?**

    $ teresa app env-set KEY=VALUE --app <app-name>

**Q: How to unset an environment variable?**

    $ teresa app env-unset KEY --app <app-name>

**Q: How to deploy an app?**

    $ teresa deploy create /path/to/project --app <app-name> --description "version 1.0"

**Q: How to set up Kubernetes health checks?**

Take a look at [here](https://github.com/luizalabs/hello-teresa#teresayaml).

**Q: I need one `teresa.yaml` per process type, how to proceed?**

If a file named `teresa-processtype.yaml` is found it is used instead of
`teresa.yaml`.

**Q: How to drain connections on shutdown?**

You can make the pods wait a fixed amount of seconds (maximum 30) before
receiving the *SIGTERM* signal by adding this lines to `teresa.yaml`:

```yaml
lifecycle:
  preStop:
    drainTimeoutSeconds: 10
```

By default teresa adds a 10 seconds drain timeout.

**Q: What's the deployment strategy?**

Teresa creates a rolling update deployment, which updates a fixed number of
pods at a time. Take a look [here](https://github.com/luizalabs/hello-teresa#rolling-update)
on how to configure the rolling update process.

**Q: How to perform tasks before a new release is deployed?**

There's a special kind of process called **release**, which is executed right
after the build process and before the rolling update. It is useful for tasks
such as sending javascript to a CDN or running database schema migrations. For
example, if you are running a django based application you may configure
automatic migrations by adding this line to the `Procfile`:

    release: python django/manage.py migrate

Note that a failing release will prevent the rolling update from happening, so
you have to keep compatibility with old code.

**Q: How to use nginx in front of my app (sidecar)?**
Create a nginx.conf file in the project root directory, for example:

```
events {
  worker_connections  1024;
}

http{
  server {
    listen $NGINX_PORT;
    location / {
        proxy_pass $NGINX_BACKEND;
    }
  }
}
```

Note that this must be the **full** nginx configuration and the environment
variables will be automatically filled.

**Q: How to get nginx logs?**

Filter the logs by container name:

    $ teresa app logs <app-name> --container nginx

**Q: How to enable ssl?**

For now we only support AWS and you need the certificate ARN:

    $ teresa service enable-ssl --app myapp --cert arn:aws:iam::xxxxx:server-certificate/cert-name

To only use SSL pass the flag `--only`.

### CronJob

**Q: How to create a CronJob?**

Teresa will infer if an app is a cronjob by checking if the prefix of _process-type_
is `cron` (e.g. `cron`, `cronjob`, `cron-method-a`).

    $ teresa app create <app-name> --team <team-name> --process-type cron

You can define the schedule of cron by adding this lines to `teresa.yaml`:

```
cron:
  schedule: "*/30 * * * *"
```

Make sure to have the related (i.e. same value of the process-type) key with the cronjob command on the `Procfile`.

### Development

**Q: How to contribute?**

Take a look at [here](./CONTRIBUTING.md).

## Homebrew Teresa Client

Run the following in your command-line:

```sh
$ brew tap luizalabs/teresa-cli
$ brew install teresa
```
