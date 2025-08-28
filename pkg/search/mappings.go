package search

const SeriesMapping = `{
	"mappings": {
		"properties": {
			"id": {"type": "keyword"},
			"title": {
				"type": "text",
				"analyzer": "standard",
				"fields": {
					"keyword": {"type": "keyword"}
				}
			},
			"description": {
				"type": "text",
				"analyzer": "standard"
			},
			"category_id": {"type": "keyword"},
			"language": {"type": "keyword"},
			"type": {"type": "keyword"},
			"created_at": {"type": "date"},
			"updated_at": {"type": "date"},
			"indexed_at": {"type": "date"}
		}
	},
	"settings": {
		"number_of_shards": 1,
		"number_of_replicas": 0,
		"analysis": {
			"analyzer": {
				"standard": {
					"type": "standard"
				}
			}
		}
	}
}`

const EpisodeMapping = `{
	"mappings": {
		"properties": {
			"id": {"type": "keyword"},
			"series_id": {"type": "keyword"},
			"uploader_id": {"type": "keyword"},
			"title": {
				"type": "text",
				"analyzer": "standard",
				"fields": {
					"keyword": {"type": "keyword"}
				}
			},
			"description": {
				"type": "text",
				"analyzer": "standard"
			},
			"duration_seconds": {"type": "integer"},
			"publish_date": {"type": "date"},
			"transcript_url": {"type": "keyword"},
			"mime_type": {"type": "keyword"},
			"size_bytes": {"type": "long"},
			"created_at": {"type": "date"},
			"updated_at": {"type": "date"},
			"indexed_at": {"type": "date"}
		}
	},
	"settings": {
		"number_of_shards": 1,
		"number_of_replicas": 0,
		"analysis": {
			"analyzer": {
				"standard": {
					"type": "standard"
				}
			}
		}
	}
}`
