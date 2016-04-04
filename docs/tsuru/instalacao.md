# Instalação

## Mongo

```bash
$ sudo vi /etc/apt/apt.conf
Acquire::http::proxy "http://10.152.112.13:54052/";
Acquire::https::proxy "https://10.152.112.13:54052/";

$ export https_proxy=http://10.152.112.13:54052
$ export http_proxy=http://10.152.112.13:54052

$ sudo apt-key adv --keyserver-options http-proxy=http://10.152.112.13:54052/ --keyserver  hkp://keyserver.ubuntu.com:80 --recv 7F0CEB10
$ apt-get update

$ ./autodisk.sh VG01 /dev/sdb VG01 lvol01 /var/lib/mongodb
$ chown mongodb:mongodb /var/lib/mongodb/

$ apt-get install 
$ sudo apt-get install -y mongodb-org
$ service mongod start

$ vi /etc/mongod.conf #alterar ip do bind
mongo 10.152.112.105
```

## Redis
```
# Master
$ adduser adm-redis
$ gpasswd -a adm-redis sudo
$ sudo su - adm-redis
$ sudo vi /etc/apt/apt.conf
Acquire::http::proxy "http://10.152.112.13:54052/";
Acquire::https::proxy "https://10.152.112.13:54052/";

$ export https_proxy=http://10.152.112.13:54052
$ export http_proxy=http://10.152.112.13:54052
$ sudo -E add-apt-repository ppa:chris-lea/redis-server
$ apt-get update
$ sudo apt-get install redis-server
$ redis-benchmark -q -n 1000 -c 10 -P 5
```

## Gandalf (git) e Archive Server

**Gandalf**  
```bash
$ sudo vi /etc/apt/apt.conf
Acquire::http::proxy "http://10.152.112.13:54052/";
Acquire::https::proxy "https://10.152.112.13:54052/";

$ export https_proxy=http://10.152.112.13:54052
$ export http_proxy=http://10.152.112.13:54052
```

**Archive Server**
```bash
$ vi /etc/default/archive-server
READ_HTTP=0.0.0.0:3232
WRITE_HTTP=127.0.0.1:3131
ARCHIVE_DIR=/var/lib/archive-server/archives
MONGODB_SERVER=tsuru-mongo-01.nuvemluiza.intranet:27017
MONGODB_DB_NAME=archives
export ARCHIVE_SERVER_OPTS="-dir=${ARCHIVE_DIR} -read-http=${READ_HTTP} -write-http=${WRITE_HTTP} -mongodb=${MONGODB_SERVER} -dbname=${MONGODB_DB_NAME}"
```
