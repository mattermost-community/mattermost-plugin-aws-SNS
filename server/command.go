package main

import (
	"fmt"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
	"strings"
)

const awssns_cmd = "awssns"

func (p *Plugin) registerCommands() error {
	if err := p.API.RegisterCommand(&model.Command{
		Trigger:          awssns_cmd,
		Description:      "Mattermost slash command to interact with AWS SNS",
		DisplayName:      "AWS SNS",
		AutoComplete:     true,
		AutoCompleteHint: "listTopics",
		AutoCompleteDesc: "List Topics which are subscribed to the channel",
	}); err != nil {
		return errors.Wrapf(err, "failed to register awssns command")
	}

	return nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	splitCmd := strings.Fields(args.Command)
	cmd := strings.TrimPrefix(splitCmd[0], "/")
	action := splitCmd[1]

	if cmd != awssns_cmd {
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         fmt.Sprintf("Unknown Command: " + cmd),
		}, nil
	}

	switch action {
	case "listTopics":
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
			Text:         fmt.Sprintf("Implementation pending"),
		}, nil
	default:
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         fmt.Sprintf("Unknown Action: " + action),
		}, nil
	}
}
