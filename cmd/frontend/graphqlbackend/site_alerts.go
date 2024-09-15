package graphqlbackend

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/hooks"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/versions"
	"github.com/sourcegraph/sourcegraph/internal/settings"
	srcprometheus "github.com/sourcegraph/sourcegraph/internal/src-prometheus"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Alert implements the GraphQL type Alert.
type Alert struct {
	TypeValue                 string
	MessageValue              string
	IsDismissibleWithKeyValue string
}

func (r *Alert) Type() string    { return r.TypeValue }
func (r *Alert) Message() string { return r.MessageValue }
func (r *Alert) IsDismissibleWithKey() *string {
	if r.IsDismissibleWithKeyValue == "" {
		return nil
	}
	return &r.IsDismissibleWithKeyValue
}

// Constants for the GraphQL enum AlertType.
const (
	AlertTypeInfo    = "INFO"
	AlertTypeWarning = "WARNING"
	AlertTypeError   = "ERROR"
)

// AlertFuncs is a list of functions called to populate the GraphQL Site.alerts value. It may be
// appended to at init time.
//
// The functions are called each time the Site.alerts value is queried, so they must not block.
var AlertFuncs []func(AlertFuncArgs) []*Alert

// AlertFuncArgs are the arguments provided to functions in AlertFuncs used to populate the GraphQL
// Site.alerts value. They allow the functions to customize the returned alerts based on the
// identity of the viewer (without needing to query for that on their own, which would be slow).
type AlertFuncArgs struct {
	IsAuthenticated     bool             // whether the viewer is authenticated
	IsSiteAdmin         bool             // whether the viewer is a site admin
	ViewerFinalSettings *schema.Settings // the viewer's final user/org/global settings
}

func (r *siteResolver) Alerts(ctx context.Context) ([]*Alert, error) {
	settings, err := settings.CurrentUserFinal(ctx, r.db)
	if err != nil {
		return nil, err
	}

	args := AlertFuncArgs{
		IsAuthenticated:     actor.FromContext(ctx).IsAuthenticated(),
		IsSiteAdmin:         auth.CheckCurrentUserIsSiteAdmin(ctx, r.db) == nil,
		ViewerFinalSettings: settings,
	}

	var alerts []*Alert
	for _, f := range AlertFuncs {
		alerts = append(alerts, f(args)...)
	}
	return alerts, nil
}

func init() {
	conf.ContributeWarning(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		if deploy.IsDeployTypeSingleDockerContainer(deploy.Type()) || deploy.IsApp() {
			return nil
		}
		if c.SiteConfig().ExternalURL == "" {
			problems = append(problems, conf.NewSiteProblem("`externalURL` is required to be set for many features of Sourcegraph to work correctly."))
		}
		return problems
	})

	// Warn if email sending is not configured.
	AlertFuncs = append(AlertFuncs, emailSendingNotConfiguredAlert)

	AlertFuncs = append(AlertFuncs, storageLimitReachedAlert)

	// Notify admins if critical alerts are firing, if Prometheus is configured.
	prom, err := srcprometheus.NewClient(srcprometheus.PrometheusURL)
	if err == nil {
		AlertFuncs = append(AlertFuncs, observabilityActiveAlertsAlert(prom))
	} else if !errors.Is(err, srcprometheus.ErrPrometheusUnavailable) {
		log15.Warn("WARNING: possibly misconfigured Prometheus", "error", err)
	}

	AlertFuncs = append(AlertFuncs, func(args AlertFuncArgs) []*Alert {
		var alerts []*Alert
		for _, notification := range conf.Get().Notifications {
			alerts = append(alerts, &Alert{
				TypeValue:                 AlertTypeInfo,
				MessageValue:              notification.Message,
				IsDismissibleWithKeyValue: notification.Key,
			})
		}
		return alerts
	})

	// Warn about invalid site configuration.
	AlertFuncs = append(AlertFuncs, func(args AlertFuncArgs) []*Alert {
		// 🚨 SECURITY: Only the site admin should care about the site configuration being invalid, as they
		// are the only one who can take action on that. Additionally, it may be unsafe to expose information
		// about the problems with the configuration (e.g. if the error message contains sensitive information).
		if !args.IsSiteAdmin {
			return nil
		}

		problems, err := conf.Validate(conf.Raw())
		if err != nil {
			return []*Alert{
				{
					TypeValue:    AlertTypeError,
					MessageValue: `Update [**site configuration**](/site-admin/configuration) to resolve problems: ` + err.Error(),
				},
			}
		}

		warnings, err := conf.GetWarnings()
		if err != nil {
			return []*Alert{
				{
					TypeValue:    AlertTypeError,
					MessageValue: `Update [**site configuration**](/site-admin/configuration) to resolve problems: ` + err.Error(),
				},
			}
		}
		problems = append(problems, warnings...)

		if len(problems) == 0 {
			return nil
		}
		alerts := make([]*Alert, 0, 2)

		siteProblems := problems.Site()
		if len(siteProblems) > 0 {
			alerts = append(alerts, &Alert{
				TypeValue: AlertTypeWarning,
				MessageValue: `[**Update site configuration**](/site-admin/configuration) to resolve problems:` +
					"\n* " + strings.Join(siteProblems.Messages(), "\n* "),
			})
		}

		externalServiceProblems := problems.ExternalService()
		if len(externalServiceProblems) > 0 {
			alerts = append(alerts, &Alert{
				TypeValue: AlertTypeWarning,
				MessageValue: `[**Update external service configuration**](/site-admin/external-services) to resolve problems:` +
					"\n* " + strings.Join(externalServiceProblems.Messages(), "\n* "),
			})
		}
		return alerts
	})

	// Warn if customer is using GitLab on a version < 12.0.
	AlertFuncs = append(AlertFuncs, gitlabVersionAlert)
}

