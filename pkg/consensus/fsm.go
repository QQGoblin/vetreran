package consensus

import (
	"io"

	"github.com/hashicorp/raft"
)

type FSM struct{}

type FSMSnapshot struct{}

func (fsm FSM) Apply(*raft.Log) interface{} { return nil }

func (fsm FSM) Restore(io.ReadCloser) error { return nil }

func (fsm FSM) Snapshot() (raft.FSMSnapshot, error) { return FSMSnapshot{}, nil }

func (fsm FSM) StoreConfiguration(uint64, raft.Configuration) {}

func (snapshot FSMSnapshot) Persist(raft.SnapshotSink) error { return nil }

func (snapshot FSMSnapshot) Release() {}
