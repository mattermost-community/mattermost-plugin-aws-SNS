package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

type Plugin struct {
	plugin.MattermostPlugin
	client *pluginapi.Client

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	BotUserID string
	Channels  []*TeamChannel
}

type TeamChannel struct {
	TeamID      string
	TeamName    string
	ChannelID   string
	ChannelName string
}

const topicsListPrefix = "topicsInChannel_"

func (t *TeamChannel) String() string {
	return fmt.Sprintf("TeamId: %s, TeamName: %s - ChannelId: %s, ChannelName: %s", t.TeamID, t.TeamName, t.ChannelID, t.ChannelName)
}
func (t *TeamChannel) NameString() string {
	return fmt.Sprintf("%s,%s", t.TeamName, t.ChannelName)
}

func (p *Plugin) OnActivate() error {
	p.client = pluginapi.NewClient(p.API, p.Driver)

	siteURL := p.API.GetConfig().ServiceSettings.SiteURL
	if siteURL == nil || *siteURL == "" {
		return errors.New("siteURL is not set. Please set a siteURL and restart the plugin")
	}

	configuration := p.getConfiguration()
	if err := p.IsValid(configuration); err != nil {
		return err
	}

	teamChannels, err := parseTeamChannelsNames(p.configuration.TeamChannel)

	if err != nil {
		return errors.New("teamChannel setting doesn't follow the pattern $TEAM_NAME,$CHANNEL_NAME")
	}

	teamChannels, err = p.resolveAndSetTeamIDs(teamChannels)
	if err != nil {
		return err
	}

	botID, err := p.client.Bot.EnsureBot(&model.Bot{
		Username:    "aws-sns",
		DisplayName: "AWS SNS Plugin",
		Description: "A bot account created by the plugin AWS SNS",
	},
		pluginapi.ProfileImagePath("assets/icon.png"),
	)
	if err != nil {
		return errors.Wrap(err, "can't ensure bot")
	}
	p.BotUserID = botID

	// get or create channel if it does not exist yet and add mattermost channel id to each teamChannel
	teamChannels, err = p.getOrCreateMattermostChannels(teamChannels)
	if err != nil {
		return err
	}

	p.Channels = teamChannels

	p.API.LogInfo("channels resvolved", "tc", teamChannels)
	if err := p.registerCommands(); err != nil {
		return err
	}

	return nil
}

func (p *Plugin) resolveAndSetTeamIDs(channels []*TeamChannel) ([]*TeamChannel, error) {
	//mattermostChannels := []TeamChannel{}
	for _, teamChannel := range channels {
		p.API.LogInfo("resolve for teamchannel", "tc", teamChannel)
		team, appErr := p.API.GetTeamByName(teamChannel.TeamName)
		if appErr != nil {
			return nil, appErr
		}
		teamChannel.TeamID = team.Id
	}
	return channels, nil
}

func parseTeamChannelsNames(teamChannel string) ([]*TeamChannel, error) {
	channels := []*TeamChannel{}
	splitChannels := strings.Split(teamChannel, ";")
	for _, splitChannel := range splitChannels {
		if len(splitChannel) < 1 {
			continue
		}
		split := strings.Split(splitChannel, ",")
		if len(split) != 2 {
			return nil, errors.New("teamChannel setting doesn't follow the pattern $TEAM_NAME,$CHANNEL_NAME")
		}
		channels = append(channels, &TeamChannel{
			TeamName:    split[0],
			ChannelName: split[1],
		})
	}
	return channels, nil
}

