package bus

import (
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

func findStreamNameForSubject(subject string) (string, error) {
	configs := GetStreamConfigs()
	for _, sc := range configs {
		for _, s := range sc.Subjects {
			pattern := strings.TrimSuffix(s, ".*")
			if strings.HasPrefix(subject, pattern) {
				return sc.Name, nil
			}
		}
	}
	return "", fmt.Errorf("no stream configured for subject: %s", subject)
}

type StreamConfig struct {
	Name     string
	Subjects []string
	Config   *nats.StreamConfig
}

func GetStreamConfigs() []StreamConfig {
	return []StreamConfig{
		{
			Name:     "identity-stream",
			Subjects: []string{"user.*"},
			Config: &nats.StreamConfig{
				Name:     "identity-stream",
				Subjects: []string{"user.*"},
				Storage:  nats.FileStorage,
				Replicas: 3,
				MaxMsgs:  10000,
				MaxAge:   24 * time.Hour,
			},
		},
		{
			Name:     "wallet-stream",
			Subjects: []string{"wallet.*"},
			Config: &nats.StreamConfig{
				Name:     "wallet-stream",
				Subjects: []string{"wallet.*"},
				Storage:  nats.FileStorage,
				Replicas: 3,
				MaxMsgs:  10000,
				MaxAge:   24 * time.Hour,
			},
		},
		{
			Name:     "notification-stream",
			Subjects: []string{"notification.*"},
			Config: &nats.StreamConfig{
				Name:     "notification-stream",
				Subjects: []string{"notification.*"},
				Storage:  nats.FileStorage,
				Replicas: 3,
				MaxMsgs:  10000,
				MaxAge:   24 * time.Hour,
			},
		},
		{
			Name:     "dlq-stream",
			Subjects: []string{"dlq.*"},
			Config: &nats.StreamConfig{
				Name:     "dlq-stream",
				Subjects: []string{"dlq.*"},
				Storage:  nats.FileStorage,
				Replicas: 3,
				MaxMsgs:  10000,
				MaxAge:   24 * time.Hour,
			},
		},
	}
}
