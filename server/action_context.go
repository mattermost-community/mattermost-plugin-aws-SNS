package main

// ActionContext passed from action buttons
type ActionContext struct {
	SubscriptionURL string `json:"subscription_url"`
	Action          string `json:"action"`
}

// Action type for decoding action buttons
type Action struct {
	UserID  string         `json:"user_id"`
	PostID  string         `json:"post_id"`
	Context *ActionContext `json:"context"`
}
