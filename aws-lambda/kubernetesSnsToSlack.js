const https = require('https');
const url = require('url');
const AWS = require('aws-sdk');
var slack_req_opts;

var ec2 = new AWS.EC2();

function getInstanceInfo (params, callback) {
   ec2.describeInstances({
    InstanceIds: [params.instanceId]
  },function (err, data) {
    if (err) {
      console.log(err.stack);
      return callback(err.stack);
    }

    var html = [];

   
    data.Reservations.forEach(function (reservation, i) {
      if (i !== 0) html.push('\n');

        reservation.Instances.forEach(function (instance, i) {
        html.push('InstanceId: ' + instance.InstanceId);
        html.push('LaunchTime: ' + instance.LaunchTime);
        html.push('State: ' + instance.State.Name);
        html.push('PrivateDnsName: ' + instance.PrivateDnsName);
        html.push('PrivateIpAddress: ' + instance.PrivateIpAddress);
        html.push('InstanceType: ' + instance.InstanceType);
        html.push('AvailabilityZone: ' + instance.Placement.AvailabilityZone);
        html.push('Tags: ');

        instance.Tags.forEach(function (tag) {
          html.push('        ' + tag.Key + ': ' + tag.Value)
        });

      });
    });

    return callback(null, html.join('\n'));
  });
}

function postToSlack(params, callback) {
  var toSlack;

  if (params.environment === 'dev') {
    toSlack = {
      "attachments": [
        {
          "color": "good",
          "title": "Development",
          "text": params.info,
          "thumb_url": "https://cdn3.iconfinder.com/data/icons/macosxstyle/macosxstyle_png/512/Xcode.png"
        }
      ]
    }
  } else {
    toSlack = {
      "attachments": [
        {
          "color": "danger",
          "title": "Production",
          "text": params.info,
          "thumb_url": "http://dalistudios.com/wp-content/uploads/2013/06/DaliStudios-Social-Icons-Production.png"
        }
      ]
    }
  }

  slack_req_opts = url.parse('https://hooks.slack.com/services/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');
  slack_req_opts.method = 'POST';
  slack_req_opts.headers = {'Content-Type': 'application/json'}
        
  var req = https.request(slack_req_opts, function (res) {
    if (res.statusCode === 200) {
      return callback(null, 'Posted to slack')
    } else {
      return callback('status code: ' + res.statusCode);
    }
  });
          
  req.on('error', function(e) {
    console.log('problem with request: ' + e.message);
    return callback(e.message);
  });

  req.write(JSON.stringify(toSlack));
  req.end();
}

exports.handler = function(event, context) {
  (event.Records || []).forEach(function (rec) {
  
    if (rec.Sns) {
    
      var obj = JSON.parse(rec.Sns.Message);
      
   
      getInstanceInfo ({
        instanceId: obj.detail['EC2InstanceId'] || obj.detail['instance-id'],
      }, function (err, instanceInfo) {
        if (err) {
          context.fail(err);
        } else {
        var moreInfo = [];

          if (obj.detail['AutoScalingGroupName']) {
            moreInfo = ['\n',
              '<< AutoscalingGroup-Minions >>',
              'Service:' + obj['source'],
              'Evento:' + obj['detail-type'],
              'Region:' + obj['region'],
              'Descrition:' + obj.detail['Description'],
              'ASG:' + obj.detail['AutoScalingGroupName'],
              'Cause:' + obj.detail['Cause'],
            ];
          }

          instanceInfo = instanceInfo + moreInfo.join('\n');
          postToSlack({
            info: instanceInfo,
            environment: rec.Sns.TopicArn === 'arn:aws:sns:us-east-1:518863443564:kubernetes-dev-events' ? 'dev' : 'prod'
            
          }, function(err, message) {
            if (err) {
              context.fail(err);
            } else {
              context.succeed(message);
            }
          });
        }
      })
    }
  });
};
