package main

import (
	"encoding/json"
	"fmt"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
	"strings"
)

const awsSNSCmd = "awssns"

func (p *Plugin) registerCommands() error {
	err := p.API.RegisterCommand(&model.Command{
		Trigger:          awsSNSCmd,
		Description:      "Mattermost slash command to interact with AWS SNS",
		DisplayName:      "AWS SNS",
		AutoComplete:     true,
		AutoCompleteHint: "listTopics",
		AutoCompleteDesc: "List Topics which are subscribed to the channel",
	})
	if err != nil {
		return errors.Wrapf(err, "failed to register awssns command")
	}

	return nil
}

// ExecuteCommand executes a command that has been previously registered via the RegisterCommand API
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	splitCmd := strings.Fields(args.Command)
	cmd := strings.TrimPrefix(splitCmd[0], "/")
	action := splitCmd[1]

	if cmd != awsSNSCmd {
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         fmt.Sprintf("Unknown Command: " + cmd),
		}, nil
	}

	switch action {
	case "listTopics":
		return p.listTopicsToChannel(), nil
	default:
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         fmt.Sprintf("Unknown Action: " + action),
		}, nil
	}
}

// listTopicsToChannel Lists topics subscribed to the channel
func (p *Plugin) listTopicsToChannel() *model.CommandResponse {
	var topics SNSTopics
	val, err := p.API.KVGet(topicsListPrefix + p.ChannelID)
	if err != nil {
		p.API.LogError("Failed to Get from KV Store")
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         fmt.Sprintf("%s", err.Error()),
		}
	}
	if val == nil {
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text: fmt.Sprintf(
				"No Topics are subscribed by the configured channel"),
		}
	}
	unMarshalErr := json.Unmarshal(val, &topics)
	if unMarshalErr != nil {
		p.API.LogError("Failed to Unmarshal")
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         fmt.Sprintf("%s", unMarshalErr.Error()),
		}
	}
	resp := "The following SNS topics are subscribed by the configured channel\n"
	for topicName := range topics.Topics {
		resp = resp + "* " + topicName + "\n"
	}
	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
		Text:         fmt.Sprintf("%s", resp),
	}
}