func (p *Plugin) getOrCreateChannel(teamChannel *TeamChannel) (string, error) {
	channel, appErr := p.API.GetChannelByName(teamChannel.TeamID, teamChannel.ChannelName, false)
	if appErr != nil && appErr.StatusCode == http.StatusNotFound {
		channelToCreate := &model.Channel{
			Name:        teamChannel.ChannelName,
			DisplayName: teamChannel.ChannelName,
			Type:        model.ChannelTypeOpen,
			TeamId:      teamChannel.TeamID,
			CreatorId:   p.BotUserID,
		}

		p.API.LogInfo("Creating Channel", "name", teamChannel.ChannelName)
		newChannel, errChannel := p.API.CreateChannel(channelToCreate)
		if errChannel != nil {
			return "", errChannel
		}
		return newChannel.Id, nil
	} else if appErr != nil {
		p.API.LogWarn("apperr", "error", appErr)
		return "", appErr
	} else {
		return channel.Id, nil
	}
}
func (p *Plugin) getOrCreateMattermostChannels(teamChannels []*TeamChannel) ([]*TeamChannel, error) {
	for _, teamChannel := range teamChannels {
		channelID, err := p.getOrCreateChannel(teamChannel)
		if err != nil {
			return nil, err
		}
		teamChannel.ChannelID = channelID
	}
	return teamChannels, nil
}

func (p *Plugin) IsValid(configuration *configuration) error {
	if configuration.TeamChannel == "" {
		return fmt.Errorf("must set a Team and a TeamChannel")
	}

	if configuration.AllowedUserIds == "" {
		return fmt.Errorf("must set at least one User")
	}

	return nil
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	if err := p.checkToken(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		p.API.LogError("AWSSNS TOKEN INVALID")
		return
	}

	channel, err := p.checkChannel(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		p.API.LogError("Channel is invalid", "error", err.Error())
		return
	}

	snsMessageType := r.Header.Get("x-amz-sns-message-type")
	if snsMessageType == "" {
		p.handleAction(w, r)
	} else {
		switch snsMessageType {
		case "SubscriptionConfirmation":
			p.handleSubscriptionConfirmation(r.Body, channel)
		case "Notification":
			p.API.LogDebug("AWSSNS HandleNotification")
			p.handleNotification(r.Body, channel)
		case "UnsubscribeConfirmation":
			p.handleUnsubscribeConfirmation(r.Body, channel)
		default:
			break
		}
	}
}
func (p *Plugin) checkToken(r *http.Request) error {
	token := r.URL.Query().Get("token")
	if token == "" || strings.Compare(token, p.configuration.Token) != 0 {
		return fmt.Errorf("invalid or missing token")
	}
	return nil
}

func (p *Plugin) checkChannel(r *http.Request) (*TeamChannel, error) {
	teamChannel := r.URL.Query().Get("channel")

	// fallback for old url configuration without channel parameter, use first channel as default
	if len(teamChannel) == 0 {
		return p.Channels[0], nil
	}

	for _, tc := range p.Channels {
		if strings.Compare(teamChannel, tc.NameString()) == 0 {
			return tc, nil
		}
	}
	return nil, fmt.Errorf("invalid channel %s", teamChannel)
}

func (p *Plugin) handleSubscriptionConfirmation(body io.Reader, channel *TeamChannel) {
	var subscribe SubscribeInput
	if err := json.NewDecoder(body).Decode(&subscribe); err != nil {
		return
	}

	p.sendSubscribeConfirmationMessage(subscribe.Message, subscribe.SubscribeURL, channel)
}

func (p *Plugin) handleNotification(body io.Reader, channel *TeamChannel) {
	var notification SNSNotification
	if err := json.NewDecoder(body).Decode(&notification); err != nil {
		p.API.LogDebug("AWSSNS HandleNotification Decode Error", "err=", err.Error())
		return
	}

	if isCloudformationEvent, messageNotification := p.isCloudformationEvent(notification.Message); isCloudformationEvent {
		p.API.LogDebug("Processing Cloudformation Event")
		p.sendPostNotification(p.createSNSCloudformationEventAttachment(notification.Subject, messageNotification), channel)
		return
	}

	if isRdsEvent, messageNotification := p.isRDSEvent(notification.Message); isRdsEvent {
		p.API.LogDebug("Processing RDS Event")
		p.sendPostNotification(p.createSNSRdsEventAttachment(notification.Subject, messageNotification), channel)
		return
	}

	if isAlarm, messageNotification := p.isCloudWatchAlarm(notification.Message); isAlarm {
		p.API.LogDebug("Processing CloudWatch alarm")
		p.sendPostNotification(p.createSNSMessageNotificationAttachment(notification.Subject, messageNotification), channel)
		return
	}

	if p.configuration.EnableUnknownTypeMessages {
		p.sendPostNotification(p.createSNSUnknownTypeMessage(notification.Subject, notification.Message), channel)
	}
}

