#-*-coding:utf-8-*-

from flask import Flask, request, stream_with_context, Response
from git import Repo
import docker
from io import BytesIO
import json

import subprocess

import k8s_client

app = Flask(__name__)
log = app.logger

# Criando client docker
tls_config = docker.tls.TLSConfig(
    verify = False,
    client_cert=('/Users/eder/.docker/machine/machines/dkr-one/cert.pem', '/Users/eder/.docker/machine/machines/dkr-one/key.pem')
)
docker_client = docker.Client(base_url='tcp://192.168.99.100:2376', tls=tls_config)


def clone_repo(src_code, dest_dir):
    Repo.clone_from(src_code, dest_dir)
    return True

@app.route("/health-check")
def healtcheck():
    return "Up"

@app.route("/stream-test")
def stream_test():
    def generate():
        for i in xrange(10,100):
            yield '>>' + str(i)  + '\n'
        # yield 'Hello '
        # yield 'asdasdasd'
        # yield '!'
    return Response(stream_with_context(generate()))


@app.route('/apps/deploys', methods=['POST'])
def app_deploy_deprecated():
    """
    DEPRECATED
    """
    # print request.form['appName']
    # print request.form['src']
    # git_src =  'https://github.com/ederavilaprado/paas-app-example-python-flask.git'
    # # TODO: gerar este randômico
    git_temp_repo = '/Users/eder/Desktop/temp_git/001/'
    docker_image_tag = 'gcr.io/k8shelloworld/hello-flask:v2'
    # docker_image_tag = '127.0.0.1:5000/hello-flask:v1.0'

    # dockerfile = '''
    # FROM tutum/buildstep
    # EXPOSE 8080
    # CMD ["python", "app.py"]
    # '''
    # dfile = BytesIO(dockerfile.encode('utf-8'))

    def generate():
        # TODO: get do diretório

        # **** Build do projeto
        yield json.dumps({'step': 1, 'message': 'Building project'}) + '\n'

        # TODO: injetar o docker_file para dentro do projeto. Criar um template jinja e substituir os valores

        # https://docker-py.readthedocs.org/en/latest/api/#build
        for line in docker_client.build(path=git_temp_repo, tag=docker_image_tag):
            yield line

        # **** Push da imagem
        yield json.dumps({'step': 2, 'message': 'Pushing image'}) + '\n'

        # Modelo de push convencional, porem não funciona com o gcloud
        # for line in docker_client.push(docker_image_tag, stream=True):
        #     yield line
        process = subprocess.Popen(['gcloud', 'docker', 'push', docker_image_tag], stdout=subprocess.PIPE)
        for line in iter(process.stdout.readline, ''):
            yield line

        # **** Run
        yield json.dumps({'step': 3, 'message': 'Running'}) + '\n'
        process = subprocess.Popen(['kubectl', 'run', 'hello-flask', '--image=' + docker_image_tag, '--port=8080'], stdout=subprocess.PIPE)
        for line in iter(process.stdout.readline, ''):
            yield line
        
        # **** Expose deployment
        yield json.dumps({'step': 4, 'message': 'Creating service to expose the pods'}) + '\n'

        process = subprocess.Popen(['kubectl', 'expose', 'deployment', 'hello-flask', '--type=LoadBalancer'], stdout=subprocess.PIPE)
        for line in iter(process.stdout.readline, ''):
            yield line

        yield json.dumps({'step': 5, 'message': 'Deploy finished'})

    return Response(stream_with_context(generate()))


# TODO: mudar para POST apenas
# @app.route('/teams/<team_name>/aplications', methods=['POST'])
@app.route('/teams/<team_name>/apps', methods=['POST', 'GET'])
def app_deploy(team_name):

    log.debug('Iniciando deploy...')

    git_temp_repo = '/Users/eder/Desktop/temp_git/001/'
    docker_image_tag = 'gcr.io/k8shelloworld/hello-flask:v2'

    # TODO: git clone + injetar o docker_file utilizando jinja como template base

    # for line in docker_client.build(path=git_temp_repo, tag=docker_image_tag):
    #     log.debug(line)

    # TODO: validar se foi com sucesso o build

    log.debug('Deploy finalizado... iniciando o envio da imagem para o registry')

    # Modelo de push convencional, porem não funciona com o gcloud
    # for line in docker_client.push(docker_image_tag, stream=True):
    #     yield line
    # process = subprocess.Popen(['gcloud', 'docker', 'push', docker_image_tag], stdout=subprocess.PIPE)
    # for line in iter(process.stdout.readline, ''):
    #     log.debug(line)

    log.debug('Imagem enviada para o registry... criando deployment...')

    deployment = k8s_client.create_deployment(
        namespace=team_name,
        name='meu-app-flask',
        replicas=1,
        labels_app='meu-app-flask-label',
        container_name='meu-app-flask',
        container_image=docker_image_tag,
        container_port=8080)

    
    # TODO: buscar pela porta correta do container (é um array)
    service = k8s_client.create_service(
        namespace=team_name,
        name='servico-do-meu-app-flask',
        selector_app_name='meu-app-flask-label',
        port=80,
        target_port=deployment['spec']['template']['spec']['containers'][0]['ports'][0]['containerPort'])

    # TODO: retornar http_status 201 com json
    return 'Foi'



if __name__ == "__main__":
    app.run(host='0.0.0.0', debug = True)
