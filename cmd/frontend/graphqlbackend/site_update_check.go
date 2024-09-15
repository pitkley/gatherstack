package graphqlbackend

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/updatecheck"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

func (r *siteResolver) UpdateCheck(_ context.Context) (*updateCheckResolver, error) {
	return &updateCheckResolver{
		last: &updatecheck.Status{
			Date:          time.Time{},
			UpdateVersion: "",
		},
	}, nil
}

type updateCheckResolver struct {
	last    *updatecheck.Status
	pending bool
}

func (r *updateCheckResolver) Pending() bool { return r.pending }

func (r *updateCheckResolver) CheckedAt() *gqlutil.DateTime {
	if r.last == nil {
		return nil
	}
	return &gqlutil.DateTime{Time: r.last.Date}
}

func (r *updateCheckResolver) ErrorMessage() *string {
	if r.last == nil || r.last.Err == nil {
		return nil
	}
	s := r.last.Err.Error()
	return &s
}

func (r *updateCheckResolver) UpdateVersionAvailable() *string {
	return nil
}
