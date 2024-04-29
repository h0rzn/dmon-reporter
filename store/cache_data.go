package store

import (
	"time"
)

type Data struct {
	containerID string
	when        time.Time
	data        any
}

func NewData(id string, time time.Time, data any) *Data {
	return &Data{
		containerID: id,
		when:        time,
		data:        data,
	}
}

func (d *Data) ID() string {
	return d.containerID
}

func (d *Data) When() time.Time {
	return d.when
}

func (d *Data) Content() any {
	return d.data
}
