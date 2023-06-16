package runtime

// CompatibilityDate is a string that represents the date of the last update.
type CompatibilityDate string

// CompatibilityFlag is a string that represents the name of a compatibility flag.
type CompatibilityFlag string

// CompatibilityFlags is a map of compatibility flags.
type CompatibilityFlags map[CompatibilityFlag]bool

// CompatibilityMatrix is a map of compatibility dates.
type CompatibilityMatrix map[CompatibilityDate]CompatibilityFlags
