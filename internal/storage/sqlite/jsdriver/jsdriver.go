//go:build js

package jsdriver

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strconv"
	"syscall/js"
	"time"
)

const (
	ConstraintUnique     = 2067
	ConstraintPrimaryKey = 1555

	columnTypeInteger = 1
	columnTypeFloat   = 2
	columnTypeText    = 3
	columnTypeBlob    = 4
)

var errOpenUnsupported = errors.New("jsdriver: use sql.OpenDB with a Connector")

type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

type Connector struct {
	capi js.Value
	db   js.Value
}

func NewConnector(sqlite3, db js.Value) *Connector {
	return &Connector{capi: sqlite3.Get("capi"), db: db}
}

func (c *Connector) Connect(context.Context) (driver.Conn, error) {
	return &conn{capi: c.capi, db: c.db}, nil
}

func (c *Connector) Driver() driver.Driver {
	return jsDriver{}
}

type jsDriver struct{}

func (jsDriver) Open(string) (driver.Conn, error) {
	return nil, errOpenUnsupported
}

type conn struct {
	capi js.Value
	db   js.Value
}

func (c *conn) call(fn func() js.Value) (result js.Value, err error) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}
		jsErr, ok := recovered.(js.Error)
		if !ok {
			panic(recovered)
		}
		err = c.errorFrom(jsErr)
	}()
	return fn(), nil
}

func (c *conn) errorFrom(jsErr js.Error) error {
	code := 0
	if !jsErr.Value.Get("resultCode").IsUndefined() {
		code = c.capi.Call("sqlite3_extended_errcode", c.db.Get("pointer")).Int()
	}
	return &Error{Code: code, Message: jsErr.Value.Get("message").String()}
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	return &stmt{conn: c, query: query}, nil
}

func (c *conn) Close() error {
	return nil
}

func (c *conn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *conn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	if err := c.execScript(`BEGIN IMMEDIATE`); err != nil {
		return nil, err
	}
	return &tx{conn: c}, nil
}

func (c *conn) execScript(script string) error {
	_, err := c.call(func() js.Value { return c.db.Call("exec", script) })
	return err
}

func (c *conn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if len(args) == 0 {
		if err := c.execScript(query); err != nil {
			return nil, err
		}
	} else {
		rows, err := c.query(query, args)
		if err != nil {
			return nil, err
		}
		for {
			if err := rows.Next(nil); err != nil {
				rows.Close()
				if errors.Is(err, io.EOF) {
					break
				}
				return nil, err
			}
		}
	}

	changes, err := c.call(func() js.Value { return c.db.Call("changes") })
	if err != nil {
		return nil, err
	}
	lastInsertId, err := c.call(func() js.Value { return c.capi.Call("sqlite3_last_insert_rowid", c.db.Get("pointer")) })
	if err != nil {
		return nil, err
	}
	id, err := strconv.ParseInt(js.Global().Call("String", lastInsertId).String(), 10, 64)
	if err != nil {
		return nil, err
	}

	return result{lastInsertId: id, rowsAffected: int64(changes.Int())}, nil
}

func (c *conn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return c.query(query, args)
}

func (c *conn) query(query string, args []driver.NamedValue) (*rows, error) {
	prepared, err := c.call(func() js.Value { return c.db.Call("prepare", query) })
	if err != nil {
		return nil, err
	}

	if len(args) > 0 {
		bindings, err := namedBindings(args)
		if err != nil {
			prepared.Call("finalize")
			return nil, err
		}
		if _, err := c.call(func() js.Value { return prepared.Call("bind", bindings) }); err != nil {
			prepared.Call("finalize")
			return nil, err
		}
	}

	return &rows{conn: c, stmt: prepared}, nil
}

func namedBindings(args []driver.NamedValue) (js.Value, error) {
	bindings := js.Global().Get("Object").New()
	for _, arg := range args {
		value, err := toJS(arg.Value)
		if err != nil {
			return js.Value{}, err
		}
		bindings.Set("$"+strconv.Itoa(arg.Ordinal), value)
	}
	return bindings, nil
}

func toJS(value driver.Value) (any, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case int64:
		if v > 1<<53-1 || v < -(1<<53-1) {
			return js.Global().Call("BigInt", strconv.FormatInt(v, 10)), nil
		}
		return float64(v), nil
	case float64, string:
		return v, nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case []byte:
		array := js.Global().Get("Uint8Array").New(len(v))
		js.CopyBytesToJS(array, v)
		return array, nil
	case time.Time:
		return v.Format(time.RFC3339Nano), nil
	default:
		return nil, fmt.Errorf("jsdriver: unsupported bind type %T", value)
	}
}

type stmt struct {
	conn  *conn
	query string
}

func (s *stmt) Close() error {
	return nil
}

func (s *stmt) NumInput() int {
	return -1
}

func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	return s.conn.ExecContext(context.Background(), s.query, positional(args))
}

func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.conn.query(s.query, positional(args))
}

func positional(args []driver.Value) []driver.NamedValue {
	named := make([]driver.NamedValue, len(args))
	for i, value := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: value}
	}
	return named
}

type rows struct {
	conn    *conn
	stmt    js.Value
	columns []string
}

func (r *rows) Columns() []string {
	if r.columns == nil {
		count := r.stmt.Get("columnCount").Int()
		r.columns = make([]string, count)
		for i := range count {
			r.columns[i] = r.stmt.Call("getColumnName", i).String()
		}
	}
	return r.columns
}

func (r *rows) Close() error {
	if r.stmt.IsUndefined() {
		return nil
	}
	_, err := r.conn.call(func() js.Value { return r.stmt.Call("finalize") })
	r.stmt = js.Undefined()
	return err
}

func (r *rows) Next(dest []driver.Value) error {
	hasRow, err := r.conn.call(func() js.Value { return r.stmt.Call("step") })
	if err != nil {
		return err
	}
	if !hasRow.Bool() {
		return io.EOF
	}

	for i := range dest {
		value, err := r.column(i)
		if err != nil {
			return err
		}
		dest[i] = value
	}
	return nil
}

func (r *rows) column(i int) (driver.Value, error) {
	columnType := r.conn.capi.Call("sqlite3_column_type", r.stmt.Get("pointer"), i).Int()
	switch columnType {
	case columnTypeInteger:
		return strconv.ParseInt(r.stmt.Call("get", i, columnTypeText).String(), 10, 64)
	case columnTypeFloat:
		return r.stmt.Call("get", i).Float(), nil
	case columnTypeText:
		return r.stmt.Call("get", i).String(), nil
	case columnTypeBlob:
		blob := r.stmt.Call("getBlob", i)
		data := make([]byte, blob.Length())
		js.CopyBytesToGo(data, blob)
		return data, nil
	default:
		return nil, nil
	}
}

type tx struct {
	conn *conn
}

func (t *tx) Commit() error {
	return t.conn.execScript(`COMMIT`)
}

func (t *tx) Rollback() error {
	return t.conn.execScript(`ROLLBACK`)
}

type result struct {
	lastInsertId int64
	rowsAffected int64
}

func (r result) LastInsertId() (int64, error) {
	return r.lastInsertId, nil
}

func (r result) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}
