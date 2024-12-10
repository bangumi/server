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

package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// https://github.com/uber-go/zap/issues/654

// make RequestKey and unique.
type key string

//nolint:gochecknoglobals
const RequestKey key = "logger.contextKey"

type RequestTrace struct {
	IP    string
	ReqID string
	Path  string
}

func (r *RequestTrace) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("ip", r.IP)
	enc.AddString("request-id", r.ReqID)
	enc.AddString("path", r.Path)

	return nil
}

func Ctx(ctx context.Context) zap.Field {
	req := ctx.Value(RequestKey)

	if req == nil {
		return zap.Skip()
	}

	r, ok := req.(*RequestTrace)
	if !ok {
		return zap.Skip()
	}

	return zap.Object("request", r)
}
