package graphqlbackend

import (
	"strings"
	"testing"

	srcprometheus "github.com/sourcegraph/sourcegraph/internal/src-prometheus"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestObservabilityActiveAlertsAlert(t *testing.T) {
	f := false
	type args struct {
		args AlertFuncArgs
	}
	tests := []struct {
		name string
		args args
		want []*Alert
	}{
		{
			name: "do not show anything for non-admin",
			args: args{
				args: AlertFuncArgs{
					IsSiteAdmin: false,
					ViewerFinalSettings: &schema.Settings{
						AlertsHideObservabilitySiteAlerts: &f,
					},
				},
			},
			want: nil,
		},
		{
			name: "prometheus unreachable for admin",
			args: args{
				args: AlertFuncArgs{
					IsSiteAdmin: true,
					ViewerFinalSettings: &schema.Settings{
						AlertsHideObservabilitySiteAlerts: &f,
					},
				},
			},
			want: []*Alert{{
				TypeValue:    AlertTypeWarning,
				MessageValue: "Failed to fetch alerts status",
			}},
		},
		{
			// blocked by https://github.com/sourcegraph/sourcegraph/issues/12190
			// see observabilityActiveAlertsAlert docstrings
			name: "alerts disabled by default for admin",
			args: args{
				args: AlertFuncArgs{
					IsSiteAdmin: true,
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// observabilityActiveAlertsAlert does not report NewClient errors,
			// the prometheus validator does
			prom, err := srcprometheus.NewClient("http://no-prometheus:9090")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			fn := observabilityActiveAlertsAlert(prom)
			gotAlerts := fn(tt.args.args)
			if len(gotAlerts) != len(tt.want) {
				t.Errorf("expected %+v, got %+v", tt.want, gotAlerts)
				return
			}
			// test for message substring equality
			for i, got := range gotAlerts {
				want := tt.want[i]
				if got.TypeValue != want.TypeValue || got.IsDismissibleWithKeyValue != want.IsDismissibleWithKeyValue || !strings.Contains(got.MessageValue, want.MessageValue) {
					t.Errorf("expected %+v, got %+v", want, got)
				}
			}
		})
	}
}
