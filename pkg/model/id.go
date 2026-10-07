package model

// Prefix defines typed entity identifier prefixes (TypeID convention).
type Prefix string

const (
	PrefixSeries    Prefix = "ser"  // Series: ser_<uuidv7>
	PrefixGroup     Prefix = "grp"  // Group: grp_<uuidv7>
	PrefixColor     Prefix = "col"  // Color: col_<uuidv7>
	PrefixMember    Prefix = "mem"  // Member: mem_<uuidv7>
	PrefixSong      Prefix = "sng"  // Song: sng_<uuidv7>
	PrefixImage     Prefix = "img"  // Member Image: img_<uuidv7>
	PrefixPhotoType Prefix = "pht"  // Photo Type: pht_<uuidv7>
	PrefixQuiz      Prefix = "quiz" // Quiz Question: quiz_<uuidv7>
	PrefixUser      Prefix = "usr"  // User: usr_<uuidv7>
	PrefixAnswer    Prefix = "ans"  // Answer Log: ans_<uuidv7>
	PrefixProposal  Prefix = "prp"  // Metadata edit proposal (ADR-0037)
)

// ID represents a prefixed UUID v7 identifier (e.g., mem_018f3a5b8c9d7a1e8f2b3c4d5e6f7a8b).
// The prefix prevents accidental misuse across entity types, while UUID v7 provides
// time-sortable, globally unique surrogate keys.
type ID string
