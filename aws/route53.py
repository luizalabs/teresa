import boto
import os
from boto.route53.connection import Route53Connection
from boto.route53.record import ResourceRecordSets

conn = Route53Connection(
    os.environ.get('AWS_ACCESS_KEY_ID'),
    os.environ.get('AWS_SECRET_KEY_ID'))

#criar fora das function?
zone = conn.get_zone("luizalabs.com")

#Verificar Convencoes de nomes (function,variables, etc)

def  recordset(action,fqdn,record_set_type,target):
    changes = ResourceRecordSets(conn, zone.id)
    change = changes.add_change(action, fqdn, record_set_type, ttl=60, identifier='unique')
    change.add_value('target')
    result = changes.commit()
    return result


# variables tests

action = 'CREATE'
fqdn = 'weavescope-tst2.staging.luizalabs.com'
target = 'dualstack.a3615e6c417be11e68c7a0a785937962-961467740.us-east-1.elb.amazonaws.com'
record_set_type = 'CNAME' # ELB aceita somente CNAME

result = recordset(action,fqdn,record_set_type,target)

# TODO fazer tratamento de erros

