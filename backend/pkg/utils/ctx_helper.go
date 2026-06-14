package utils

import "context"

// GetUserIDFromCtx extracts the user_id from the context.
// Returns 1 (Super Admin/System ID) as a fallback if not found.
func GetUserIDFromCtx(ctx context.Context) uint64 {
	if val := ctx.Value("user_id"); val != nil {
		if uid, ok := val.(uint64); ok {
			return uid
		}
	}
	return 1
}
