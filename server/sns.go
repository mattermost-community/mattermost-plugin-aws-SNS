package main

import (
	"time"
)

// SubscribeInput - holds subscription and unsubscription confirmation
type SubscribeInput struct {
	Type             string    `json:"Type,omitempty"`
	MessageID        string    `json:"MessageId,omitempty"`
	Token            string    `json:"Token,omitempty"`
	TopicArn         string    `json:"TopicArn,omitempty"`
	Message          string    `json:"Message,omitempty"`
	SubscribeURL     string    `json:"SubscribeURL,omitempty"`
	Timestamp        time.Time `json:"Timestamp,omitempty"`
	SignatureVersion string    `json:"SignatureVersion,omitempty"`
	Signature        string    `json:"Signature,omitempty"`
	SigningCertURL   string    `json:"SigningCertURL,omitempty"`
}

// SNSNotification holds SNS Notification from AWS
type SNSNotification struct {
	Type             string    `json:"Type,omitempty"`
	MessageID        string    `json:"MessageId,omitempty"`
	TopicArn         string    `json:"TopicArn,omitempty"`
	Subject          string    `json:"Subject,omitempty"`
	Message          string    `json:"Message,omitempty"`
	SubscribeURL     string    `json:"SubscribeURL,omitempty"`
	Timestamp        time.Time `json:"Timestamp,omitempty"`
	SignatureVersion string    `json:"SignatureVersion,omitempty"`
	Signature        string    `json:"Signature,omitempty"`
	SigningCertURL   string    `json:"SigningCertURL,omitempty"`
	UnsubscribeURL   string    `json:"UnsubscribeURL,omitempty"`
}

// SNSMessageNotification holds the CloudWatch Alarm message from AWS
type SNSMessageNotification struct {
	AlarmName        string `json:"AlarmName"`
	AlarmDescription string `json:"AlarmDescription,omitempty"`
	AWSAccountID     string `json:"AWSAccountId"`
	NewStateValue    string `json:"NewStateValue"`
	NewStateReason   string `json:"NewStateReason"`
	StateChangeTime  string `json:"StateChangeTime"`
	Region           string `json:"Region"`
	OldStateValue    string `json:"OldStateValue"`
	Trigger          struct {
		MetricName    string `json:"MetricName"`
		Namespace     string `json:"Namespace"`
		StatisticType string `json:"StatisticType"`
		Statistic     string `json:"Statistic"`
		Unit          string `json:"Unit,omitempty"`
		Dimensions    []struct {
			Value string `json:"value"`
			Name  string `json:"name"`
		} `json:"Dimensions"`
		Period                           int     `json:"Period"`
		EvaluationPeriods                int     `json:"EvaluationPeriods"`
		ComparisonOperator               string  `json:"ComparisonOperator"`
		Threshold                        float32 `json:"Threshold"`
		TreatMissingData                 string  `json:"TreatMissingData"`
		EvaluateLowSampleCountPercentile string  `json:"EvaluateLowSampleCountPercentile"`
	} `json:"Trigger"`
}

type SNSRdsEventNotification struct {
	EventSource    string `json:"Event Source"`
	EventTime      string `json:"Event Time"`
	IdentifierLink string `json:"Identifier Link"`
	SourceID       string `json:"Source ID"`
	EventID        string `json:"Event ID"`
	EventMessage   string `json:"Event Message"`
}

type SNSCloudformationEventNotification struct {
	StackID              string `json:"StackId"`
	Timestamp            string `json:"Timestamp"`
	EventID              string `json:"EventId"`
	LogicalResourceID    string `json:"LogicalResourceId"`
	Namespace            string `json:"Namespace"`
	PhysicalResourceID   string `json:"PhysicalResourceId"`
	PrincipalID          string `json:"PrincipalId"`
	ResourceProperties   string `json:"ResourceProperties"`
	ResourceStatus       string `json:"ResourceStatus"`
	ResourceStatusReason string `json:"ResourceStatusReason"`
	ResourceType         string `json:"ResourceType"`
	StackName            string `json:"StackName"`
	ClientRequestToken   string `json:"ClientRequestToken"`
}

type SNSCodeBuildEventNotification struct {
	AccountID  string    `json:"account"`
	Region     string    `json:"region"`
	DetailType string    `json:"detailType"`
	Source     string    `json:"source"`
	Time       time.Time `json:"time"`
	Resources  []string  `json:"resources"`
	Detail     struct {
		ProjectName            string  `json:"project-name"`
		BuildID                string  `json:"build-id"`
		BuildStatus            string  `json:"build-status,omitempty"`
		CurrentPhase           string  `json:"current-phase,omitempty"`
		CurrentPhaseContext    string  `json:"current-phase-context,omitempty"`
		CompletedPhaseStatus   string  `json:"completed-phase-status,omitempty"`
		CompletedPhase         string  `json:"completed-phase,omitempty"`
		CompletedPhaseContext  string  `json:"completed-phase-context,omitempty"`
		CompletedPhaseDuration float32 `json:"completed-phase-duration-seconds,omitempty"`
		CompletedPhaseStart    string  `json:"completed-phase-start,omitempty"`
		CompletedPhaseEnd      string  `json:"completed-phase-end,omitempty"`
		AdditionalInformation  struct {
			Initiator      string  `json:"initiator"`
			BuildNumber    float32 `json:"build-number,omitempty"`
			BuildStartTime string  `json:"build-start-time"`
			BuildComplete  bool    `json:"build-complete"`
			Timeout        float32 `json:"timeout-in-minutes"`
			Artifact       struct {
				MD5Sum    string `json:"md5sum,omitempty"`
				SHA256Sum string `json:"sha256sum,omitempty"`
				Location  string `json:"location"`
			} `json:"artifact"`
			Environment struct {
				Image                string `json:"image"`
				PrivilegedMode       bool   `json:"privileged-mode"`
				ComputeType          string `json:"compute-type"`
				Type                 string `json:"type"`
				EnvironmentVariables []struct {
					Name  string `json:"name"`
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"environment-variables"`
			} `json:"environment"`
			Source struct {
				Location string `json:"location"`
				Type     string `json:"type"`
			} `json:"source"`
			SourceVersion string `json:"source-version"`
			Logs          struct {
				GroupName  string `json:"group-name"`
				StreamName string `json:"stream-name"`
				DeepLink   string `json:"deep-link"`
			} `json:"logs"`
			Phases []struct {
				PhaseContext []interface{} `json:"phase-context,omitempty"`
				StartTime    string        `json:"start-time"`
				EndTime      string        `json:"end-time,omitempty"`
				Duration     float32       `json:"duration-in-seconds,omitempty"`
				PhaseType    string        `json:"phase-type"`
				PhaseStatus  string        `json:"phase-status,omitempty"`
			} `json:"phases"`
		} `json:"additional-information"`
	} `json:"detail"`
}
