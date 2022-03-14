// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
//
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

package handler

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/strparse"
	"github.com/bangumi/server/model"
)

func parseSubjectID(s string) (model.SubjectIDType, error) {
	if s == "" {
		return 0, nil
	}

	subjectID, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, fiber.NewError(http.StatusBadRequest, "bad subject id: "+strconv.Quote(s))
	}

	return model.SubjectIDType(subjectID), nil
}

func parseSubjectType(s string) (uint8, error) {
	if s == "" {
		return 0, nil
	}

	t, err := strparse.Uint8(s)
	if err != nil {
		return 0, fiber.NewError(http.StatusBadRequest, "bad subject type: "+strconv.Quote(s))
	}

	switch t {
	case model.SubjectAnime, model.SubjectBook,
		model.SubjectMusic, model.SubjectReal, model.SubjectGame:
		return t, nil
	}

	return 0, fiber.NewError(http.StatusBadRequest, strconv.Quote(s)+"is not a valid subject type")
}
