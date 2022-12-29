// SPDX-License-Identifier: AGPL-3.0-only
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>

package dal

import (
	"github.com/bangumi/server/dal/query"
)

type Transaction interface {
	Transaction(fc func(tx *query.Query) error) error
}

func Tx(tx Transaction, fc func(tx *query.Query) error) error {
	return tx.Transaction(fc)
}

var _ Transaction = NoopTransaction{}
var _ Transaction = MysqlTransaction{}

func NewMysqlTransaction(q *query.Query) Transaction {
	return MysqlTransaction{q: q}
}

type MysqlTransaction struct {
	q *query.Query
}

func (t MysqlTransaction) Transaction(fc func(tx *query.Query) error) error {
	return t.q.Transaction(fc)
}

type NoopTransaction struct {
}

func (t NoopTransaction) Transaction(fc func(tx *query.Query) error) error {
	return fc(nil)
}
