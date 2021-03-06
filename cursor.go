package aranGO

import (
	"encoding/json"
	"errors"
	"reflect"
	"time"
)

type Cursor struct {
	db *Database `json:"-"`
	Id string    `json:"Id"`

	Index  int           `json:"-"`
	Result []interface{} `json:"result"`
	More   bool          `json:"hasMore"`
	Amount int           `json:"count"`
	Data   Extra         `json:"extra"`

	Err       bool   `json:"error"`
	ErrMsg    string `json:"errorMessage"`
	ErrorExec error  `json:"execError"`
	Code      int    `json:"code"`
	max       int
	Time      time.Duration `json:"time"`
}

func NewCursor(db *Database) *Cursor {
	var c Cursor
	if db == nil {
		return nil
	}
	c.db = db
	return &c
}

// Delete cursor in server and free RAM
func (c *Cursor) Delete() error {
	if c.Id == "" {
		return errors.New("Invalid cursor to delete")
	}
	res, err := c.db.send("cursor", c.Id, "DELETE", nil, c, c)
	if err != nil {
		return nil
	}

	switch res.Status() {
	case 202:
		return nil
	case 404:
		return errors.New("Cursor does not exist")
	default:
		return nil
	}

}

func (c *Cursor) FetchBatch(r interface{}) error {
	var kind reflect.Kind
	if reflect.TypeOf(r).Kind() != reflect.Ptr {
		kind = reflect.ValueOf(r).Kind()
	} else {
		kind = reflect.ValueOf(r).Elem().Kind()
	}

	if kind != reflect.Slice && kind != reflect.Array {
		return errors.New("Container must be Slice of array kind")
	}
	b, err := json.Marshal(c.Result)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, r)
	if err != nil {
		return err
	}

	// fetch next batch
	if c.HasMore() {
		res, err := c.db.send("cursor", c.Id, "PUT", nil, c, c)

		if res.Status() == 200 {
			return nil
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// Iterates over cursor, returns false when no more values into batch, fetch next batch if necesary.
//
/*
b, err := json.Marshal(c.Result[c.Index])
err = json.Unmarshal(b, r)
*/
//
func (c *Cursor) FetchOne(r interface{}) bool {
	var max int = len(c.Result) - 1

	if c.Index > max {
		//fetch rest from server
		res, err := c.db.send("cursor", c.Id, "PUT", nil, c, c)
		if err != nil {
			c.ErrorExec = err
			return false
		}

		if res.Status() == 200 {
			c.Index = 0
			return true
		} else {
			return false
		}
	} else {
		b, err := json.Marshal(c.Result[c.Index])
		err = json.Unmarshal(b, r)
		if err != nil {
			c.ErrorExec = err
			return false
		}
		c.Index++
		return true
	}
}

// move cursor index by 1
func (c *Cursor) Next() bool {
	if c.Index == c.max {
		return false
	} else {
		c.Index++
		if c.Index == c.max {
			return true
		} else {
			return false
		}
	}
}

type Extra struct {
	Stat Stats `json:"stats"`
}

type Stats struct {
	WritesExecuted int    `json:"writesExecuted"`
	WritesIgnored  int    `json:"writesIgnored"`
	ScannedFull    int    `json:"scannedFull"`
	FullCount      uint64 `json:"fullCount"`
	ScannedIndex   int    `json:"scannedIndex"`
}

func (c Cursor) Count() int {
	return c.Amount
}

func (c *Cursor) FullCount() uint64 {
	return c.Data.Stat.FullCount
}

func (c Cursor) HasMore() bool {
	return c.More
}

func (c Cursor) Error() bool {
	return c.Err
}

func (c Cursor) ErrCode() int {
	return c.Code
}
