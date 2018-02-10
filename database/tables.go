package database

var createPlanesSQL = `
CREATE TABLE "planes" (
	id INTEGER PRIMARY KEY,
	icao VARCHAR(6) NOT NULL,
	latitude REAL NOT NULL,
	longitude REAL NOT NULL,
	flight_number VARCHAR(255) NULL,
	transponder_code INTEGER NULL,
	altitude INTEGER NULL,
	speed INTEGER NULL,
	heading INTEGER NULL,
	time INTEGER NOT NULL
)
`
