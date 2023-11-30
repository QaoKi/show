package mdb

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
)

const (
	Mdb_type_default = 0
	Mdb_type_slave   = 1
	Mdb_type_reserve = 2

	log_limit_len = 60
)

func Dbg(since time.Duration, err error, format string, a ...interface{}) {
	// if fmt.Sprintf("%v", err) == "not found" {
	// 	return
	// }
	// if err != nil || since > 500*time.Millisecond {
	// 	log.Dbg(format+fmt.Sprintf(",耗时:%.3f,err:%v", float64(since)/1e6, err), a...)
	// } else {
	// 	log.Dbg(format+fmt.Sprintf(",耗时:%.3f,err:%v", float64(since)/1e6, err), a...)
	// }
}

type Database struct {
	*mgo.Database
	tp int
}

type Collection struct {
	*mgo.Collection
}

type Query struct {
	*mgo.Query
	dbname   string
	query    interface{}
	selector interface{}
}

type Pipe struct {
	*mgo.Pipe
	dbname   string
	pipeline interface{}
}

/***********************************/
func (db *Database) Type() int { return db.tp }
func (db *Database) SetType(tp int) {
	db.tp = tp
}

func (db *Database) C(name string) *Collection {
	return &Collection{Collection: db.Database.C(name)}
}

/***********************************/
func (c *Collection) Update(selector interface{}, update interface{}) error {
	start := time.Now()
	err := c.Collection.Update(selector, update)
	Dbg(time.Since(start), err, "[MGO],update,%v,%v,%v",
		c.FullName, selector, update)
	return err
}

func (c *Collection) UpdateAll(selector interface{}, update interface{}) error {
	start := time.Now()
	info, err := c.Collection.UpdateAll(selector, update)
	Dbg(time.Since(start), err, "[MGO],updateAll,%v,%v,%v,%v",
		c.FullName, selector, update, info)
	return err
}

func (c *Collection) Upsert(selector interface{}, update interface{}) (info *mgo.ChangeInfo, err error) {
	start := time.Now()
	info, err = c.Collection.Upsert(selector, update)
	updateInfo := fmt.Sprintf("%v", update)
	if len(updateInfo) > log_limit_len && err == nil {
		updateInfo = fmt.Sprintf("updateInfo_size_%v", len(updateInfo))
	}
	Dbg(time.Since(start), err, "[MGO],upsert,%v,%v,%v,%v",
		c.FullName, selector, updateInfo, info)
	return info, err
}

func (c *Collection) Insert(docs ...interface{}) error {
	start := time.Now()
	err := c.Collection.Insert(docs...)
	Dbg(time.Since(start), err, "[MGO],insert,%v,%v",
		c.FullName, docs)
	return err
}

func (c *Collection) Remove(selector interface{}) error {
	start := time.Now()
	err := c.Collection.Remove(selector)
	Dbg(time.Since(start), err, "[MGO],remove,%v,%v",
		c.FullName, selector)
	return err
}

func (c *Collection) Find(query interface{}) *Query {
	q := c.Collection.Find(query)
	return &Query{Query: q, dbname: c.FullName, query: query}
}

func (c *Collection) Pipe(pipline interface{}) *Pipe {
	p := c.Collection.Pipe(pipline)
	return &Pipe{Pipe: p, dbname: c.FullName, pipeline: pipline}
}

/***********************************/
func (q *Pipe) All(result interface{}) (err error) {
	start := time.Now()
	err = q.Pipe.All(result)
	Dbg(time.Since(start), err, "[MGO],pipe all,%v,%v",
		q.dbname, q.pipeline)
	return err
}

/***********************************/
func (q *Query) Select(selector interface{}) *Query {
	q.selector = selector
	q.Query.Select(selector)
	return q
}

func (q *Query) All(result interface{}) (err error) {
	start := time.Now()
	err = q.Query.All(result)
	Dbg(time.Since(start), err, "[MGO],all,%v,%v,%v",
		q.dbname, q.query, q.selector)
	return err
}

func (q *Query) One(result interface{}) (err error) {
	start := time.Now()
	err = q.Query.One(result)
	Dbg(time.Since(start), err, "[MGO],one,%v,%v,%v",
		q.dbname, q.query, q.selector)
	return err
}

func (q *Query) Count() (n int, err error) {
	start := time.Now()
	n, err = q.Query.Count()
	Dbg(time.Since(start), err, "[MGO],count,%v,%v,%v",
		q.dbname, q.query, q.selector)
	return n, err
}

func (q *Query) Apply(change mgo.Change, result interface{}) (info *mgo.ChangeInfo, err error) {
	start := time.Now()
	info, err = q.Query.Apply(change, result)
	changeInfo := fmt.Sprintf("%v", change)
	if len(changeInfo) > log_limit_len && err == nil {
		changeInfo = fmt.Sprintf("changeInfo_size_%v", len(changeInfo))
	}
	Dbg(time.Since(start), err, "[MGO],apply,%v,%v,%v,%v",
		q.dbname, q.query, changeInfo, info)
	return info, err
}
