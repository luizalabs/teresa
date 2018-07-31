# Teresa FAQ

## Basics

**Q: What components do I need?**

Teresa uses a client-server model. The server is installed and managed by the
Kubernetes administrators. Most people only needs the client to do day-to-day
tasks.

**Q: Just got the client, what now?**

The teresa administrators will provide a server endpoint and create an account
on a team. After configuring the client and logging in you will be able to do
all operations on the team's applications.

**Q: How to configure the client?**

You have to choose a name for the cluster (environment), for example:

    $ teresa config set-cluster production --server <teresa-endpoint>

**Q: Does teresa support TLS?**

Yes, but it is optional, ask the administrators and in positive case use the
`--tls` flag to configure the client:

    $ teresa config set-cluster <cluster-name> --server <teresa-endpoint> --tls

**Q: How to login?**

    $ teresa login --user <user-email>

**Q: How to change the temporary password created by the administrator?**

    $ teresa set-password

**Q: How to add/update a cluster?**

Use the `set-cluster` command again:

    $ teresa config set-cluster <cluster-name> --server <teresa-endpoint>

**Q: How to start using another cluster?**

    $ teresa config use-cluster <cluster-name>
    $ teresa login --user <user-email>

**Q: Do I always have to login?**

When you login you get a token that by default is valid por 15 days, so you are
only supposed to login again after the token expires. You can customize the
time interval by using the  `--expires-in ` flag:

     $ teresa login --user <user-email> --expires-in 720h

**Q: How to list the available clusters?**

By viewing the full client configuration:

    $ teresa config view

**Q: How to get more help?**

The client has a builtin thorough help:

    $ teresa --help

## Applications

**Q: What is the usual flow for deploying a new app?**

First you have to create it:

    $ teresa app create <app-name> --team <team-name>

After that you usually have to set some environment variables:

    $ teresa app env-set KEY=VALUE --app <app-name>

At last you deploy (assuming you are at the app root directory):

    $ teresa deploy create . --app <app-name> --description 'v1.0.0'

**Q: How to just deploy an app?**

    $ teresa deploy create /path/to/project --app <app-name> --description "v1.0.0"

**Q: Do I have to build the app first? What languages are supported?**

