package dbpq

var ValuesPoolConnection = struct {
	MaxIdleConnections int
	MaxOpenConnections int
	ConnMaxIdleTime    int
	ConnMaxLifetime    int
	ConnWithTimeout    int
}{
	MaxIdleConnections: 100,
	MaxOpenConnections: 100,
	ConnMaxIdleTime:    1,
	ConnMaxLifetime:    30,
	ConnWithTimeout:    3,
}
