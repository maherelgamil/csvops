// Package csvops provides reusable CSV operations (split, dedupe, filter, merge,
// stats, preview, to-sqlite) as a library. Both the csvops CLI and the desktop
// app depend on this package.
package csvops

// Progress is an optional callback invoked during long-running operations.
// total may be 0 when not known in advance; done monotonically increases.
type Progress func(done, total int64)

// safeProgress is a no-op if p is nil.
func safeProgress(p Progress, done, total int64) {
	if p != nil {
		p(done, total)
	}
}
