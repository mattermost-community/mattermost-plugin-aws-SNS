# AWS SNS Plugin [![CircleCI](https://circleci.com/gh/mattermost/mattermost-plugin-aws-SNS.svg?style=svg)](https://circleci.com/gh/mattermost/mattermost-plugin-aws-SNS)

This plugin receives SNS notification from Alerts created by AWS Cloudwatch and sent via AWS SNS.

## Configuration

### Mattermost side:

  - Install the plugin in your Mattermost instance
  - Configure the plugin
    - Set the `Team` and `Channel` that you want to receive the messages. If the Channel does not exist the plugin will create the channel for you.
    Set the `Team` and `Channel` with comma, ie. `teamb,channelx`.
  - Set the `AllowedUserIds` with the users ids, this will allow specific users to approve the subscription from AWS SNS.
  - Set the `Username`, this will be user that will post the messages.
  - Set the `Token`, this will be use when creating the SNS subscription.

### AWS Side:

 - Setup your AWS Cloudwatch alarms
 - Create an AWS SNS Topic
 - Create a HTTPS subscription poiting to `https://<YOUR_SITE_URL>/plugins/com.mattermost.aws-sns?token=<TOKEN_GENERATED_IN_MATTERMOST_PLUGIN_CONFIG>`
 - You will receive a message in the channel you setup to confirm the subscription.
 - Configure your Alarms to use the topic you created.
 - Start getting alerts when Cloudwatch trigger.

 Have fun!

 ## Next Steps

  - Add some tests
  - Add option to list topic
  - Add option to subcribe to topics
  - Add option to unsubscribe to topics
  - Add option to post in other channels
