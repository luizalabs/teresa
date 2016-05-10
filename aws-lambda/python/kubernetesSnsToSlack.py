from __future__ import print_function

import json
import logging

from base64 import b64decode
from urllib2 import Request, urlopen, URLError, HTTPError


notifications = [
    {
        'topic_arn': 'arn:aws:sns:us-east-1:518863443564:kubernetes-dev-events',
        'slack_channel': 'kubernetes-dev',
        'hook_url': 'https://hooks.slack.com/services/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'
    }
]

logger = logging.getLogger()
logger.setLevel(logging.INFO)

def lambda_handler(event, context):
    logger.info("Event: " + str(event))
    message = json.loads(event['Records'][0]['Sns']['Message'])
    logger.info("Message: " + str(message))

    topic_arn = event['Records'][0]['Sns']['TopicArn']

    for notification in notifications:
        if notification['topic_arn'] != topic_arn:
            continue
        
        slack_message = {
            'channel': notification['slack_channel'],
            'text': event['Records'][0]['Sns']['Message']
        }

        req = Request(notification['hook_url'], json.dumps(slack_message))
        try:
            response = urlopen(req)
            response.read()
            logger.info("Message posted to %s", slack_message['channel'])
        except HTTPError as e:
            logger.error("Request failed: %d %s", e.code, e.reason)
        except URLError as e:
            logger.error("Server connection failed: %s", e.reason)