func (p *Plugin) sendPostNotification(attachment model.SlackAttachment, channel *TeamChannel) {
	post := &model.Post{
		ChannelId: channel.ChannelID,
		UserId:    p.BotUserID,
	}
	model.ParseSlackAttachment(post, []*model.SlackAttachment{&attachment})
	if _, appErr := p.API.CreatePost(post); appErr != nil {
		return
	}
}

func (p *Plugin) isCloudWatchAlarm(message string) (bool, SNSMessageNotification) {
	var messageNotification SNSMessageNotification
	if err := json.Unmarshal([]byte(message), &messageNotification); err != nil {
		p.API.LogError(
			"AWSSNS HandleNotification Decode Error on CloudWatch message notification",
			"err", err.Error(),
			"message", message)
		return false, messageNotification
	}

	return len(messageNotification.AlarmName) > 0, messageNotification
}

func (p *Plugin) isRDSEvent(message string) (bool, SNSRdsEventNotification) {
	var messageNotification SNSRdsEventNotification
	if err := json.Unmarshal([]byte(message), &messageNotification); err != nil {
		p.API.LogError(
			"AWSSNS HandleNotification Decode Error on RDS-Event message notification",
			"err", err.Error(),
			"message", message)
		return false, messageNotification
	}
	return len(messageNotification.EventID) > 0, messageNotification
}

func (p *Plugin) isCloudformationEvent(message string) (bool, SNSCloudformationEventNotification) {
	var messageNotification SNSCloudformationEventNotification

	// alter message in order to decode it in json format
	messagejson, err := messageToJSON(message)

	if err != nil {
		p.API.LogError(
			"AWSSNS HandleNotification Decode Error on Cloudformation-Event message notification",
			"err", err.Error(),
			"message", message)
		return false, messageNotification
	}

	if messagejson != nil {
		if err := json.Unmarshal(messagejson, &messageNotification); err != nil {
			p.API.LogError(
				"AWSSNS HandleNotification Decode Error on Cloudformation-Event message notification",
				"err", err.Error(),
				"message", message)
			return false, messageNotification
		}
		return len(messageNotification.EventID) > 0, messageNotification
	}
	return false, messageNotification
}

func (p *Plugin) createSNSRdsEventAttachment(subject string, messageNotification SNSRdsEventNotification) model.SlackAttachment {
	p.API.LogDebug("AWSSNS HandleNotification RDS Event", "MESSAGE", subject)

	var fields []*model.SlackAttachmentField

	fields = addFields(fields, "Event Source", messageNotification.EventSource, true)
	fields = addFields(fields, "Event Time", messageNotification.EventTime, true)
	fields = addFields(fields, "Identifier Link", messageNotification.IdentifierLink, true)
	fields = addFields(fields, "Source ID", messageNotification.SourceID, true)
	fields = addFields(fields, "Event ID", messageNotification.EventID, true)
	fields = addFields(fields, "Event Message", messageNotification.EventMessage, true)

	attachment := model.SlackAttachment{
		Title:  subject,
		Fields: fields,
	}

	return attachment
}

func (p *Plugin) createSNSCloudformationEventAttachment(subject string, messageNotification SNSCloudformationEventNotification) model.SlackAttachment {
	p.API.LogDebug("AWSSNS HandleNotification Cloudformation Event", "SUBJECT", subject)
	var fields []*model.SlackAttachmentField

	fields = addFields(fields, "StackId", messageNotification.StackID, true)
	fields = addFields(fields, "StackName", messageNotification.StackName, true)
	fields = addFields(fields, "LogicalResourceId", messageNotification.LogicalResourceID, true)
	fields = addFields(fields, "PhysicalResourceId", messageNotification.PhysicalResourceID, true)
	fields = addFields(fields, "ResourceType", messageNotification.ResourceType, true)
	fields = addFields(fields, "Timestamp", messageNotification.Timestamp, true)
	fields = addFields(fields, "ResourceStatus", messageNotification.ResourceStatus, true)

	attachment := model.SlackAttachment{
		Title:  subject,
		Fields: fields,
	}

	return attachment
}

