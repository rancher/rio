package model

import "database/sql"

type Token struct {
	ID        int64  `db:"id"`
	Token     string `db:"token"`
	Fqdn      string `db:"fqdn"`
	CreatedOn int64  `db:"created_on"`
}

type FrozenPrefix struct {
	ID        int64  `db:"id"`
	Prefix    string `db:"prefix"`
	CreatedOn int64  `db:"created_on"`
}

type RecordA struct {
	ID        int64         `db:"id"`
	Fqdn      string        `db:"fqdn"`
	Type      int           `db:"type"`
	Content   string        `db:"content"`
	CreatedOn int64         `db:"created_on"`
	UpdatedOn sql.NullInt64 `db:"updated_on"`
	TID       int64         `db:"tid"`
}

type SubRecordA struct {
	ID        int64         `db:"id"`
	Fqdn      string        `db:"fqdn"`
	Type      int           `db:"type"`
	Content   string        `db:"content"`
	CreatedOn int64         `db:"created_on"`
	UpdatedOn sql.NullInt64 `db:"updated_on"`
	PID       int64         `db:"pid"`
}

type RecordTXT struct {
	ID        int64         `db:"id"`
	Fqdn      string        `db:"fqdn"`
	Type      int           `db:"type"`
	Content   string        `db:"content"`
	CreatedOn int64         `db:"created_on"`
	UpdatedOn sql.NullInt64 `db:"updated_on"`
	TID       int64         `db:"tid"`
}

type RecordCNAME struct {
	ID        int64         `db:"id"`
	Fqdn      string        `db:"fqdn"`
	Type      int           `db:"type"`
	Content   string        `db:"content"`
	CreatedOn int64         `db:"created_on"`
	UpdatedOn sql.NullInt64 `db:"updated_on"`
	TID       int64         `db:"tid"`
}
