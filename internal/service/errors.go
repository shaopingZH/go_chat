package service

import "errors"

var (
	ErrUsernameTaken      = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidRegister    = errors.New("invalid register payload")
	ErrUserNotFound       = errors.New("user not found")

	ErrGroupNotFound          = errors.New("group not found")
	ErrNotGroupMember         = errors.New("not group member")
	ErrInvalidGroupName       = errors.New("invalid group name")
	ErrNotGroupOwner          = errors.New("not group owner")
	ErrGroupOwnerCannotLeave  = errors.New("group owner cannot leave")
	ErrCannotRemoveGroupOwner = errors.New("cannot remove group owner")
	ErrTargetNotGroupMember   = errors.New("target is not group member")

	ErrTargetUserNotFound = errors.New("target user not found")
	ErrInvalidChatType    = errors.New("invalid chat type")
	ErrInvalidMsgType     = errors.New("invalid msg type")
	ErrInvalidContent     = errors.New("content is empty")
	ErrInvalidLastMsgID   = errors.New("invalid last_msg_id")
	ErrInvalidAroundMsgID = errors.New("invalid around_msg_id")
	ErrInvalidKeyword     = errors.New("invalid keyword")
	ErrInvalidParamCombo  = errors.New("invalid parameter combination")

	ErrInvalidProfilePayload = errors.New("invalid profile payload")
	ErrInvalidDisplayName    = errors.New("invalid display_name")
	ErrInvalidAvatar         = errors.New("invalid avatar")
	ErrInvalidBio            = errors.New("invalid bio")
	ErrInvalidUserIDs        = errors.New("invalid user ids")

	ErrCannotAddSelf           = errors.New("cannot send request to yourself")
	ErrRequestAlreadyPending   = errors.New("friend request already pending")
	ErrAlreadyFriends          = errors.New("already friends")
	ErrRequestNotFound         = errors.New("friend request not found")
	ErrRequestAlreadyProcessed = errors.New("request already processed")
	ErrNotYourRequest          = errors.New("not your request")
	ErrFriendshipNotFound      = errors.New("friendship not found")
	ErrNotFriend               = errors.New("not friend")
)