func (p *Plugin) createSNSUnknownTypeMessage(subject string, message string) model.SlackAttachment {
	p.API.LogDebug("AWSSNS HandleNotification Unknown Type Message", "SUBJECT", subject)

	text := message

	jsonBytes := []byte(message)
	prettyJSON := &bytes.Buffer{}
	err := json.Indent(prettyJSON, jsonBytes, "", "  ")
	if err == nil {
		text = "```json\n" + prettyJSON.String() + "\n```"
	}

	post := model.SlackAttachment{
		Title: subject,
		Text:  text,
	}

	return post
}

func (p *Plugin) createSNSMessageNotificationAttachment(subject string, messageNotification SNSMessageNotification) model.SlackAttachment {
	p.API.LogDebug("AWSSNS HandleNotification", "MESSAGE", subject)
	var fields []*model.SlackAttachmentField

	fields = addFields(fields, "AlarmName", messageNotification.AlarmName, true)
	fields = addFields(fields, "AlarmDescription", messageNotification.AlarmDescription, true)
	fields = addFields(fields, "AWS Account", messageNotification.AWSAccountID, true)
	fields = addFields(fields, "Region", messageNotification.Region, true)
	fields = addFields(fields, "New State", messageNotification.NewStateValue, true)
	fields = addFields(fields, "Old State", messageNotification.OldStateValue, true)
	fields = addFields(fields, "New State Reason", messageNotification.NewStateReason, false)
	fields = addFields(fields, "MetricName", messageNotification.Trigger.MetricName, true)
	fields = addFields(fields, "Namespace", messageNotification.Trigger.Namespace, true)
	fields = addFields(fields, "StatisticType", messageNotification.Trigger.StatisticType, true)
	fields = addFields(fields, "Statistic", messageNotification.Trigger.Statistic, true)
	fields = addFields(fields, "Period", strconv.Itoa(messageNotification.Trigger.Period), true)
	fields = addFields(fields, "EvaluationPeriods", strconv.Itoa(messageNotification.Trigger.EvaluationPeriods), true)
	fields = addFields(fields, "ComparisonOperator", messageNotification.Trigger.ComparisonOperator, true)
	fields = addFields(fields, "Threshold", fmt.Sprintf("%f", messageNotification.Trigger.Threshold), true)

	var dimensions []string
	for _, dimension := range messageNotification.Trigger.Dimensions {
		dimensions = append(dimensions, fmt.Sprintf("%s: %s", dimension.Name, dimension.Value))
	}
	fields = addFields(fields, "Dimensions", strings.Join(dimensions, "\n"), false)

	msgColor := "#008000"
	if messageNotification.NewStateValue == "ALARM" {
		msgColor = "#FF0000"
	} else if messageNotification.NewStateValue == "INSUFFICIENT" {
		msgColor = "#FFFF00"
	}

	attachment := model.SlackAttachment{
		Title:  subject,
		Fields: fields,
		Color:  msgColor,
	}

	return attachment
}
func (p *Plugin) handleUnsubscribeConfirmation(body io.Reader, channel *TeamChannel) {
	var subscribe SubscribeInput
	if err := json.NewDecoder(body).Decode(&subscribe); err != nil {
		return
	}
	topic := strings.Split(subscribe.TopicArn, ":")[5]
	if err := p.deleteFromKVStore(topic, channel.ChannelID); err != nil {
		p.API.LogError("Unable to delete %s from KV Store", topic)
	}
}

