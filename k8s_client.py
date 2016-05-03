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
        headers['content-type'] = content_type

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
        {'name':'meu container nginx', 'image':'nginx:latest', 'ports':[8080]}
    ]
)

print deployment
