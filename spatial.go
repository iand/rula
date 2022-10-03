package rula

// A Length represents the linear distance between two points
// as an int64 millimetre count
type Length int64

const (
	Millimetre Length = 1
	Centimetre        = 10 * Millimetre
	Metre             = 100 * Centimetre
	Kilometre         = 1000 * Metre
	Yard              = 914 * Millimetre
	Mile              = 1609344 * Millimetre
)

type Position struct {
	East, North Length // distances from centre of map
}

// A Location is a physical location that can be occupied by an agent
type Location struct {
	id  int64
	pos Position
}

// Connection is a link between two locations, such as a road, river or sea route
type Connection struct {
	id       int64
	from     *Location
	to       *Location
	distance Length
	// Difficulty float64 // 0 is best conditions, e.g. well maintained highway
}

type Network interface {
	// Location returns the location with the given ID if it exists
	// in the network, and nil otherwise.
	Location(id int64) Location

	// Locations returns all the locations in the network.
	Locations() []Location

	// Connection returns all the connections between a and b in the network.
	Connection(a, b int64) []Connection
}
