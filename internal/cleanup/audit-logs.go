package cleanup

import (
	"context"
	"log/slog"
	"time"

	"github.com/tubblelive/star/internal/state"
)

func PrepareAuditLogs(ctx context.Context, star *state.StarServer) {
	if err := star.Database.EnsureAuditPartitions(ctx, time.Now()); err != nil {
		slog.Error("failed to create audit log partitions", "err", err)
	}

	stranded, err := star.Database.CountDefaultAuditLogs(ctx)
	if err != nil {
		slog.Error("failed to check the default audit log partition", "err", err)
	} else if stranded > 0 {
		slog.Error("audit logs are in the default partition, partitions for their months can not be created until they are moved", "rows", stranded)
	}
}

func CleanupAuditLogs(ctx context.Context, star *state.StarServer) {
	settings, err := star.Database.GetSettings(ctx)
	if err != nil || settings == nil || settings.AuditRetentionDays == nil {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -*settings.AuditRetentionDays)

	dropped, err := star.Database.DropExpiredAuditPartitions(ctx, cutoff)
	if err != nil {
		slog.Error("failed to drop expired audit log partitions", "err", err)
		return
	}

	var deleted int64

	for {
		count, err := star.Database.DeleteExpiredAuditLogs(ctx, cutoff, 5000)
		if err != nil {
			slog.Error("failed to delete expired audit logs", "err", err)
			return
		}

		deleted += count
		if count < 5000 {
			break
		}
	}

	if dropped != 0 || deleted != 0 {
		slog.Info("purged expired audit logs", "partitions", dropped, "rows", deleted)
	}
}
