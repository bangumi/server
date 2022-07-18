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
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"

	"github.com/bangumi/server/internal/app"
	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/oauth"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/web/captcha"
	"github.com/bangumi/server/internal/web/frontend"
	"github.com/bangumi/server/internal/web/handler/common"
	"github.com/bangumi/server/internal/web/handler/subject"
	"github.com/bangumi/server/internal/web/rate"
	"github.com/bangumi/server/internal/web/session"
)

var errTranslationNotFound = errors.New("failed to find translation for zh")

func New(
	common common.Common,
	cfg config.AppConfig,
	c domain.CharacterService,
	p domain.PersonService,
	a domain.AuthService,
	collect domain.CollectionRepo,
	r domain.RevisionRepo,
	topic domain.TopicRepo,
	g domain.GroupRepo,
	index domain.IndexRepo,
	user domain.UserRepo,
	cache cache.Cache,
	app app.App,
	captcha captcha.Manager,
	session session.Manager,
	rateLimit rate.Manager,
	log *zap.Logger,
	engine frontend.TemplateEngine,
	oauth oauth.Manager,
) (Handler, error) {
	validate, trans, err := getValidator()
	if err != nil {
		return Handler{}, err
	}

	return Handler{
		Common:               common,
		app:                  app,
		cfg:                  cfg,
		cache:                cache,
		log:                  log.Named("web.handler"),
		rateLimit:            rateLimit,
		session:              session,
		p:                    p,
		a:                    a,
		u:                    user,
		c:                    c,
		collect:              collect,
		i:                    index,
		r:                    r,
		topic:                topic,
		captcha:              captcha,
		g:                    g,
		v:                    validate,
		validatorTranslation: trans,

		oauth:    oauth,
		template: engine,
		buffPool: buffer.NewPool(),
	}, nil
}

type Handler struct {
	common.Common
	Subject              subject.Subject
	app                  app.App
	p                    domain.PersonService
	a                    domain.AuthService
	collect              domain.CollectionRepo
	session              session.Manager
	captcha              captcha.Manager
	c                    domain.CharacterService
	u                    domain.UserRepo
	rateLimit            rate.Manager
	i                    domain.IndexRepo
	validatorTranslation ut.Translator
	g                    domain.GroupRepo
	cache                cache.Cache
	r                    domain.RevisionRepo
	oauth                oauth.Manager
	topic                domain.TopicRepo
	template             frontend.TemplateEngine
	buffPool             buffer.Pool
	log                  *zap.Logger
	v                    *validator.Validate
	cfg                  config.AppConfig
}

func getValidator() (*validator.Validate, ut.Translator, error) {
	validate := validator.New()
	uni := ut.New(en.New(), zh.New())

	// this is usually know or extracted from http 'Accept-Language' header
	// also see uni.FindTranslator(...)
	trans, found := uni.GetTranslator(zh.New().Locale())
	if !found {
		return nil, nil, errTranslationNotFound
	}

	err := zh_translations.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		return nil, nil, errgo.Wrap(err, "failed to register translation")
	}

	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		docName := field.Tag.Get("validateName")
		if docName != "" {
			return docName
		}

		tag := field.Tag.Get("json")
		if tag == "" {
			return field.Name
		}

		name := strings.SplitN(tag, ",", 2)[0]
		if name == "-" {
			return ""
		}

		return name
	})

	return validate, trans, nil
}
