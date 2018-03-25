package kad

const bootstrapSearchContactNbr = 10

// SearchDecision x
type SearchDecision struct {
	pOnliner *ContactOnliner
}

func (sd *SearchDecision) start(pOnliner *ContactOnliner) {
	sd.pOnliner = pOnliner
}

func (sd *SearchDecision) newSearch(pSearch *Search) []*Contact {
	return sd.pOnliner.getSearchContacts(pSearch)
}
