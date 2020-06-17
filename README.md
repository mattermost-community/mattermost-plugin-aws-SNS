# AWS SNS Plugin

[![Build Status](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-aws-SNS/master)](https://circleci.com/gh/mattermost/mattermost-plugin-aws-SNS)
[![Code Coverage](https://img.shields.io/codecov/c/github/mattermost/mattermost-plugin-aws-SNS/master)](https://codecov.io/gh/mattermost/mattermost-plugin-aws-SNS)
[![Release](https://img.shields.io/github/v/release/mattermost/mattermost-plugin-aws-SNS)](https://github.com/mattermost/mattermost-plugin-aws-SNS/releases/latest)
[![HW](https://img.shields.io/github/issues/mattermost/mattermost-plugin-aws-SNS/Up%20For%20Grabs?color=dark%20green&label=Help%20Wanted)](https://github.com/mattermost/mattermost-plugin-aws-SNS/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3A%22Up+For+Grabs%22+label%3A%22Help+Wanted%22)

**Maintainer:** [@mickmister](https://github.com/mickmister)
**Co-Maintainer:** [@jfrerich](https://github.com/jfrerich)

Send alert notifications from [Amazon AWS CloudWatch](https://aws.amazon.com/cloudwatch/) to Mattermost channels via AWS SNS.

Originally developed by [Carlos Tadeu Panato Junior](https://github.com/cpanato/).

![image](https://user-images.githubusercontent.com/13119842/58750029-df501000-845a-11e9-88f2-63fc0db5bc26.png)

## Configuration

### Step 1: Configure plugin in Mattermost

1. Go to **System Console > Plugins > AWS SNS**:

  1. Set the channel to send notifications to, specified in the format `teamname,channelname`. If the specified channel does not exist, the plugin will create the channel for you. 
      - Note: Must be the team and channel handle used in the URL. For example, in the following URL, set the value to `myteam,mychannel`: https://example.com/myteam/channels/mychannel

  2. Set authorized users who can accept AWS SNS subscriptions. Must be a comma-separated list of user ID's.
      - Tip: Use the [mattermost user search](https://mattermost.com/pl/cli-mattermost-user-search) CLI command to determine a user ID.

  3. Set the username that this integration is attached to.
  4. Generate a token used for an AWS SNS subscription. Copy this value as you will use it in a later step.

2. Go to **System Console > Plugins > Management** and click **Enable** to enable the AWS SNS plugin.

Note: If you are running Mattermost v5.11 or earlier, you must first go to the [releases page of this GitHub repository](https://github.com/mattermost/mattermost-plugin-aws-SNS/releases), download the latest release, and upload it to your Mattermost instance [following this documentation](https://docs.mattermost.com/administration/plugins.html#plugin-uploads).

### Step 2: Configure plugin in Amazon AWS

1. Create an [AWS CloudWatch alarm for your instance](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-cloudwatch-createalarm.html). 

2. Create an AWS SNS Topic with an HTTPS subscription to https://your-mattermost-url/plugins/com.mattermost.aws-sns?token=your-mattermost-token`, where `your-mattermost-url` refers to your Mattermost URL, and `your-mattermost-token` was generated on a previous step. [Follow this documentation](https://docs.safe.com/fme/html/FME_Server_Documentation/ReferenceManual/Amazon_SNS_Publisher_Configure_AWS_Subscription.html) for additional configuration options.

3. Switch to the Mattermost channel you configured to receive notifications. Accept the subscription posted to the channel by clicking the "Confirm" button.

4. Configure your AWS CloudWatch Alarms to use the topic you created previously.

You're all set! Alerts should now get posted from AWS CloudWatch to Mattermost.
  
## Development

This plugin contains a server portion.

Use `make dist` to build distributions of the plugin that you can upload to a Mattermost server.
Use `make check-style` to check the style.
Use `make deploy` to deploy the plugin to your local server.

For additional information on developing plugins, refer to [our plugin developer documentation](https://developers.mattermost.com/extend/plugins/).
