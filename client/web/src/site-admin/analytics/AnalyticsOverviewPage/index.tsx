import React, { useEffect, useMemo } from 'react'

import { mdiAccount, mdiSourceRepository } from '@mdi/js'
import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import { Card, H2, Link, LoadingSpinner, Text } from '@sourcegraph/wildcard'

import { ErrorBoundary } from '../../../components/ErrorBoundary'
import { OverviewStatisticsResult, OverviewStatisticsVariables } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { checkRequestAccessAllowed } from '../../../util/checkRequestAccessAllowed'
import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'
import { HorizontalSelect } from '../components/HorizontalSelect'
import { useChartFilters } from '../useChartFilters'
import { getByteUnitLabel, getByteUnitValue } from '../utils'

import { DevTimeSaved } from './DevTimeSaved'
import { OVERVIEW_STATISTICS } from './queries'
import { Sidebar } from './Sidebar'

import styles from './index.module.scss'

interface Props {}

export const AnalyticsOverviewPage: React.FunctionComponent<Props> = () => {
    const { dateRange } = useChartFilters({ name: 'Overview' })
    const { data, error, loading } = useQuery<OverviewStatisticsResult, OverviewStatisticsVariables>(
        OVERVIEW_STATISTICS,
        {}
    )
    useEffect(() => {
        eventLogger.logPageView('AdminAnalyticsOverview')
    }, [])

    const userStatisticsItems = useMemo(() => {
        if (!data) {
            return []
        }
        const items = [
            { label: 'Total users', value: data.users.totalCount },
            {
                label: 'Administrators',
                value: data.site.adminUsers.totalCount,
            },
            {
                label: 'Users licenses',
                value: data.site.productSubscription.license?.userCount || 0,
            },
        ]

        const isRequestAccessAllowed = checkRequestAccessAllowed(window.context)

        if (isRequestAccessAllowed) {
            items.push({ label: 'Pending requests', value: data.pendingAccessRequests.totalCount || 0 })
        }
        return items
    }, [data])

    if (error) {
        throw error
    }

    if (loading || !data) {
        return <LoadingSpinner />
    }

    const changelogUrl = getChangelogUrl(data.site.productVersion)
    return (
        <>
            <AnalyticsPageTitle>Overview</AnalyticsPageTitle>

            <Card className="p-3" data-testid="product-certificate">
                <div className="d-flex justify-content-between align-items-start mb-3 text-nowrap">
                    <div className="w-100">
                        <div className="d-flex">
                            <H2 className="mb-3">{data.site.productSubscription.productNameWithBrand}</H2>
                            <HorizontalSelect<typeof dateRange.value> {...dateRange} className="mb-3 ml-auto" />
                        </div>
                        <div className="d-flex">
                            <Text className="text-muted">
                                Version{' '}
                                {changelogUrl ? (
                                    <Link to={changelogUrl} className={styles.purple}>
                                        {data.site.productVersion}
                                    </Link>
                                ) : (
                                    <span className={styles.purple}>{data.site.productVersion}</span>
                                )}
                            </Text>
                        </div>
                    </div>
                </div>
                <div className={classNames('d-flex mt-3', styles.padded)}>
                    <div className={styles.main}>
                        <ErrorBoundary location={null}>
                            <DevTimeSaved
                                showAnnualProjection={data.users.totalCount > 1}
                                dateRange={dateRange.value}
                            />
                        </ErrorBoundary>
                    </div>
                    <div className={styles.sidebar}>
                        <Sidebar
                            sections={[
                                {
                                    title: 'Users statistics',
                                    icon: mdiAccount,
                                    link: '/site-admin/analytics/users',
                                    items: userStatisticsItems,
                                },
                                {
                                    title: 'Code statistics',
                                    icon: mdiSourceRepository,
                                    link: '/site-admin/repositories',
                                    items: [
                                        {
                                            label: 'Repositories',
                                            value: data.repositories.totalCount || 0,
                                        },
                                        {
                                            label: `${getByteUnitLabel(
                                                Number(data.repositoryStats.gitDirBytes)
                                            )} stored`,
                                            value: getByteUnitValue(Number(data.repositoryStats.gitDirBytes)),
                                        },
                                        {
                                            label: 'Lines of code',
                                            value: Number(data.repositoryStats.indexedLinesCount),
                                        },
                                    ],
                                },
                            ]}
                        />
                    </div>
                </div>
            </Card>
        </>
    )
}

function getChangelogUrl(version: string): string | null {
    const versionAnchor = version.replace(/\./g, '-')
    // Only show changelog link for versions that match the X.Y.Z format.
    // Other versions don't have a changelog entry.
    return version.match(/^\d+-\d+-\d+$/)
        ? `https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/CHANGELOG.md#${versionAnchor}`
        : null
}