No, teresa uses the Heroku
[buildpacks](https://devcenter.heroku.com/articles/buildpacks) to build the
app. The list of supported languages is located [here](https://devcenter.heroku.com/articles/buildpacks#officially-supported-buildpacks).

**Q: Do I have to provide an init script or supervisor configuration?**

No, teresa uses a [Procfile](https://devcenter.heroku.com/articles/procfile) to
describe the launch process. In short, you create the Procfile in the root
directory of the app, for example a django app:

    web: uwsgi --http :$PORT --processes 2 --master --pythonpath django --module app.wsgi

`web` is the `process type` and it is the default for new apps.  
**It is important to always listen to the port specified in the environment
variable PORT.**

**Q: Can I execute more than one process per app?**

No, it's one process per app.

**Q: I have many apps sharing the same repository, how to proceed?**

For each app add a Procfile entry respecting the restrictions:

* process types can't be repeated
* `web`, `release` and `cron*` process types are special: `web` is the only
  process type that creates an external endpoint, `cron*` is used for cronjobs
  and `release` for one off tasks after deploys.

When creating the apps use the flag `process-type` to select the appropriate
entry. For example, consider the Procfile:

    web: app1
    worker: app2

Apps creation:

    $ teresa app create app1 --team <team-name>
    $ teresa app create app2 --process-type worker --team <team-name>

There's nothing special with the name `worker`, it can be anything different
from `web` with a matching `Procfile` line. The two apps are deployed
separately (assuming you are at the shared repo root directory):

    $ teresa deploy create . --app app1 --description 'v1.0.0'
    $ teresa deploy create . --app app2 --description 'v2.0.0'

**Q: How to get info about an app?**

    $ teresa app info <app-name>

**Q: How to get the list of apps I have access to?**

    $ teresa app list

**Q: How to get app logs?**

    $ teresa app logs <app-name>

**Q: How to set an environment variable?**

    $ teresa app env-set KEY=VALUE --app <app-name>

Note: it will implicitly create a new deploy.

**Q: How to unset an environment variable?**

    $ teresa app env-unset KEY --app <app-name>

Note: it will implicitly create a new deploy.

**Q: How to create an app without an endpoint (a worker for example)?**

Use any name different from `web` as the process type:

    $ teresa app create <app-name> --team <team-name> --process-type worker

You also have to adjust the `Procfile` to have a corresponding `worker` key.

**Q: How to create an app with an internal endpoint only?**

    $ teresa app create <app-name> --team <team-name> --internal

The app is only visible inside the cluster. The flag `--internal` only works
with the `web` process type (the default).

**Q: Where can I get some hello world examples?**

Check out some examples [here](https://github.com/luizalabs/hello-teresa). 

**Q: How to run an app locally?**

You can perform the build and run your application locally with [Docker](https://docker.com),
to do that create a `Dockerfile` in the root directory of your application, like this:

```
FROM luizalabs/slugbuilder:v3.4.0 AS builder
COPY . /tmp/app
RUN /builder/build.sh

FROM luizalabs/slugrunner:v3.3.0
COPY --from=builder /tmp/slug.tgz /slug/slug.tgz
ENV SLUG_DIR /slug
EXPOSE 5000
ENTRYPOINT ["/runner/init"]
```

Use the command `docker build` to create a Docker image:

    $ docker build . -t runlocal --no-cache

And then use the command `docker run` to start your application:

    $ docker run -p 5000:5000 -e PORT=5000 runlocal start web

> The last argument is relative of the Procfile key of your application.

## Advanced Topics

**Q: How to set up Kubernetes health checks?**

Take a look at [here](https://github.com/luizalabs/hello-teresa#teresayaml).

**Q: I need one `teresa.yaml` per process type, how to proceed?**

If a file named `teresa-processtype.yaml` is found it is used instead of
`teresa.yaml`.

**Q: Can I use one teresa.yaml for all my apps?**

Yes, from version 0.24 we support a new format `v2` where each app has its own
section, e.g.:

```yaml
version: v2
applications:
  app1:
    lifecycle:
      preStop:
        drainTimeoutSeconds: 15
  app2:
    lifecycle:
      preStop:
        drainTimeoutSeconds: 20
```

The file `teresa-processtype.yaml` is still processed.

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

**Q: How to use nginx in front of my app (as a sidecar)?**

Create a nginx.conf file in the project root directory, for example:

```
events {
  worker_connections  1024;
}

http{
  server {
    listen $NGINX_PORT;
    location / {
        proxy_set_header HOST $host;
        proxy_pass $NGINX_BACKEND;
    }
    
    server_tokens off;
  }
}
```

Note that this must be the **full** nginx configuration and the environment
variables will be automatically filled.

**Q: How to use nginx to serve static files?**

When you activate nginx sidecar Teresa will share the `/app` ($HOME of apps) directory
with nginx container, so you can use a simple _location_ rule on your nginx.conf

```
    location /static {
        alias /app/static/;
    }
```

**Q: How to use unix socket in communication between nginx and app**

Just create an _upstream_ pointing to an _unix socket_ on the shared directory (`/app`) and
set the `proxy_pass` to previous created upstream, for instance:
```
events {
  worker_connections  1024;
}

http{
  upstream myapp {
      server unix:/app/myapp.sock;
  }
  server {
    listen $NGINX_PORT;
    location / {
        proxy_set_header HOST $host;
        proxy_pass http://myapp;
    }
  }
}
```

**Q: How to get nginx logs?**

Filter the logs by container name:

    $ teresa app logs <app-name> --container nginx

**Q: How to build a dynamic nginx configuration based on the app's env vars?**

By default teresa uses a nginx image with perl support, so an example
configuration with a conditional redirect would be:

```
load_module modules/ngx_http_perl_module-debug.so;

env MY_VAR;

events {
    worker_connections 1024;
}

http {
    perl_set $my_var 'sub { return $ENV{"MY_VAR"}; }';
    server {
        listen $NGINX_PORT;
        if ($my_var = "my_value") {
            return 301 https://mydomain;
        }
        location / {
            proxy_set_header HOST $host;
            proxy_pass $NGINX_BACKEND;
        }
    }
}
```

**Q: How to enable ssl?**

For now we only support AWS and you need the certificate ARN:

    $ teresa service enable-ssl --app <app-name> --cert arn:aws:iam::xxxxx:server-certificate/cert-name

To only use SSL pass the flag `--only`.

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

**Q: How to configure the cloud provider firewall?**

If your app uses the cloud provider load balancer (this is the default) you can
set up a whitelist:

   $ teresa service whitelist-source-ranges myapp x.x.x.0/24 y.y.y.0/24

Only these ranges can access the service. To remove the whitelist just pass no
ranges. You can view the current whitelist with:

   $ teresa service info myapp

**Q: How to use Cloud SQL on GCP?**

You have to use the [Cloud SQL Proxy](https://cloud.google.com/sql/docs/mysql/sql-proxy).
First, you need to create a [service account](https://cloud.google.com/video-intelligence/docs/common/auth#set_up_a_service_account)
with Cloud SQL permissions and generate a credentials file. After that, create
a secret with the contents of this file:

   $ teresa app secret-set --app myapp -f credentials.json

Migrate to the teresa yaml configuration format v2 and add a section
describing the proxy configuration:

```yaml
version: v2
applications:
  myapp:
    sidecars:
      cloudsql-proxy:
        instances: project:zone:name=tcp:3306
        credentialFile: credentials.json
```

After deploying the app the instance will be available on `localhost` using port
`3306`. The Cloud SQL Proxy supports both MySQL and PostgreSQL.

## Administration

**Q: How does access control work?**

User are identified by email and can belong to teams, which are just sets of
users. An application belongs to a team and all team users can perform all
operations on it. Users can't see apps that belong to teams they don't belong
to. There are administrative users, which are users with an admin flag set and
only them can do user and team management.

**Q: How to create a team?**

    $ teresa team create <team-name> --email <team-email>

**Q: How to create a user?**

    $ teresa create user --name <user-name> --email <user-email> --password <user-password>

**Q: How to add a user to a team?**

    $ teresa team add-user --team <team-name> --user <user-email>

**Q: How to remove a user from team?**

    $ teresa team remove-user --team <team-name> --user <user-email>

**Q: How to delete an user?**

    $ teresa delete user --email <user-email>

**Q: How to create a new admin user?**

You need access to the environment where the Teresa server is running (often a
Kubernetes pod):

    $ teresa-server create-super-user --email <admin-email> --password <admin-password>

**Q: How to change the app team?**

You need to be an admin to change the team:

    $ teresa-server app change-team <app-name> <team-name>

## Development

**Q: How to contribute?**

Take a look at [here](./CONTRIBUTING.md).
