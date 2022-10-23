package status

type Aggregator struct {
	statuses []RolloutStatus
	fatal    *RolloutStatus
}

func (agg *Aggregator) Add(status RolloutStatus) {
	agg.statuses = append(agg.statuses, status)

	if status.Error != nil && !status.Continue {
		agg.fatal = &status
	}

}

func (agg *Aggregator) Fatal() *RolloutStatus {
	return agg.fatal
}

func (agg *Aggregator) Resolve() RolloutStatus {
	if agg.fatal != nil {
		return *agg.fatal
	}

	// if there are no errors and all are !continue, return ok
	hasErrors := false
	hasContinue := false
	for _, status := range agg.statuses {
		if status.Error != nil {
			hasErrors = true
			break
		}
		if status.Continue {
			hasContinue = true
			break
		}
	}
	if !hasErrors && !hasContinue {
		return RolloutOk()
	}

	// otherwise we have error but can continue, returning first error
	for _, status := range agg.statuses {
		if status.Error != nil {
			return status
		}
	}
	// fmt.Printf("agg: %v\n", agg)
	panic("invalid aggregator state")
}
