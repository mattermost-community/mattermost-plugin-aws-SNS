package main

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest/mock"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTeamChannelsNames(t *testing.T) {
	type args struct {
		teamChannel string
	}
	tests := []struct {
		name    string
		args    args
		want    []*TeamChannel
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "One team channel",
			args: args{teamChannel: "team1,channel1"},
			want: []*TeamChannel{
				{TeamName: "team1", ChannelName: "channel1", TeamID: "", ChannelID: ""},
			},
			wantErr: assert.NoError,
		},
		{
			name: "One team channel with channel separator",
			args: args{teamChannel: "team1,channel1;"},
			want: []*TeamChannel{
				{TeamName: "team1", ChannelName: "channel1", TeamID: "", ChannelID: ""},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Two team channels",
			args: args{teamChannel: "team1,channel1;team2,channel2"},
			want: []*TeamChannel{
				{TeamName: "team1", ChannelName: "channel1", TeamID: "", ChannelID: ""},
				{TeamName: "team2", ChannelName: "channel2", TeamID: "", ChannelID: ""},
			},
			wantErr: assert.NoError,
		},
		{
			name:    "Invalid team channel, channel missing",
			args:    args{teamChannel: "team1"},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name:    "Invalid team channel, channel2 missing",
			args:    args{teamChannel: "team1,channel1;team2"},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name:    "Invalid team channel, team1 missing",
			args:    args{teamChannel: ",channel1;team2;channel2"},
			want:    nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTeamChannelsNames(tt.args.teamChannel)
			if !tt.wantErr(t, err, fmt.Sprintf("parseTeamChannelsNames(%v)", tt.args.teamChannel)) {
				return
			}
			assert.Equalf(t, tt.want, got, "parseTeamChannelsNames(%v)", tt.args.teamChannel)
		})
	}
}

func TestOnActivate(t *testing.T) {
	for name, test := range map[string]struct {
		SetupAPI         func(*plugintest.API) *plugintest.API
		TeamChannel      string
		ExpectedChannels []*TeamChannel
		ShouldError      bool
	}{
		"SiteURL is not set": {
			SetupAPI: func(api *plugintest.API) *plugintest.API {
				api.On("GetConfig").Return(&model.Config{ServiceSettings: model.ServiceSettings{
					SiteURL: model.NewString(""),
				}})
				return api
			},
			ShouldError: true,
		},
		"Valid team channel": {
			SetupAPI: func(api *plugintest.API) *plugintest.API {
				botUserID := "yei0BahL3cohya8vuaboShaeSi"

				api.On("GetConfig").Return(&model.Config{ServiceSettings: model.ServiceSettings{
					SiteURL: model.NewString("mattermost.com"),
				}})
				api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything).Return()
				//p.API.GetTeamByName(teamChannel.TeamName)
				api.On("GetTeamByName", "team1").Return(&model.Team{Id: "teamId1"}, nil)
				api.On("GetChannelByName", "teamId1", "channel1", false).Return(&model.Channel{Id: "channelId1"}, nil)

				// Mock client ensure bot call
				api.On("GetServerVersion").Return("7.1.0")
				api.On("KVSetWithOptions", "mutex_mmi_bot_ensure", mock.AnythingOfType("[]uint8"), model.PluginKVSetOptions{Atomic: true, OldValue: []uint8(nil), ExpireInSeconds: 15}).Return(true, nil)
				api.On("KVSetWithOptions", "mutex_mmi_bot_ensure", []byte(nil), model.PluginKVSetOptions{ExpireInSeconds: 0}).Return(true, nil)
				path, err := filepath.Abs("..")
				require.Nil(t, err)
				api.On("GetBundlePath").Return(path, nil)
				api.On("SetProfileImage", botUserID, mock.Anything).Return(nil)
				api.On("EnsureBotUser", &model.Bot{
					Username:    "aws-sns",
					DisplayName: "AWS SNS Plugin",
					Description: "A bot account created by the plugin AWS SNS",
				}).Return(botUserID, nil)

				api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

				return api
			},
			TeamChannel: "team1,channel1",
			ExpectedChannels: []*TeamChannel{
				{TeamName: "team1", ChannelName: "channel1", TeamID: "teamId1", ChannelID: "channelId1"}},
			ShouldError: false,
		},
		"Valid multiple team channels": {
			SetupAPI: func(api *plugintest.API) *plugintest.API {
				botUserID := "yei0BahL3cohya8vuaboShaeSi"

				api.On("GetConfig").Return(&model.Config{ServiceSettings: model.ServiceSettings{
					SiteURL: model.NewString("mattermost.com"),
				}})
				api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything).Return()

				api.On("GetTeamByName", "team1").Return(&model.Team{Id: "teamId1"}, nil)
				api.On("GetChannelByName", "teamId1", "channel1", false).Return(&model.Channel{Id: "channelId1"}, nil)

				api.On("GetTeamByName", "team2").Return(&model.Team{Id: "teamId2"}, nil)
				api.On("GetChannelByName", "teamId2", "channel2", false).Return(&model.Channel{Id: "channelId2"}, nil)

				// Mock client ensure bot call
				api.On("GetServerVersion").Return("7.1.0")
				api.On("KVSetWithOptions", "mutex_mmi_bot_ensure", mock.AnythingOfType("[]uint8"), model.PluginKVSetOptions{Atomic: true, OldValue: []uint8(nil), ExpireInSeconds: 15}).Return(true, nil)
				api.On("KVSetWithOptions", "mutex_mmi_bot_ensure", []byte(nil), model.PluginKVSetOptions{ExpireInSeconds: 0}).Return(true, nil)
				path, err := filepath.Abs("..")
				require.Nil(t, err)
				api.On("GetBundlePath").Return(path, nil)
				api.On("SetProfileImage", botUserID, mock.Anything).Return(nil)
				api.On("EnsureBotUser", &model.Bot{
					Username:    "aws-sns",
					DisplayName: "AWS SNS Plugin",
					Description: "A bot account created by the plugin AWS SNS",
				}).Return(botUserID, nil)

				api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

				return api
			},
			TeamChannel: "team1,channel1;team2,channel2",
			ExpectedChannels: []*TeamChannel{
				{TeamName: "team1", ChannelName: "channel1", TeamID: "teamId1", ChannelID: "channelId1"},
				{TeamName: "team2", ChannelName: "channel2", TeamID: "teamId2", ChannelID: "channelId2"}},
			ShouldError: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			api := test.SetupAPI(&plugintest.API{})
			defer api.AssertExpectations(t)

			p := Plugin{}
			p.setConfiguration(&configuration{
				TeamChannel:    test.TeamChannel,
				AllowedUserIds: model.NewId(),
				Token:          model.NewId(),
			})
			p.SetAPI(api)
			p.client = pluginapi.NewClient(&plugintest.API{}, &plugintest.Driver{})
			err := p.OnActivate()

			if test.ShouldError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			assert.ElementsMatch(t, test.ExpectedChannels, p.Channels)
		})
	}
}
