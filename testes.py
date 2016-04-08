#-*-coding:utf-8-*-

import yaml
import json
from attrdict import AttrDict


try:
    from yaml import CLoader as Loader, CDumper as Dumper
except ImportError:
    from yaml import Loader, Dumper

def yaml_dump(data):
    return yaml.dump(data, Dumper=Dumper, default_flow_style=False, allow_unicode=True)

def yaml_load(data):
    return yaml.load(data, Loader=Loader)


class Hero:
    def __init__(self, name=None, hp=None, sp=None):
        self.name = name
        self.hp = hp
        self.sp = sp

    def append_container(self, container):
        # if self.containers:
        self.containers = []

        self.containers.append(container)


h = Hero("Galain Ysseleg", hp=-3, sp=2)
h.append_container('x')

import jsonpickle
# frozen = jsonpickle.encode(h, unpicklable=False)

# print frozen

# print 'json encode'
# print frozen

# json_loaded = json.loads(frozen)


# print 'Json Loaded'
# print json_loaded

# print yaml_dump(json.loads(frozen))

# # print yaml_dump(h)

# print h.__dict__

# yaml.safe_dump(h)


# weird_json = '{"x": 1, "x": 2, "x": 3}'
# d = json.loads(weird_json)
# print d['x']

# class Container(object):
#     """Container"""
#     def __init__(self, name, image, port):
#         self.name = name
#         self.image = image
#         self.port = port

class Deployment(object):
    """Deployment"""
    def __init__():
        pass

    @staticmethod
    def create(name, app_name, containers, replicas=1):
        
        lst_containers = []

        for container in containers:
            c = {
                'name': container['name'],
                'image': container['image'],
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
                'name': name
            },
            'spec': {
                'replicas': replicas,
                'template': {
                    'metadata': {
                        'labels': [{'app': app_name}]
                    },
                    'spec': {
                        'containers': lst_containers
                    }
                }
            }
        }

        # TODO: Validar e fazer a chamada para o método create

        print json.dumps(data)

    @staticmethod
    def get(j):
        return AttrDict(json.loads(j))

Deployment.create(
    name='meu deploy v1',
    app_name='meu app',
    replicas=4,
    # containers=[
    #     {'name':'meu nome', 'image':'minhaimagem', 'port':80},
    #     {'name':'meu nome', 'image':'minhaimagem', 'ports':[80, 8080]},
    #     {'name':'meu nome 2', 'image':'minhaimagem 2', 'ports':{'tcp': [80, 8080], 'udp':8081}}
    # ]
    containers=[{'name':'meu nome', 'image':'minhaimagem', 'ports':[80]}]
    )

# "containerPort": 8080,
# "protocol": "TCP"

# di = Deployment.get(j)



# print di.kind
