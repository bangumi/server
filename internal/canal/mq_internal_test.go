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

package canal

import (
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/pkg/logger"
)

func TestOnSubjectChange(t *testing.T) {
	t.Parallel()
	session := mocks.NewSessionManager(t)

	c, err := config.NewAppConfig()
	require.NoError(t, err)

	search := mocks.NewSearchClient(t)

	eh := &eventHandler{
		config:  c,
		session: session,
		reader:  nil,
		search:  search,
		log:     logger.Named("eventHandler"),
	}

	err = eh.onMessage(kafka.Message{})
	require.NoError(t, err)
}
