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

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bangumi/server/cmd/archive"
	"github.com/bangumi/server/cmd/canal"
	"github.com/bangumi/server/cmd/web"
)

var Root = cobra.Command{
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd:   false,
		DisableNoDescFlag:   false,
		DisableDescriptions: false,
		HiddenDefaultCmd:    true,
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	Root.PersistentFlags().String("config", "", "config file location")
	Root.AddCommand(canal.Command, web.Command, archive.Command)
}
