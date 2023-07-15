package dbfinder

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func TripSelect(limit *int, after *model.Cursor, ids []int, active bool, permFilter *model.PermFilter, where *model.TripFilter) sq.SelectBuilder {
	q := sq.StatementBuilder.Select(
		"gtfs_trips.*",
		"current_feeds.id AS feed_id",
		"current_feeds.onestop_id AS feed_onestop_id",
		"feed_versions.sha1 AS feed_version_sha1",
	).
		From("gtfs_trips").
		Join("feed_versions ON feed_versions.id = gtfs_trips.feed_version_id").
		Join("current_feeds ON current_feeds.id = feed_versions.feed_id").
		OrderBy("gtfs_trips.feed_version_id,gtfs_trips.id")

	if where != nil {
		if where.StopPatternID != nil {
			q = q.Where(sq.Eq{"stop_pattern_id": where.StopPatternID})
		}
		if len(where.RouteOnestopIds) > 0 {
			q = q.Join("tl_route_onestop_ids tlros on tlros.route_id = gtfs_trips.route_id and tlros.feed_version_id = gtfs_trips.feed_version_id")
			q = q.Where(sq.Eq{"tlros.onestop_id": where.RouteOnestopIds})
		}
		if where.FeedVersionSha1 != nil {
			q = q.Where(sq.Eq{"feed_versions.sha1": *where.FeedVersionSha1})
		}
		if where.FeedOnestopID != nil {
			q = q.Where(sq.Eq{"current_feeds.onestop_id": *where.FeedOnestopID})
		}
		if where.TripID != nil {
			q = q.Where(sq.Eq{"gtfs_trips.trip_id": *where.TripID})
		}
		if len(where.RouteIds) > 0 {
			q = q.Where(sq.Eq{"gtfs_trips.route_id": where.RouteIds})
		}

		if where.ServiceDate != nil {
			serviceDate := where.ServiceDate.Val
			q = q.JoinClause(`
			join lateral (
				select gc.id
				from gtfs_calendars gc 
				left join gtfs_calendar_dates gcda on gcda.service_id = gc.id and gcda.exception_type = 1 and gcda.date = ?::date
				left join gtfs_calendar_dates gcdb on gcdb.service_id = gc.id and gcdb.exception_type = 2 and gcdb.date = ?::date
				where 
					gc.id = gtfs_trips.service_id 
					AND ((
						gc.start_date <= ?::date AND gc.end_date >= ?::date
						AND (CASE EXTRACT(isodow FROM ?::date)
						WHEN 1 THEN monday = 1
						WHEN 2 THEN tuesday = 1
						WHEN 3 THEN wednesday = 1
						WHEN 4 THEN thursday = 1
						WHEN 5 THEN friday = 1
						WHEN 6 THEN saturday = 1
						WHEN 7 THEN sunday = 1
						END)
					) OR gcda.date IS NOT NULL)
					AND gcdb.date is null
				LIMIT 1
			) gc on true
			`, serviceDate, serviceDate, serviceDate, serviceDate, serviceDate)
		}
		// Handle license filtering
		q = licenseFilter(where.License, q)
	}
	if active {
		q = q.Join("feed_states on feed_states.feed_version_id = gtfs_trips.feed_version_id")
	}
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"gtfs_trips.id": ids})
	}

	// Handle cursor
	if after != nil && after.Valid && after.ID > 0 {
		if after.FeedVersionID == 0 {
			q = q.Where(sq.Expr("(gtfs_trips.feed_version_id, gtfs_trips.id) > (select feed_version_id,id from gtfs_trips where id = ?)", after.ID))
		} else {
			q = q.Where(sq.Expr("(gtfs_trips.feed_version_id, gtfs_trips.id) > (?,?)", after.FeedVersionID, after.ID))
		}
	}

	// Handle permissions
	q = q.
		Join("feed_states fsp on fsp.feed_id = current_feeds.id").
		Where(sq.Or{
			sq.Expr("fsp.public = true"),
			sq.Eq{"fsp.feed_id": permFilter.GetAllowedFeeds()},
			sq.Eq{"feed_versions.id": permFilter.GetAllowedFeedVersions()},
		})

	// Outer query
	qView := sq.StatementBuilder.Select("t.*").FromSelect(q, "t").Limit(checkLimit(limit))
	return qView
}