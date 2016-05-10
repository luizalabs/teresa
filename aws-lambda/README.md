## Criar Monitoramento EC2 e Autoscaling Kubernetes integrado com Slack ##

**1 - SNS - Crie Topic SNS**
 
**2 - Lambda - Clique em Create new Function**

- Nome --> KubernetesSnsToSlack
- Description --> Envio de evento EC2 e Autoscaling para Slack
- Runtime --> Nodejs 4.3
- Handler --> index.handler
- Role--> iam-role-lambda-sns-to-slack-alerts
- VPC --> No Vpc

**3 - Lambda - Abra a função criada etapa 2 **
-  Vá para Event Sources
- Clique em Add event source
- Event source type selecione SNS
- SNS Topic selecione topic criada na etapa 1

**4 - CloudWatch -  ** Selecione Events  -- > EC2**
- Clique em Create Roles
- Select Event Source --> EC2 instance state change Notification
- Specific state(s) --> Marque os tipos de eventos que deseja monitorar (stopped, Running, Terminated, etc)
- Specific instance(s) --> Marque instance que quer monitorar
- Clique Add target
- Select Target type --> SNS Topic
- Topic --> Selecione a Topic criada na etapa 1

**5 - CloudWatch -  ** Selecione Events --> Autoscaling**
- Clique em Create Roles
- Select Event Source --> Autoscaling
- Specific state(s) --> Marque os tipos de eventos que deseja monitorar ( "EC2 Instance Launch Successful", etc)
- Specific group name(s) --> Marque AutoScaling Group que quer monitorar
- Clique Add target
- Select Target type --> SNS Topic
- Topic --> Selecione a Topic criada na etapa 1



