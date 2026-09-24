// Package protocol defines message types, framing, and wire protocol definitions for StraitGateway.
package protocol

// Magic and protocol versions for StraitGateway IPC and wire streams.
const (
	MagicHeader    uint32 = 0x53475759 // "SGWY" in ASCII
	CurrentVersion uint16 = 1
)

// MessageType indicates the type of protocol payload.
type MessageType uint16

const (
	MsgHeartbeat    MessageType = 1
	MsgRouteUpdate  MessageType = 2
	MsgServiceSync  MessageType = 3
	MsgPolicySync   MessageType = 4
	MsgPacketEvent  MessageType = 5
	MsgError        MessageType = 99
)

// Header is the common message envelope for StraitGateway wire communication.
type Header struct {
	Magic   uint32      `json:"magic"`
	Version uint16      `json:"version"`
	Type    MessageType `json:"type"`
	Length  uint32      `json:"length"`
	SeqNum  uint64      `json:"seqNum"`
}

// StatusCode represents standard StraitGateway response status codes.
type StatusCode uint32

const (
	StatusOK              StatusCode = 0
	StatusInvalidArgument StatusCode = 1
	StatusNotFound        StatusCode = 2
	StatusAlreadyExists   StatusCode = 3
	StatusPermissionDenied StatusCode = 4
	StatusDatapathError   StatusCode = 5
	StatusInternalError   StatusCode = 6
)
