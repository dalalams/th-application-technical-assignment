package cms

import (
	"th-application-technical-assignment/pkg/auth"
	"th-application-technical-assignment/pkg/database"
	"th-application-technical-assignment/pkg/http"
	"th-application-technical-assignment/pkg/storage"
	"th-application-technical-assignment/pkg/tasks"
	"th-application-technical-assignment/pkg/telemetry"
)

type Config struct {
	Database  database.Config   `envPrefix:"DB_"`
	Redis     tasks.RedisConfig `envPrefix:"REDIS_"`
	Queue     tasks.QueueConfig `envPrefix:"QUEUE_"`
	MinIO     storage.Config    `envPrefix:"MINIO_"`
	Auth      auth.Config       `envPrefix:"AUTH_"`
	Telemetry telemetry.Config  `envPrefix:"OTEL_"`
	HTTP      http.Config       `envPrefix:"HTTP_"`
}
