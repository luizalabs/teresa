#-*-coding:utf-8-*-

import requests
import json
import base64
from jinja2 import Template
from enum import Enum
import json
from attrdict import AttrDict
from slugify import slugify


class ContentType(Enum):
    json = 'application/json'
    yaml = 'application/yaml'
    json_patch = 'application/json-patch+json'
    merge_patch = 'application/merge-patch+json'
    strategic_merge_patch = 'application/strategic-merge-patch+json'


class Deployment(object):
    """
        Deployment
        http://kubernetes.io/docs/user-guide/deployments/
    """

    @staticmethod
    def create(namespace, name, app_name, containers, replicas=1):
        lst_containers = []

        for container in containers:
            c = {
                'name': slugify(container['name']),
                'image': container['image'].lower(),
                'ports': []
            }

            if isinstance(container['ports'], list):
                for port in container['ports']:
                    c['ports'].append({
                        'containerPort': port
                    })
            elif isinstance(container['ports'], dict):
                pass

            lst_containers.append(c)

        data = {
            'apiVersion': 'extensions/v1beta1',
            'kind': 'Deployment',
            'metadata': {
                'name': slugify(name)
            },
            'spec': {
                'replicas': replicas,
                'template': {
                    'metadata': {
                        'labels': {'app': slugify(app_name)}
                    },
                    'spec': {
                        'containers': lst_containers
                    }
                }
            }
        }

        # TODO: Validar e fazer a chamada para o método create
        return {
            'url': '/apis/extensions/v1beta1/namespaces/{}/deployments'.format(slugify(namespace)),
            'content_type': ContentType.json,
            'payload': json.dumps(data)
        }


class K8sClient(object):
    """docstring for K8sClient"""
    def __init__(self, base_url, user, password, ssl_verify=True):
        self.base_url = base_url
        self.user = user
        self.password = password
        self.ssl_verify = ssl_verify
        self.headers = {
            'authorization': 'Basic {}'.format(base64.b64encode('{}:{}'.format(user, password))),
            'cache-control': 'no-cache'
            # 'content-type': 'application/json',
        }

    def request(self, verb, url, content_type, payload):
        print '--> do http request'

        headers = self.headers.copy()
        headers['content-type'] = content_type.value

        url = self.base_url + url

        # print headers
        # print payload
        # print url

        response = requests.request(verb, url, data=payload, headers=headers, verify=self.ssl_verify)
        print '--> http status code: %s' % response.status_code

        # TODO: Validar erro aqui... se tiver diferente de 200(1), raise exception foo

        return AttrDict(json.loads(response.text))

    def create(self, resource, namespace, **kwargs):
        re = resource.create(namespace, **kwargs)
        return self.request('POST', **re)




# kubernetes config
k8s = K8sClient(base_url='https://52.9.192.117', user='admin', password='3J4oFyDXfI3VQhJ5', ssl_verify=False)

# mesmo efeito que um "$kubectl run"
deployment = k8s.create(Deployment,
    namespace='default',
    name='DeplOyment Meu app      nginx',
    app_name='meu app nginx',
    replicas=2,
    # containers=[
    #     {'name':'meu nome', 'image':'minhaimagem', 'port':80},
    #     {'name':'meu nome', 'image':'minhaimagem', 'ports':[80, 8080]},
    #     {'name':'meu nome 2', 'image':'minhaimagem 2', 'ports':{'tcp': [80, 8080], 'udp':8081}}
    # ]
    containers=[
        {'name':'meu container nginx', 'image':'nginx;latest', 'ports':[8080]}
    ]
)

print deployment








# def http_request(url, verb='GET', payload=None, contentType=None):
#     print '--> do http request'

#     if contentType != None:
#         headers['content-type'] = contentType

#     response = requests.request(verb, url, data=payload, headers=headers, verify=False)
#     print '--> http status code: %s' % response.status_code
#     return response

# def create_deployment(namespace, name, replicas, labels_app, container_name, container_image, container_port):
#     """
#     Cria um novo deployment

#     http://kubernetes.io/docs/user-guide/deployments/
#     """
#     url = base_url + '/apis/extensions/v1beta1/namespaces/%s/deployments' % namespace

#     template = Template(open('templates/deployment.yaml').read())

#     payload = template.render(name=name, 
#         replicas=replicas, labels_app=labels_app, container_name=container_name,
#         container_image=container_image, container_port=container_port)

#     response = http_request(url, 'POST', payload)

#     # TODO: Verificar status code e emitir erro
#     # response.status_code = 409 -> Quando já existe o deployment
#     # print response.status_code

#     return json.loads(response.text)

# def get_deployments(namespace):
#     """
#     Retorna uma lista de deployments do namespace
#     """
#     url = base_url + '/apis/extensions/v1beta1/namespaces/%s/deployments' % namespace

#     response = http_request(url)

#     # TODO: check http status
#     # response.status_code
#     return json.loads(response.text)


# def get_deployment(namespace, name):
#     """
#     Retorna o deployment pelo nome
#     """
#     url = base_url + '/apis/extensions/v1beta1/namespaces/%s/deployments/%s' % (namespace, name)

#     response = http_request(url)

#     # TODO: check http status
#     # response.status_code
#     return json.loads(response.text)

# def get_deployment_scale(namespace, name):
#     """
#     Retorna o scale do deployment (informações referente a replicas)
#     """
#     url = base_url + '/apis/extensions/v1beta1/namespaces/%s/deployments/%s/scale' % (namespace, name)

#     response = http_request(url)

#     return json.loads(response.text)

# def update_deployment_replicas(namespace, name, replicas):
#     """
#     Atualiza a quantidade de replicas de um deploy
#     """
#     url = base_url + u'/apis/extensions/v1beta1/namespaces/{}/deployments/{}'.format(namespace, name)

#     payload = {
#         "spec": {
#             "replicas": replicas
#         }
#     }

#     response = http_request(url, 'PATCH', json.dumps(payload), 'application/merge-patch+json')

#     return json.loads(response.text)


# def create_service(namespace, name, selector_app_name, port, target_port):
#     """
#     Cria um novo serviço.
#     O parâmetro selector_app_name é referente ao label que será utilizado para buscar o deployment ou pod que será exposto
#     """

#     url = base_url + u'/api/v1/namespaces/{}/services'.format(namespace)

#     template = Template(open('templates/service.yaml').read())

#     # TODO: http://kubernetes.io/docs/user-guide/services/#multi-port-services

#     payload = template.render(
#         name=name, selector_app_name=selector_app_name,
#         port=port, target_port=target_port
#     )

#     response = http_request(url, 'POST', payload)

#     return json.loads(response.text)


# ------------------------------

# Cria deployment
# d = create_deployment(
#         namespace='default',
#         name='nginx-teste',
#         replicas=1,
#         labels_app='nginx-label',
#         container_name='nginx',
#         container_image='nginx:1.7.9',
#         container_port=80)
# print d

# Get dos deployments
# print get_deployments('default')

# print get_deployment('default', 'nginx-teste')

# print get_deployment_scale('default', 'nginx-teste')

# print update_deployment_replicas('default', 'nginx-teste', 1)

# print create_service('default', 'nginx-service-teste', 'nginx-label', 8081, 80)


