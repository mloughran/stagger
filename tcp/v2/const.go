package v2

// Encoding constants
//
// see ../../docs/ENCODING.md

const (
	PAIR_PING        = method(0x28) // Both
	PAIR_PONG        = method(0x29) // Both
	REPORT_ALL       = method(0x30) // Server -> Client
	REGISTER_PROCESS = method(0x41) // Client -> Server
	STATS_PARTIAL    = method(0x42) // Client -> Server
	STATS_COMPLETE   = method(0x43) // Client -> Server
)

const (
	MAGIC_BYTE1   = 0x83 // 'S'
	MAGIC_BYTE2   = 0x84 // 'T'
	MAGIC_VERSION = 0x00
)
