# Contributing to Teresa

The Teresa project welcomes contributions from all developers. The high level
process for development matches many other open source projects. See below for
an outline:

* Fork this repository.
* Make your changes. Format your code with `gofmt` or `goimports` and try to
  follow the style of the surrounding code.
* Add an entry to the CHANGELOG.
* [Submit a pull request][prs] (PR) to this repository with your changes, and
  unit tests whenever possible.
* If your PR fixes any [issues][issues], make sure you write Fixes #1234 in
  your PR description (where #1234 is the number of the issue you're closing).
* The Teresa core contributors will review your code. After each of them sign
  off on your code, they'll comment your PR with `LGTM`. Once that happens, the
  contributors will merge it.

# Tips on Testing

You basically need a Kubernetes cluster (with the correponding kubectl config)
and a S3 API compatible server.

Compile the server and client with you modifications:

    $ make build-client build-server

Start the teresa-server (in this case we are using S3):

    $ export TERESA_STORAGE_AWS_KEY=xxxx
    $ export TERESA_STORAGE_AWS_SECRET=xxxx
    $ export TERESA_STORAGE_AWS_REGION=us-east-1
    $ export TERESA_STORAGE_AWS_BUCKET=bucket
    $ export TERESA_K8S_CONFIG_FILE=~/.kube/config
    $ export TERESA_SECRETS_PRIVATE_KEY=$(pwd)/pkg/server/secrets/testdata/fake.rsa
    $ export TERESA_SECRETS_PUBLIC_KEY=$(pwd)/pkg/server/secrets/testdata/fake.rsa.pub
    $ ./teresa-server run

Configure the client:

    $ ./teresa config set-cluster testing --server localhost --current

Now you are ready to use the modified client and server with effects on your
kubectl current context.
