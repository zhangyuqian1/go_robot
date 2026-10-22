package module

import (
	"testing"

	"stressTest/gameproto/skill"
)

func TestHandleSkillDataRecordsMinSkillID(t *testing.T) {
	r := newTestRobot()
	body := mustProtoBytes(t, &skill.SSkillData_100{
		Skills: []*skill.SSkillData_100SkillInfo{
			{Id: 1102},
			{Id: 1101},
			{Id: 1103},
		},
	})

	r.handleSkillData(body)

	if r.Client.MinSkillID != 1101 {
		t.Fatalf("MinSkillID = %d, want 1101", r.Client.MinSkillID)
	}
	if len(r.Client.Skills) != 3 {
		t.Fatalf("Skills len = %d, want 3", len(r.Client.Skills))
	}
}

func TestHandleSkillDataIgnoresZeroID(t *testing.T) {
	r := newTestRobot()
	body := mustProtoBytes(t, &skill.SSkillData_100{
		Skills: []*skill.SSkillData_100SkillInfo{
			{Id: 0},
			{Id: 1102},
		},
	})

	r.handleSkillData(body)

	if r.Client.MinSkillID != 1102 {
		t.Fatalf("MinSkillID = %d, want 1102", r.Client.MinSkillID)
	}
}

func TestHandleSkillDataExcludes1012(t *testing.T) {
	r := newTestRobot()
	body := mustProtoBytes(t, &skill.SSkillData_100{
		Skills: []*skill.SSkillData_100SkillInfo{
			{Id: 1012},
			{Id: 1101},
			{Id: 1015},
		},
	})

	r.handleSkillData(body)

	if r.Client.MinSkillID != 1101 {
		t.Fatalf("MinSkillID = %d, want 1101 (1012 excluded)", r.Client.MinSkillID)
	}
}

func TestHandleSkillDataAllExcluded(t *testing.T) {
	r := newTestRobot()
	body := mustProtoBytes(t, &skill.SSkillData_100{
		Skills: []*skill.SSkillData_100SkillInfo{
			{Id: 1012},
		},
	})

	r.handleSkillData(body)

	if r.Client.MinSkillID != 0 {
		t.Fatalf("MinSkillID = %d, want 0 (only 1012 was received)", r.Client.MinSkillID)
	}
}
