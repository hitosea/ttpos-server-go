module ttpos-server-go/tests

go 1.23

require (
	github.com/go-sql-driver/mysql v1.8.1
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.6.0
	github.com/hitosea/wiremock v0.0.0
)

require filippo.io/edwards25519 v1.1.0 // indirect

// Use phona/wiremock which has module name github.com/hitosea/wiremock
replace github.com/hitosea/wiremock => github.com/phona/wiremock v0.0.0-20260310141106-6dd8831fdabc
