package shared

// const SOCKET_PATH = "/home/dungeon/.dungeon.sock"
// const LOG_FILE = "/home/dungeon/.dungeon.log"

// DEBUG
const SOCKET_PATH = "/home/dungeon/.dungeon.dbg.sock"
const LOG_FILE = "/home/dungeon/.dungeon.dbg.log"

const (
	RequestTypeLogin byte = iota
	RequestTypeSay
	RequestTypeWho
	RequestTypeLook
	RequestTypeMovement
)

const (
	ResponseTypeLogin byte = iota
	ResponseTypeLogout
	ResponseTypeLoggedInUsers
	ResponseTypeSay
	ResponseTypeLook
	ResponseTypeCantMove
)

type Direction byte

const (
	DirectionNorth Direction = iota
	DirectionEast
	DirectionSouth
	DirectionWest
)

const (
	CantMoveReasonNoExit byte = iota
	CantMoveReasonTM
)
