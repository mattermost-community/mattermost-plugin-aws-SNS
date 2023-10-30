package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-api/experimental/command"
)

const awsSNSCmd = "awssns"

func (p *Plugin) registerCommands() error {
	iconData, iconError := command.GetIconData(p.API, "assets/icon.svg")
	if iconError != nil {
		return errors.Wrapf(iconError, "Failed to get icon data")
	}

	err := p.API.RegisterCommand(&model.Command{
		Trigger:              awsSNSCmd,
		Description:          "Mattermost slash command to interact with AWS SNS",
		DisplayName:          "AWS SNS",
		AutoComplete:         true,
		AutoCompleteHint:     "[command]",
		AutoCompleteDesc:     "Available commands: list-topics",
		AutocompleteData:     getAutoCompleteData(),
		AutocompleteIconData: iconData,
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

	action := ""
	if len(splitCmd) > 1 {
		action = splitCmd[1]
	}

	if cmd != awsSNSCmd {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Unknown Command: " + cmd),
		}, nil
	}

	switch action {
	case "list-topics":
		return p.listTopicsToChannel(args.ChannelId), nil
	default:
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Unknown Action: " + action),
		}, nil
	}
}

// listTopicsToChannel Lists topics subscribed to the channel
func (p *Plugin) listTopicsToChannel(channelID string) *model.CommandResponse {
	var topics SNSTopics
	val, err := p.API.KVGet(topicsListPrefix + channelID)
	if err != nil {
		p.API.LogError("Failed to Get from KV Store")
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         err.Error(),
		}
	}
	if val == nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "No Topics are subscribed by the configured channel",
		}
	}
	unMarshalErr := json.Unmarshal(val, &topics)
	if unMarshalErr != nil {
		p.API.LogError("Failed to Unmarshal")
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         unMarshalErr.Error(),
		}
	}
	resp := "The following SNS topics are subscribed by the configured channel\n"
	for topicName := range topics.Topics {
		resp = resp + "* " + topicName + "\n"
	}
	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeInChannel,
		Text:         resp,
	}
}

func getAutoCompleteData() *model.AutocompleteData {
	aws := model.NewAutocompleteData(awsSNSCmd, "[command]", "Available commands: list-topics")
	listTopics := model.NewAutocompleteData("list-topics", "", "Lists Topics which are subscribed to the channel")
	aws.AddCommand(listTopics)
	return aws
}
