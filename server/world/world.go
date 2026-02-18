package world

import "math"

type Pos struct {
	X uint16
	Y uint16
	Z uint16
}
type Cell struct {
	Title       string
	Description string
}

func (p Pos) Hash() uint64 {
	return (uint64(p.Z) << 32) | (uint64(p.Y) << 16) | uint64(p.X)
}

var Cells = make(map[uint64]Cell)
var CenterPos = Pos{
	X: math.MaxUint16 / 2,
	Y: math.MaxUint16 / 2,
	Z: math.MaxUint16 / 2,
}

func CreateWorld() {
	Cells[CenterPos.Hash()] = Cell{
		Title:       "The Chapel",
		Description: "You wonder as to why there is a chapel deep beneath the earth, where the stone corridors of the dungeon wind like the roots of a great oak tree, yet here exists one, hewn from the living rock. It is not so vast as the cathedrals of great cities, yet neither is it small; its vaulted ceiling rises high enough that a tall banner might hang untroubled, and its nave would seat a modest gathering without crowding. The air within is cool and still, carrying the faint scent of old incense long settled into the stone. Its wooden pews that line the two sides of the room have seen better days, but have also seen days of use recently.",
	}
}
