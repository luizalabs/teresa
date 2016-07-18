#Sensu Monitoring

## 1 - Pré-requisitos

### * Instalação Redis
`http://redis.io/topics/quickstart`

### * Instalation RabbitMQ
`http://www.rabbitmq.com/install-debian.html`

## 2 - Instalação Sensu

**Install Sensu using APT (recommended)** <br />
######Install the GPG public key: <br />
`wget -q http://sensu.global.ssl.fastly.net/apt/pubkey.gpg -O- | sudo apt-key add -`  <br />
######Create an APT configuration file at /etc/apt/sources.list.d/sensu.list:  <br />
`echo "deb     http://sensu.global.ssl.fastly.net/apt sensu main" | sudo tee /etc/apt/sources.list.d/sensu.list` <br />
######Update APT: <br />
`sudo apt-get update` <br />
######Install Sensu: <br />
`sudo apt-get install sensu` <br />

## 3 - Configuração do Sensu

`mkdir -p /etc/sensu/conf.d/sensu-server`

`/etc/sensu/conf.d/sensu-server/redis.json`
```
{
  "redis": {
    "host": "redis address",
    "port": 6379
  }
}
```
`/etc/sensu/conf.d/sensu-server/api.json`
```
{
  "api": {
    "host": "<api address>",
    "bind": "0.0.0.0",
    "port": 4567
  }
}
```


`/etc/sensu/conf.d/sensu-server/rabbitmp.json`
```
{
  "rabbitmq": {
    "host": "<rabbitmq  address>",
    "port": 5672,
    "vhost": "/sensu",
    "user": "sensu",
    "password": "secret"
  }
}
```

`/etc/sensu/conf.d/sensu-server/client.json`
```
{
  "client": {
    "name": "<sensu-server>",
    "address": "client  address",
    "subscriptions": [ "ALL" ]
  }
}
```
######Enable the Sensu client on system boot
`sudo update-rc.d sensu-client defaults`
######Enable the Sensu server and API to start on system boot
`sudo update-rc.d sensu-server defaults`
`sudo update-rc.d sensu-api defaults`



## 4 - Instalação do Sensu - Dashboard ( Uchiwa )

```
cd ~ 
wget http://dl.bintray.com/palourde/uchiwa/uchiwa_0.16.0-1_amd64.deb
dpkg -i uchiwa_0.16.0-1_amd64.deb
```
## 5 - Configuração do Sensu - Dashboard ( Uchiwa )

`/etc/sensu/uchiwa.json`

```
{
  "sensu": [
    {
      "name": "dc",
      "host": "<api server address>",
      "port": 4567
    }
  ],
  "dashboard": {
    "host": "0.0.0.0",
    "port": 3000
  }
}
```

```
sudo update-rc.d uchiwa defaults
/etc/init.d/uchiwa start
```

## 6 - Configuração do Sensu Handlers ( slack )

```
apt-get install build-essential
ln -s /opt/sensu/embedded/bin/gem /usr/bin/gem

sensu-install slack
cp /opt/sensu/embedded/bin/handler-slack* /etc/sensu/plugins/
```

`mkdir -p /etc/sensu/conf.d/handlers/`

`/etc/sensu/conf.d/handlers/slack_config.json`

```
{
    "handlers": {
        "slack": {
            "type": "pipe",
            "command": "/etc/sensu/plugins/handler-slack.rb",
            "severites": ["critical", "unknown"]
        }
    },
    "slack": {
        "webhook_url": "https://hooks.slack.com/services/<...>",
        "template" : ""
    }
}
```


## 7 - Instalação Sensu Client

**Install Sensu using APT (recommended)** <br />
######Install the GPG public key: <br />
`wget -q http://sensu.global.ssl.fastly.net/apt/pubkey.gpg -O- | sudo apt-key add -`  <br />
######Create an APT configuration file at /etc/apt/sources.list.d/sensu.list:  <br />
`echo "deb     http://sensu.global.ssl.fastly.net/apt sensu main" | sudo tee /etc/apt/sources.list.d/sensu.list` <br />
######Update APT: <br />
`sudo apt-get update` <br />
######Install Sensu: <br />
`sudo apt-get install sensu` <br />

## 8 - Configuração do Sensu client

`/etc/sensu/config.json`

```
{
  "rabbitmq": {
    "host": "<rabbitmq address>",
    "port": 5672,
    "vhost": "/sensu",
    "user": "sensu",
    "password": "<password>"
  }
}
```

`/etc/sensu/conf.d/client.json`

```
{
  "client": {
    "name": "<Client hostname>",
    "address": "<Client address>",
    "Environment": "<environment>",
    "subscriptions": [
      "<subscriptions>"
    ],
    "socket": {
      "bind": "127.0.0.1",
      "port": 3030
    },
    "keepalive": {
        "thresholds": {
            "warning": 30,
            "critical": 40
        },
        "handlers": ["<handler name>"],
        "refresh": 300
    }
  }
}
```

```
sudo update-rc.d sensu-client defaults
sudo update-rc.d sensu-api disable
sudo update-rc.d sensu-server disable
sudo /etc/init.d/sensu-client start

```

## 9 - Configuração kubernetes check  

######Obs.: Fazer processo no sensu-server e no kubernetes-master

```
sensu-install kubernetes
ls /opt/sensu/embedded/bin/check-kube-* -l
-rwxr-xr-x 1 root root 556 Jul 16 00:48 /opt/sensu/embedded/bin/check-kube-apiserver-available.rb
-rwxr-xr-x 1 root root 548 Jul 16 00:48 /opt/sensu/embedded/bin/check-kube-nodes-ready.rb
-rwxr-xr-x 1 root root 549 Jul 16 00:48 /opt/sensu/embedded/bin/check-kube-pods-pending.rb
-rwxr-xr-x 1 root root 549 Jul 16 00:48 /opt/sensu/embedded/bin/check-kube-pods-runtime.rb
-rwxr-xr-x 1 root root 554 Jul 16 00:48 /opt/sensu/embedded/bin/check-kube-service-available.rb
cp /opt/sensu/embedded/bin/check-kube-* /etc/sensu/plugins/
```

`mkdir -p /etc/sensu/conf.d/kube-master-staging`

######Exemplo:

`/etc/sensu/conf.d/kube-master-staging/kube_api_check.json`

```
{
  "checks": {
    "kube_api_stg": {
      "command": "/etc/sensu/plugins/check-kube-apiserver-available.rb --token <token> --api-server https://k8s-staging.a.luizalabs.com",
      "interval": 30,
      "handler": "<handler>",
      "refresh": 300,
      "subscribers": [ "<subscriber>" ]
    }
  }
}
```
