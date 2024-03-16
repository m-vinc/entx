package entx

import (
	"embed"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

//go:embed templates/*
var templatesFS embed.FS

var _ entc.Extension = (*Entx)(nil)

type Entx struct {
	postgres  bool
	queryable bool
	tx        bool
}

type Options struct {
	Postgres  bool
	Queryable bool
	Tx        bool
}

func New(options *Options) entc.Extension {
	if options == nil {
		options = &Options{
			Tx:        true,
			Postgres:  true,
			Queryable: true,
		}
	}

	return &Entx{
		postgres:  options.Postgres,
		queryable: options.Queryable,
		tx:        options.Tx,
	}
}

func (entx *Entx) Hooks() []gen.Hook {
	return []gen.Hook{}
}

func (entx *Entx) Annotations() []entc.Annotation {
	return []entc.Annotation{}
}

func (entx *Entx) Templates() []*gen.Template {
	tpls := []*gen.Template{}

	if entx.postgres {
		tpls = append(tpls, gen.MustParse(gen.NewTemplate("").ParseFS(templatesFS, "templates/postgres.tmpl")))
	}

	if entx.tx {
		tpls = append(tpls, gen.MustParse(gen.NewTemplate("").ParseFS(templatesFS, "templates/tx.tmpl")))
	}

	if entx.queryable {
		tpls = append(tpls, gen.MustParse(gen.NewTemplate("").ParseFS(templatesFS, "templates/queryable.tmpl")))
	}

	return tpls
}

func (entx *Entx) Options() []entc.Option {
	return []entc.Option{}
}
