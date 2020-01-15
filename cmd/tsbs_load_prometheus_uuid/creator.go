package main

// Prometheus remote storage don't have a database abstraction
type dbCreator struct{}

func (d *dbCreator) Init() {}

func (d *dbCreator) DBExists(_ string) bool { return true }

func (d *dbCreator) CreateDB(_ string) error { return nil }

func (d *dbCreator) RemoveOldDB(_ string) error { return nil }
