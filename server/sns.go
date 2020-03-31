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
