#-*-coding:utf-8-*-
# from enum import Enum

# class ResourceKind(Enum):
#     deployment = 'deployment'
#     service = 'service'

class Deployment(object):
    """docstring for Deployment"""
    def __init__(self):
        pass

    @staticmethod
    def create(name, app_name, containers, replicas=1):
        return 'Aeeeeeee'


class K8sClient(object):
    """Deployment"""
    def __init__(self, config):
        self.config = config

    def request(self, url, payload):
        print 'chegou aqui'

    def create(self, resource, **kwargs):
        print kwargs['replicas']
        


k8s = K8sClient({
    'host': '0.0.0.0'
})

# k8s.create(Deployment,
#     name='meu deploy v1',
#     app_name='meu app',
#     replicas=4,
#     # containers=[
#     #     {'name':'meu nome', 'image':'minhaimagem', 'port':80},
#     #     {'name':'meu nome', 'image':'minhaimagem', 'ports':[80, 8080]},
#     #     {'name':'meu nome 2', 'image':'minhaimagem 2', 'ports':{'tcp': [80, 8080], 'udp':8081}}
#     # ]
#     containers=[{'name':'meu nome', 'image':'minhaimagem', 'ports':[80]}]
# )

# from enum import Enum

# class Animal(Enum):
#     ant = 1
#     bee = 2
#     cat = 3
#     dog = 4

# print Animal.dog.value

from slugify import slugify

a = 'DeploymenT meu app flask staging'

print slugify(a)








