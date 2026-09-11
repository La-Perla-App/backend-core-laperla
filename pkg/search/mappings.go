package search

// Default mappings for directory indexes (keyword filters + text search).
var (
	MappingDestinations = []byte(`{
  "settings": {"number_of_shards": 1, "number_of_replicas": 0},
  "mappings": {
    "properties": {
      "id": {"type": "keyword"},
      "slug": {"type": "keyword", "fields": {"text": {"type": "text"}}},
      "name": {
        "properties": {
          "es": {"type": "text"},
          "en": {"type": "text"}
        }
      },
      "is_published": {"type": "boolean"},
      "updated_at": {"type": "date"}
    }
  }
}`)

	MappingCategories = []byte(`{
  "settings": {"number_of_shards": 1, "number_of_replicas": 0},
  "mappings": {
    "properties": {
      "id": {"type": "keyword"},
      "destination_id": {"type": "keyword"},
      "slug": {"type": "keyword", "fields": {"text": {"type": "text"}}},
      "name": {
        "properties": {
          "es": {"type": "text"},
          "en": {"type": "text"}
        }
      },
      "sort_order": {"type": "integer"},
      "updated_at": {"type": "date"}
    }
  }
}`)

	MappingBusinesses = []byte(`{
  "settings": {"number_of_shards": 1, "number_of_replicas": 0},
  "mappings": {
    "properties": {
      "id": {"type": "keyword"},
      "destination_id": {"type": "keyword"},
      "category_id": {"type": "keyword"},
      "name": {"type": "text", "fields": {"keyword": {"type": "keyword"}}},
      "slug": {"type": "keyword", "fields": {"text": {"type": "text"}}},
      "description": {
        "properties": {
          "es": {"type": "text"},
          "en": {"type": "text"}
        }
      },
      "status": {"type": "keyword"},
      "address": {"type": "text"},
      "phone": {"type": "keyword"},
      "website": {"type": "keyword"},
      "is_featured": {"type": "boolean"},
      "is_verified_ally": {"type": "boolean"},
      "tags": {"type": "keyword"},
      "lat": {"type": "float"},
      "lng": {"type": "float"},
      "updated_at": {"type": "date"}
    }
  }
}`)

	MappingFeatured = []byte(`{
  "settings": {"number_of_shards": 1, "number_of_replicas": 0},
  "mappings": {
    "properties": {
      "id": {"type": "keyword"},
      "destination_id": {"type": "keyword"},
      "title": {
        "properties": {
          "es": {"type": "text"},
          "en": {"type": "text"}
        }
      },
      "kind": {"type": "keyword"},
      "ref_id": {"type": "keyword"},
      "is_published": {"type": "boolean"},
      "sort_order": {"type": "integer"},
      "updated_at": {"type": "date"}
    }
  }
}`)
)
