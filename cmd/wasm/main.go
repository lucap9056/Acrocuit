//go:build js

package main

import (
	"acrocuit/internal/storage/repository"
	"acrocuit/internal/storage/sqlite"
	"context"
	"syscall/js"
)

type repositories struct {
	SpaceGroups  repository.SpaceGroups
	Spaces       repository.Spaces
	SwitchPanels repository.SwitchPanels
	Switches     repository.Switches
	Devices      repository.Devices
}

func main() {
	ctx := context.Background()
	boot := js.Global().Get("acrocuitBoot")

	db, err := sqlite.Connect(ctx, boot.Get("sqlite3"), boot.Get("db"))
	if err != nil {
		boot.Call("reject", err.Error())
		return
	}

	if err := sqlite.ApplySchema(ctx, db); err != nil {
		db.Close()
		boot.Call("reject", err.Error())
		return
	}

	boot.Call("resolve", exportRepositories(ctx, repositories{
		SpaceGroups:  sqlite.NewSpaceGroups(db),
		Spaces:       sqlite.NewSpaces(db),
		SwitchPanels: sqlite.NewSwitchPanels(db),
		Switches:     sqlite.NewSwitches(db),
		Devices:      sqlite.NewDevices(db),
	}))

	select {}
}
