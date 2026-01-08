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

package dam_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/internal/pkg/dam"
)

func TestAllPrintableChar(t *testing.T) {
	t.Parallel()

	require.True(t, dam.AllPrintableChar("0123456789abcdEfg\t、\n 汉字"))
	require.True(t, dam.AllPrintableChar("abc\r\nabc"))

	require.False(t, dam.AllPrintableChar("\u202c"))
	require.False(t, dam.AllPrintableChar("\u202d"))
	// 普通emoji
	require.True(t, dam.AllPrintableChar("\U0001f436"))
}

func TestDam_NeedReview(t *testing.T) {
	t.Parallel()

	d, err := dam.New(config.AppConfig{
		NsfwWord:     "",
		DisableWords: "假身份证|代办",
		BannedDomain: "lista.cc|snapmail.cc|ashotmail.com",
	})

	require.NoError(t, err)

	require.True(t, d.NeedReview("1 x 代办 xx1 "))

	require.True(t, d.CensoredWords("https://lista.cc/"))
}
