package encoding

// Encoding constants
//
// see ../docs/ENCODING.md

const (
	PAIR_PING        = byte(0x28) // Both
	PAIR_PONG        = byte(0x29) // Both
	REPORT_ALL       = byte(0x30) // Server -> Client
	REGISTER_PROCESS = byte(0x41) // Client -> Server
	STATS_PARTIAL    = byte(0x42) // Client -> Server
	STATS_COMPLETE   = byte(0x43) // Client -> Server
)

const (
	MAGIC_BYTE1   = 0x83 // 'S'
	MAGIC_BYTE2   = 0x84 // 'T'
	MAGIC_VERSION = 0x00
)
