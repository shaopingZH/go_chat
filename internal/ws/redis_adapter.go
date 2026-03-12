package ws

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

const ChannelChatMessages = "chat:broadcast"

type ClusterMessage struct {
	TargetUserID uint64          `json:"target_id"`
	Payload      json.RawMessage `json:"payload"`
}

type RedisAdapter struct {
	client  *redis.Client
	manager *Manager
}

func NewRedisAdapter(client *redis.Client, manager *Manager) *RedisAdapter {
	adapter := &RedisAdapter{
		client:  client,
		manager: manager,
	}
	manager.SetRedisAdapter(adapter)
	return adapter
}

func (a *RedisAdapter) Publish(targetID uint64, payload []byte) error {
	msg := ClusterMessage{
		TargetUserID: targetID,
		Payload:      payload,
	}
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return a.client.Publish(context.Background(), ChannelChatMessages, bytes).Err()
}

func (a *RedisAdapter) Subscribe(ctx context.Context) {
	pubsub := a.client.Subscribe(ctx, ChannelChatMessages)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		var clusterMsg ClusterMessage
		if err := json.Unmarshal([]byte(msg.Payload), &clusterMsg); err != nil {
			continue
		}

		// Try to send to local user. If user is not on this node, manager will just return false (no-op).
		// This is correct: every node receives the message, checks if it has the user, and sends if so.
		a.manager.SendToUser(clusterMsg.TargetUserID, clusterMsg.Payload)
	}
}