func (p *Plugin) sendSubscribeConfirmationMessage(message string, subscriptionURL string, channel *TeamChannel) {
	config := p.API.GetConfig()
	siteURLPort := *config.ServiceSettings.SiteURL
	action1 := &model.PostAction{
		Name: "Confirm Subscription",
		Type: model.PostActionTypeButton,
		Integration: &model.PostActionIntegration{
			Context: map[string]interface{}{
				"action":           "confirm",
				"subscription_url": subscriptionURL,
			},
			URL: fmt.Sprintf("%v/plugins/%v/confirm?token=%v&channel=%s", siteURLPort, manifest.Id, p.configuration.Token, channel.NameString()),
		},
	}

	actionMsg := strings.Split(message, ".")
	sa1 := &model.SlackAttachment{
		Text: actionMsg[0],
		Actions: []*model.PostAction{
			action1,
		},
	}
	attachments := make([]*model.SlackAttachment, 0)
	attachments = append(attachments, sa1)

	spinPost := &model.Post{
		Message:   "",
		ChannelId: channel.ChannelID,
		UserId:    p.BotUserID,
		Props: model.StringInterface{
			"attachments": attachments,
		},
	}

	if _, err := p.API.CreatePost(spinPost); err != nil {
		p.API.LogError(
			"We could not create subscription post",
			"user_id", p.BotUserID,
			"err", err.Error(),
		)
	}
	p.API.LogDebug(
		"Posted new subscription",
		"user_id", p.BotUserID,
		"subscriptionURL", subscriptionURL,
	)
}

func (p *Plugin) handleAction(w http.ResponseWriter, r *http.Request) {
	var action *Action
	err := json.NewDecoder(r.Body).Decode(&action)
	if err != nil || action == nil {
		encodeEphermalMessage(w, fmt.Sprintf("SNS BOT Error: We could not decode the action. Error=%s", err.Error()))
		p.API.LogError("SNS BOT Error: We could not decode the action.", "err=", err.Error())
		return
	}

	if err := p.checkAllowedUsers(action.UserID); err != nil {
		encodeEphermalMessage(w, err.Error())
		return
	}

	switch r.URL.Path {
	case "/confirm":
		resp, err := http.Get(action.Context.SubscriptionURL)
		if err != nil {
			encodeEphermalMessage(w, err.Error())
			return
		}
		defer resp.Body.Close()

		updatePost := &model.Post{}
		updateAttachment := &model.SlackAttachment{}
		actionPost, errPost := p.API.GetPost(action.PostID)
		if errPost != nil {
			p.API.LogError("AWSSNS Update Post Error", "err=", errPost.Error())
		} else {
			for _, attachment := range actionPost.Attachments() {
				if attachment.Text != "" {
					userName, errUser := p.API.GetUser(action.UserID)
					if errUser != nil {
						updateAttachment.Text = fmt.Sprintf("%s\n**Subscription Confirmed.**", attachment.Text)
					}
					updateAttachment.Text = fmt.Sprintf("%s\n**Subscription Confirmed by %s**", attachment.Text, userName.Username)
				}
			}
			retainedProps := []string{"override_username", "override_icon_url"}
			updatePost.AddProp("from_webhook", "true")

			for _, prop := range retainedProps {
				if value, ok := actionPost.Props[prop]; ok {
					updatePost.AddProp(prop, value)
				}
			}

			model.ParseSlackAttachment(updatePost, []*model.SlackAttachment{updateAttachment})
			updatePost.Id = actionPost.Id
			updatePost.ChannelId = actionPost.ChannelId
			updatePost.UserId = actionPost.UserId
			if _, err := p.API.UpdatePost(updatePost); err != nil {
				encodeEphermalMessage(w, "Subscription Confirmed.")
				return
			}

			encodeEphermalMessage(w, "Subscription Confirmed.")

			// Extract the topic from Subscription Confirmation URL
			query, err := url.ParseQuery(action.Context.SubscriptionURL)
			if err != nil {
				p.API.LogError("Unable to parse Subscribe URL from AWS SNS")
				return
			}
			topic := strings.Split(query.Get("TopicArn"), ":")[5]
			// Store this topic in KV Store
			if err = p.updateKVStore(topic, actionPost.ChannelId); err != nil {
				p.API.LogError("Unable to store AWS SNS Topic in KV Store")
			}
			return
		}
	default:
		http.NotFound(w, r)
		return
	}
}

