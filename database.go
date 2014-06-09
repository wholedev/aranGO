package aranGO

import (
	"errors"
	nap "github.com/jmcvetta/napping"
)

// Database
type Database struct {
	Name        string
	Collections []Collection
	sess        *Session
	baseURL     string
}

// Execute AQL query into the server
func (d *Database) Execute(q *Query) (*Cursor, error) {
	if q == nil {
		return nil, errors.New("Cannot execute nil query")
	} else {
		// check if I need to validate query
		if q.Validate {
			if !d.IsValid(q) {
				return nil, errors.New(q.ErrorMsg)
			}
		}
		// create cursor
		c := NewCursor(d)
		_, err := d.send("cursor", "", "POST", q, c, c)
		if err != nil {
			return nil, err
		}
		return c, nil
	}
}

func (d *Database) ExecuteTran(t *Transaction) error {
	if t.Action == "" {
		return errors.New("Action must not be nil")
	}
	_, err := d.send("transaction", "", "POST", t, t, t)

	if err != nil {
		return err
	}

	return nil
}

func (d *Database) IsValid(q *Query) bool {
	if q != nil {
		res, err := d.send("query", "", "POST", map[string]string{"query": q.Aql}, q, q)
		if err != nil {
			return false
		}
		if res.Status() == 200 {
			return true
		} else {
			// could check error into query
			return false
		}
	} else {
		return false
	}
}

// Do a request to test if the database is up and user authorized to use it

func (d *Database) get(resource string, id string, method string, param *nap.Params, result, err interface{}) (*nap.Response, error) {
	url := d.buildRequest(resource, id)
	var r *nap.Response
	var e error

	switch method {
	case "OPTIONS":
		r, e = d.sess.nap.Options(url, result, err)
	case "HEAD":
		r, e = d.sess.nap.Head(url, result, err)
	case "DELETE":
		r, e = d.sess.nap.Delete(url, result, err)
	default:
		r, e = d.sess.nap.Get(url, param, result, err)
	}

	return r, e
}

func (d *Database) send(resource string, id string, method string, payload, result, err interface{}) (*nap.Response, error) {
	url := d.buildRequest(resource, id)
	var r *nap.Response
	var e error

	switch method {
	case "POST":
		r, e = d.sess.nap.Post(url, payload, result, err)
	case "PUT":
		r, e = d.sess.nap.Put(url, payload, result, err)
	case "PATCH":
		r, e = d.sess.nap.Patch(url, payload, result, err)
	}
	return r, e
}

func (db Database) buildRequest(t string, id string) string {
	var r string
	if id == "" {
		r = db.baseURL + t
	} else {
		r = db.baseURL + t + "/" + id
	}
	return r
}

type DatabaseResult struct {
	Result []string `json:"result"`
	Error  bool     `json:"error"`
	Code   int      `json:"code"`
}

// Returns Collection attached to current Database
func (db Database) Col(name string) *Collection {
	var col Collection
	var found bool
	// need to validate this more
	for _, c := range db.Collections {
		if c.Name == name {
			col = c
			col.db = &db
			found = true
			break
		}
	}

	if !found {
		// should I create the collection?
		// TODO add option to create collection if not available
		panic("Invalid Dbs")
		return nil
	}
	return &col
}