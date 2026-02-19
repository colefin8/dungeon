package world

import (
	"math"

	"dungeon/shared"
)

type Pos struct {
	X uint16
	Y uint16
	Z uint16
}

func (p Pos) Hash() uint64 {
	return (uint64(p.Z) << 32) | (uint64(p.Y) << 16) | uint64(p.X)
}

type Cell struct {
	Title       string
	Description string
	Exits       byte // each of this cell's exit `Direction`s OR'd together
}

var Cells = make(map[uint64]Cell)
var CenterPos = Pos{
	X: math.MaxUint16 / 2,
	Y: math.MaxUint16 / 2,
	Z: math.MaxUint16 / 2,
}

func CreateWorld() {
	cursor := CenterPos
	Cells[cursor.Hash()] = Cell{
		Title:       "The Chapel",
		Description: "You wonder as to why there is a chapel deep beneath the earth, where the stone corridors of the dungeon wind like the roots of a great oak tree, yet here exists one, hewn from the living rock. It is not so vast as the cathedrals of great cities, yet neither is it small; its vaulted ceiling rises high enough that a tall banner might hang untroubled, and its nave would seat a modest gathering without crowding. The air within is cool and still, carrying the faint scent of old incense long settled into the stone. Its wooden pews that line the two sides of the room have seen better days, but have also seen days of use recently.",
		Exits:       byte(shared.DirectionEast | shared.DirectionSouth),
	}
	cursor.X++
	Cells[cursor.Hash()] = Cell{
		Title:       "Dark Hallway Next to the Chapel",
		Description: "The hallway stretches long and low beneath the earth, carved from ancient stone that weeps with a chill and stubborn damp. Water gathers in the cracks between the blocks and falls at slow intervals, each drop striking the flagstones with a hollow note that travels farther than it ought. The air hangs thick with the scent of mildew and forgotten years, and a pale moss clings to the walls like a tattered cloak, lending a faint and ghostly gleam to the gloom. Iron sconces, bent and gnawed by rust, cradle dying torches whose flames flutter uneasily, as though they would sooner flee than keep watch. A wandering draft slips along the corridor, stirring cloaks and courage alike, and whispers of deeper chambers where darker things wait in patient silence.",
		Exits:       byte(shared.DirectionWest),
	}
	cursor.X--
	cursor.Y++
	Cells[cursor.Hash()] = Cell{
		Title:       "Great Wooden Archway to the Chapel",
		Description: "Some kind of description goes here describing a great wooden archway leading to an underground chapel incorporated into this underground dungeon.",
		Exits:       byte(shared.DirectionNorth),
	}
}
