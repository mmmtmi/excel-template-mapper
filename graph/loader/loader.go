package loaders

// import graph gophers with your other imports
import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/graph-gophers/dataloader/v7"
	"github.com/mmmtmi/excel-template-mapper/graph/model"
)

type ctxKey string

const (
	loadersKey = ctxKey("dataloaders")
)

// ruleReader reads rules from a database
type ruleReader struct {
	db *sql.DB
}

// getRules implements a batch function that can retrieve many rules by ID,
// for use in a dataloader
func (u *ruleReader) getRules(ctx context.Context, ruleIds []string) []*dataloader.Result[*model.Rule] {
	stmt, err := u.db.PrepareContext(ctx, `SELECT id, name FROM rules WHERE id IN (?`+strings.Repeat(",?", len(ruleIds)-1)+`)`)
	if err != nil {
		return handleError[*model.Rule](len(ruleIds), err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, ruleIds)
	if err != nil {
		return handleError[*model.Rule](len(ruleIds), err)
	}
	defer rows.Close()

	result := make([]*dataloader.Result[*model.Rule], 0, len(ruleIds))
	for rows.Next() {
		var rule model.Rule
		if err := rows.Scan(&rule.ID); err != nil {
			result = append(result, &dataloader.Result[*model.Rule]{Error: err})
			continue
		}
		result = append(result, &dataloader.Result[*model.Rule]{Data: &rule})
	}
	return result
}

// handleError creates array of result with the same error repeated for as many items requested
func handleError[T any](itemsLength int, err error) []*dataloader.Result[T] {
	result := make([]*dataloader.Result[T], itemsLength)
	for i := 0; i < itemsLength; i++ {
		result[i] = &dataloader.Result[T]{Error: err}
	}
	return result
}

// Loaders wrap your data loaders to inject via middleware
type Loaders struct {
	ruleLoader *dataloader.Loader[string, *model.Rule]
}

// NewLoaders instantiates data loaders for the middleware
func NewLoaders(conn *sql.DB) *Loaders {
	// define the data loader
	ur := &ruleReader{db: conn}
	return &Loaders{
		ruleLoader: dataloader.NewBatchedLoader(ur.getRules, dataloader.WithWait[string, *model.Rule](time.Millisecond)),
	}
}

// Middleware injects data loaders into the context
func Middleware(conn *sql.DB, next http.Handler) http.Handler {
	// return a middleware that injects the loader to the request context
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loader := NewLoaders(conn)
		r = r.WithContext(context.WithValue(r.Context(), loadersKey, loader))
		next.ServeHTTP(w, r)
	})
}

// For returns the dataloader for a given context
func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}

// GetRule returns single rule by id efficiently
func GetRule(ctx context.Context, ruleID string) (*model.Rule, error) {
	loaders := For(ctx)
	return loaders.ruleLoader.Load(ctx, ruleID)()
}

// GetRules returns many rules by ids efficiently
func GetRules(ctx context.Context, ruleIDs []string) ([]*model.Rule, []error) {
	loaders := For(ctx)
	return loaders.ruleLoader.LoadMany(ctx, ruleIDs)()
}
