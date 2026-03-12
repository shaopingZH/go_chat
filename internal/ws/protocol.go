package ws

type IncomingEnvelope struct {
	Type    string              `json:"type"`
	Payload IncomingChatPayload `json:"payload"`
}

type IncomingChatPayload struct {
	TargetID uint64 `json:"target_id"`
	ChatType string `json:"chat_type"`
	MsgType  int8   `json:"msg_type,omitempty"`
	Content  string `json:"content"`
}

type OutgoingEnvelope struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

type SenderPayload struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

type OutgoingChatPayload struct {
	ID        uint64        `json:"id"`
	Sender    SenderPayload `json:"sender"`
	TargetID  uint64        `json:"target_id"`
	ChatType  string        `json:"chat_type"`
	MsgType   int8          `json:"msg_type"`
	Content   string        `json:"content"`
	CreatedAt string        `json:"created_at"`
}

type OutgoingOnlineStatusPayload struct {
	UserID     uint64 `json:"user_id"`
	Online     bool   `json:"online"`
	LastSeenAt string `json:"last_seen_at,omitempty"`
}

func ErrorEnvelope(message string) OutgoingEnvelope {
	return OutgoingEnvelope{
		Type: "error",
		Payload: map[string]string{
			"message": message,
		},
	}
}
