package module

import (
	"testing"

	"stressTest/gameproto/common"
	"stressTest/gameproto/observer"
)

func TestHandlePetAppearCachesPetInfo(t *testing.T) {
	r := newTestRobot()
	body := mustProtoBytes(t, &observer.SPetAppear_10{
		Common: &common.EntityCommon{
			Handle: 20426260480,
			Id:     331,
			X:      9150,
			Y:      7350,
			Dir:    0,
		},
		CurHp:     1,
		MaxHp:     1,
		Name:      "万域幼神",
		MonsterId: 331,
		Skills:    []int32{8230, 8231},
	})

	r.handlePetAppear(body)

	if r.Client.PetHandle != 20426260480 {
		t.Fatalf("PetHandle = %d, want 20426260480", r.Client.PetHandle)
	}
	if r.Client.PetName != "万域幼神" {
		t.Fatalf("PetName = %q, want 万域幼神", r.Client.PetName)
	}
	if r.Client.PetMonsterID != 331 {
		t.Fatalf("PetMonsterID = %d, want 331", r.Client.PetMonsterID)
	}
}

func TestHandlePetAppearNilCommon(t *testing.T) {
	r := newTestRobot()
	body := mustProtoBytes(t, &observer.SPetAppear_10{
		Name: "无实体宠物",
	})

	r.handlePetAppear(body)

	if r.Client.PetHandle != 0 {
		t.Fatalf("PetHandle = %d, want 0 (no common entity)", r.Client.PetHandle)
	}
}
