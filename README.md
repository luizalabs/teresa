# Teresa

Teresa is an extremely simple platform as a service that runs on top of Kubernetes.

To have a full Teresa setup you'll need the API running as a POD on Kubernetes and the CLI. Steps to setup a new cluster and deploy a new application, assuming you already have the API running:

    1. teresa config set-cluster cluster_name --server https://mycluster.com
    2. teresa config use-cluster cluster_name
    3. teresa login --user myuser@mydomain.com
    4. teresa create app myapp --team dev --scale 2
    5. teresa deploy /path/to/myapp --app myapp --team dev --description "release 1.2"


## View API documentation

  To view the auto-generated swagger API documentation, the following command will compile, run a webserver and open your browser on the swagger-ui:

    cd api; make swagger-docs
