package test

import "testing"

func TestE2E(t *testing.T) {
	c := newClient(t)
	b := newBuilding(t, c)

	for _, step := range []struct {
		name string
		run  func(t *testing.T, c *client, b *building)
	}{
		{"Links", testLinks},
		{"IncludeRejectsUnknownColumn", testIncludeRejectsUnknownColumn},
		{"PowerChain", testPowerChain},
		{"SpaceBackgroundImage", testSpaceBackgroundImage},
		{"DeleteSpaceRemovesCrossSpaceReferences", testDeleteSpaceRemovesCrossSpaceReferences},
		{"DeleteSpaceGroupCascades", testDeleteSpaceGroupCascades},
	} {
		t.Run(step.name, func(t *testing.T) {
			step.run(t, c.with(t), b)
		})
	}
}
