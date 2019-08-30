module github.com/mattermost/mattermost-plugin-aws-SNS

go 1.12

require (
	github.com/go-ldap/ldap v3.0.3+incompatible // indirect
	github.com/lib/pq v1.1.1 // indirect
	github.com/mattermost/mattermost-server v5.14.1+incompatible
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors v0.8.1
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/crypto v0.0.0-20190513172903-22d7a77e9e5f // indirect
	google.golang.org/appengine v1.6.0 // indirect
)

// Workaround for https://github.com/golang/go/issues/30831 and fallout.
replace github.com/golang/lint => github.com/golang/lint v0.0.0-20190227174305-8f45f776aaf1
