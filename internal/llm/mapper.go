package llm

func normalizeIntent(v string) Intent {
	switch v {
	case "nearby":
		return IntentNearby
	case "category":
		return IntentCategory
	case "source":
		return IntentSource
	default:
		return IntentSearch
	}
}

func Normalize(raw *RawAnalysis) *AnalysisResult {
	entities := append(raw.People, raw.Orgs...)
	entities = append(entities, raw.Topics...)

	loc := ""
	if len(raw.Locations) > 0 {
		loc = raw.Locations[0]
	}

	analysisResult := &AnalysisResult{
		Intent:   normalizeIntent(raw.Intent),
		Entities: entities,
		Location: loc,
	}
	return analysisResult
}
