package adapters

import (
	"context"
	"sync"
)

type SpreadsheetsAPIStub struct {
	lock sync.Mutex
	Rows map[string][][]string
}

func (c *SpreadsheetsAPIStub) AppendRow(
	ctx context.Context,
	sheetName string,
	row []string,
) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.Rows == nil {
		c.Rows = make(map[string][][]string)
	}

	c.Rows[sheetName] = append(c.Rows[sheetName], row)

	return nil
}
