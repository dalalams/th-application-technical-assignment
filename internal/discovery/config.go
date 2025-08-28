package discovery

import (
	"th-application-technical-assignment/pkg/http"
	"th-application-technical-assignment/pkg/search"
	"th-application-technical-assignment/pkg/telemetry"
)

type Config struct {
	Search    search.Config    `envPrefix:"OPENSEARCH_"`
	Telemetry telemetry.Config `envPrefix:"OTEL_"`
	HTTP      http.Config      `envPrefix:"HTTP_"`
}
