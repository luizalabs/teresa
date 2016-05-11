## Criar Monitoramento EC2 e Autoscaling Kubernetes integrado com Slack ##

**1 - SNS - Crie Topic SNS**
 
**2 - Lambda - Clique em Create new Function**

- Nome --> KubernetesSnsToSlack
- Description --> Envio de evento EC2 e Autoscaling para Slack
- Runtime --> Nodejs 4.3
- Handler --> index.handler
- Role--> iam-role-lambda-sns-to-slack-alerts ( Nesta role precisa ter as policies pol-aws-lambda-slack e  AmazonEC2ReadOnlyAccess, listadas abaixo:

**Policy - pol-aws-lambda-slack**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*"
    }
  ]
}
```
**Policy - AmazonEC2ReadOnlyAccess**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "ec2:Describe*",
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": "elasticloadbalancing:Describe*",
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "cloudwatch:ListMetrics",
        "cloudwatch:GetMetricStatistics",
        "cloudwatch:Describe*"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": "autoscaling:Describe*",
      "Resource": "*"
    }
  ]
}
```

- VPC --> No Vpc
- Code entry type --> 
  Edit code inline coloque o codigo --> (https://github.com/luizalabs/paas/blob/master/aws-lambda/nodejs/kubernetesSnsToSlack.js)


**3 - Lambda - Abra a função criada etapa 2**
-  Vá para Event Sources
- Clique em Add event source
- Event source type selecione SNS
- SNS Topic selecione topic criada na etapa 1

**4 - CloudWatch - Selecione Events  -- > EC2**
- Clique em Create Roles
- Select Event Source --> EC2 instance state change Notification
- Specific state(s) --> Marque os tipos de eventos que deseja monitorar (stopped, Running, Terminated, etc)
- Specific instance(s) --> Marque instance que quer monitorar
- Clique Add target
- Select Target type --> SNS Topic
- Topic --> Selecione a Topic criada na etapa 1

**5 - CloudWatch -   Selecione Events --> Autoscaling**
- Clique em Create Roles
- Select Event Source --> Autoscaling
- Specific state(s) --> Marque os tipos de eventos que deseja monitorar ( "EC2 Instance Launch Successful", etc)
- Specific group name(s) --> Marque AutoScaling Group que quer monitorar
- Clique Add target
- Select Target type --> SNS Topic
- Topic --> Selecione a Topic criada na etapa 1



