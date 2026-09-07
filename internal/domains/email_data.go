package domains

type Project string

const (
	AuthSession Project = "auth-session"
	RateLimiter Project = "rate-limiter"
	URLShorten  Project = "url-shorten"
	IPGeo       Project = "ip-geolocation"
)

type EmailDataEntity struct {
	Name        string
	TestingLink string
}

type ContactEmailDataEntity struct {
	Name      string
	FromEmail string
	Message   string
}

type DemoRequestEmailDataEntity struct {
	FromEmail string
	Project   Project
	Message   string
}