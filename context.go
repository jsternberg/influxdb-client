package influxdb

import "context"

type contextKey int

const (
	dbOption contextKey = iota
	rpOption
)

// WithDB sets the database to use for a write or query.
func WithDB(ctx context.Context, db string) context.Context {
	return context.WithValue(ctx, dbOption, db)
}

// DBFromContext retrieves the database from the context.
func DBFromContext(ctx context.Context) string {
	db, _ := ctx.Value(dbOption).(string)
	return db
}

// WithRP sets the retention policy for a write.
func WithRP(ctx context.Context, rp string) context.Context {
	return context.WithValue(ctx, rpOption, rp)
}

// RPFromContext retrieves the retention policy from the context.
func RPFromContext(ctx context.Context) string {
	rp, _ := ctx.Value(rpOption).(string)
	return rp
}

// WithDBAndRP sets the database and retention policy for a write.
func WithDBAndRP(ctx context.Context, db, rp string) context.Context {
	return WithRP(WithDB(ctx, db), rp)
}