func (p *Plugin) updateKVStore(topicName string, channelID string) error {
	var topics = SNSTopics{}
	val, err := p.API.KVGet(topicsListPrefix + channelID)
	if err != nil {
		p.API.LogError("Unable to Get from KV Store")
		return err
	}
	if val == nil {
		// Initialize the map before first assignment
		topics.Topics = make(map[string]bool)
		topics.Topics[topicName] = true
	} else {
		unmarshalErr := json.Unmarshal(val, &topics)
		if unmarshalErr != nil {
			p.API.LogError("Unmarshal failed for existing Topics in KV Store")
			return unmarshalErr
		}
		topics.Topics[topicName] = true
	}
	b, marshalErr := json.Marshal(topics)
	if marshalErr != nil {
		p.API.LogError("Unable to marshal Topics struct to JSON")
		return marshalErr
	}
	p.API.KVSet(topicsListPrefix+channelID, b)
	return nil
}

func (p *Plugin) deleteFromKVStore(topicName string, channelID string) error {
	val, err := p.API.KVGet(topicsListPrefix + channelID)
	if err != nil {
		p.API.LogError("Unable to Get from KV Store")
		return err
	}
	if val == nil {
		p.API.LogError("Unexpected: No item found in KV Store")
		return err
	}
	var topics SNSTopics
	if unmarshalErr := json.Unmarshal(val, &topics); unmarshalErr != nil {
		p.API.LogError("Failed to Unmarshal into struct")
		return err
	}
	delete(topics.Topics, topicName)
	b, marshalErr := json.Marshal(topics)
	if marshalErr != nil {
		p.API.LogError("Unable to Marshal the Topics struct")
		return err
	}
	p.API.KVSet(topicsListPrefix+channelID, b)
	return nil
}

func (p *Plugin) checkAllowedUsers(userID string) error {
	if userID == "" {
		return fmt.Errorf("need a user id")
	}

	hasPremissions := false
	AllowedUserIds := strings.Split(p.configuration.AllowedUserIds, ",")
	for _, allowedUserID := range AllowedUserIds {
		if allowedUserID == userID {
			hasPremissions = true
			break
		}
	}

	if !hasPremissions {
		return fmt.Errorf("you don't have permissions to use this command. Please talk with your SysAdmin")
	}

	return nil
}

func encodeEphermalMessage(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	payload := map[string]interface{}{
		"ephemeral_text": message,
	}

	json.NewEncoder(w).Encode(payload)
}

func addFields(fields []*model.SlackAttachmentField, title, msg string, short bool) []*model.SlackAttachmentField {
	return append(fields, &model.SlackAttachmentField{
		Title: title,
		Value: msg,
		Short: model.SlackCompatibleBool(short),
	})
}

func messageToJSON(message string) ([]byte, error) {
	messagefields := strings.Split(message, "\n")
	if len(messagefields) == 0 {
		return nil, errors.New("no message fields present in message string")
	}
	// examine if the message refers to a cloudformation event by checking if a valid StackId field is included in the first line
	stackIDParts := strings.Split(messagefields[0], "=")
	if len(stackIDParts) == 2 && stackIDParts[0] == "StackId" {
		containsCloudformationArn := strings.Contains(stackIDParts[1], "arn:aws:cloudformation")
		if !containsCloudformationArn {
			return nil, errors.New("invalid value of StackId field")
		}
	} else {
		return nil, nil
	}

	var numOfFields int

	// if "\n" existed at the end of the message, do not parse the last field
	if messagefields[len(messagefields)-1] == "" {
		numOfFields = len(messagefields) - 1
	} else {
		numOfFields = len(messagefields)
	}

	//split each line of the cloudformation event message to field and value
	var fields = make(map[string]string)
	for _, field := range messagefields[:numOfFields] {
		parts := strings.Split(field, "=")
		if len(parts) == 2 && parts[1] != "" {
			fields[parts[0]] = parts[1]
		} else {
			return nil, errors.New("format of Cloudformation event message is incorrect")
		}
	}

	jsonmessage, err := json.Marshal(fields)
	if err != nil {
		return nil, errors.Wrap(err, "Error marshaling in messageToJSON")
	}
	return jsonmessage, nil
}
