package module

import "stressTest/gameproto/common"

type EntityState struct {
	Handle int64
	X      int32
	Y      int32
	Dir    int32
	Props  map[int32]int64
}

type MonsterState struct {
	Handle int64
	ID     int32
	Midx   int32
	X      int32
	Y      int32
	CurHP  int64
	MaxHP  int64
	Name   string
	Skills []int32
}

type BossState struct {
	ID     int32
	X      int32
	Y      int32
	Relive int32
}

type PetState struct {
	SeriesID int32
	ID       int32
	Level    int32
	Star     int32
}

type EquipActionKind int

const (
	EquipActionNone EquipActionKind = iota
	EquipActionTakeOn
	EquipActionSell
)

type EquipAction struct {
	Kind   EquipActionKind
	Pos    int32
	Series *common.Series
	Items  []*common.UserItem
}
