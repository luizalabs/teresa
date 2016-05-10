## Criar Monitoramento EC2 d ELB Kubernetes integrado com Slack ##

**1 - SNS - Criar Topic SNS**
 
**2 - Lambda - Create new Function**

- Nome --> KubernetesSnsToSlack
- Description --> Envio de evento EC2 e ELB para Slack
- Runtime --> Nodejs 4.3
- Handler --> index.handler
- Role--> iam-role-lambda-sns-to-slack-alerts
- VPC --> No Vpc

**3 - Lambda - Abra a função criada etapa 2 **
-  Vá para Event Sources
- Clique em Add event source
- Event source type selecione SNS
- SNS Topic selecione topic criada na etapa 1



