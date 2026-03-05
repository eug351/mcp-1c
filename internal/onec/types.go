package onec

// MetadataTree represents the metadata tree of a 1C configuration.
type MetadataTree struct {
	Catalogs  []string `json:"Справочники"`
	Documents []string `json:"Документы"`
	Registers []string `json:"Регистры"`
}
