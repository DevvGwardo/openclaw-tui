package tui

// mergeStreamContent accepts either token deltas or cumulative snapshots and
// returns the updated full content plus only the newly-added suffix.
func mergeStreamContent(current, incoming string) (next, delta string) {
	if incoming == "" {
		return current, ""
	}

	if current == "" {
		return incoming, incoming
	}

	if len(incoming) >= len(current) && incoming[:len(current)] == current {
		return incoming, incoming[len(current):]
	}

	if len(current) >= len(incoming) && current[:len(incoming)] == incoming {
		return current, ""
	}

	maxOverlap := len(current)
	if len(incoming) < maxOverlap {
		maxOverlap = len(incoming)
	}
	for overlap := maxOverlap; overlap > 0; overlap-- {
		if current[len(current)-overlap:] == incoming[:overlap] {
			suffix := incoming[overlap:]
			return current + suffix, suffix
		}
	}

	return current + incoming, incoming
}