func storageLimitReachedAlert(args AlertFuncArgs) []*Alert {
	licenseInfo := hooks.GetLicenseInfo()
	if licenseInfo == nil {
		return nil
	}

	if licenseInfo.CodeScaleCloseToLimit {
		return []*Alert{{
			TypeValue:    AlertTypeWarning,
			MessageValue: "You're about to reach the 100GiB storage limit. Upgrade to [Sourcegraph Enterprise](https://about.sourcegraph.com/pricing) for unlimited storage for your code.",
		}}
	} else if licenseInfo.CodeScaleExceededLimit {
		return []*Alert{{
			TypeValue:    AlertTypeError,
			MessageValue: "You've used all 100GiB of storage. Upgrade to [Sourcegraph Enterprise](https://about.sourcegraph.com/pricing) for unlimited storage for your code.",
		}}
	}
	return nil
}

func emailSendingNotConfiguredAlert(args AlertFuncArgs) []*Alert {
	if !args.IsSiteAdmin || deploy.IsDeployTypeSingleDockerContainer(deploy.Type()) || deploy.IsApp() {
		return nil
	}
	if conf.Get().EmailSmtp == nil || conf.Get().EmailSmtp.Host == "" {
		return []*Alert{{
			TypeValue:                 AlertTypeWarning,
			MessageValue:              "Warning: Sourcegraph cannot send emails! [Configure `email.smtp`](/help/admin/config/email) so that features such as Code Monitors, password resets, and invitations work. [documentation](/help/admin/config/email)",
			IsDismissibleWithKeyValue: "email-sending",
		}}
	}
	if conf.Get().EmailAddress == "" {
		return []*Alert{{
			TypeValue:                 AlertTypeWarning,
			MessageValue:              "Warning: Sourcegraph cannot send emails! [Configure `email.address`](/help/admin/config/email) so that features such as Code Monitors, password resets, and invitations work. [documentation](/help/admin/config/email)",
			IsDismissibleWithKeyValue: "email-sending",
		}}
	}
	return nil
}

// observabilityActiveAlertsAlert directs admins to check Grafana if critical alerts are firing
func observabilityActiveAlertsAlert(prom srcprometheus.Client) func(AlertFuncArgs) []*Alert {
	return func(args AlertFuncArgs) []*Alert {
		// true by default - change settings.schema.json if this changes
		// blocked by https://github.com/sourcegraph/sourcegraph/issues/12190
		observabilitySiteAlertsDisabled := true
		if args.ViewerFinalSettings != nil && args.ViewerFinalSettings.AlertsHideObservabilitySiteAlerts != nil {
			observabilitySiteAlertsDisabled = *args.ViewerFinalSettings.AlertsHideObservabilitySiteAlerts
		}

		if !args.IsSiteAdmin || observabilitySiteAlertsDisabled {
			return nil
		}

		// use a short timeout to avoid having this block problems from loading
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		status, err := prom.GetAlertsStatus(ctx)
		if err != nil {
			return []*Alert{{TypeValue: AlertTypeWarning, MessageValue: fmt.Sprintf("Failed to fetch alerts status: %s", err)}}
		}

		// decide whether to render a message about alerts
		if status.Critical == 0 {
			return nil
		}
		msg := fmt.Sprintf("%s across %s currently firing - [view alerts](/-/debug/grafana)",
			pluralize(status.Critical, "critical alert", "critical alerts"),
			pluralize(status.ServicesCritical, "service", "services"))
		return []*Alert{{TypeValue: AlertTypeError, MessageValue: msg}}
	}
}

func gitlabVersionAlert(args AlertFuncArgs) []*Alert {
	// We only show this alert to site admins.
	if !args.IsSiteAdmin {
		return nil
	}

	chvs, err := versions.GetVersions()
	if err != nil {
		log15.Warn("Failed to get code host versions for GitLab minimum version alert", "error", err)
		return nil
	}

	// NOTE: It's necessary to include a "-0" prerelease suffix on each constraint so that
	// prereleases of future versions are still considered to satisfy the constraint. See
	// https://github.com/Masterminds/semver#working-with-prerelease-versions for more.
	mv, err := semver.NewConstraint(">=12.0.0-0")
	if err != nil {
		log15.Warn("Failed to create minimum version constraint for GitLab minimum version alert", "error", err)
	}

	for _, chv := range chvs {
		if chv.ExternalServiceKind != extsvc.KindGitLab {
			continue
		}

		cv, err := semver.NewVersion(chv.Version)
		if err != nil {
			log15.Warn("Failed to parse code host version for GitLab minimum version alert", "error", err, "external_service_kind", chv.ExternalServiceKind)
			continue
		}

		if !mv.Check(cv) {
			log15.Debug("Detected GitLab instance running a version below 12.0.0", "version", chv.Version)

			return []*Alert{{
				TypeValue:    AlertTypeError,
				MessageValue: "One or more of your code hosts is running a version of GitLab below 12.0, which is not supported by Sourcegraph. Please upgrade your GitLab instance(s) to prevent disruption.",
			}}
		}
	}

	return nil
}

func pluralize(v int, singular, plural string) string {
	if v == 1 {
		return fmt.Sprintf("%d %s", v, singular)
	}
	return fmt.Sprintf("%d %s", v, plural)
}
