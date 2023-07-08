package mongodb

func mapKey(field string, keys ...string) string {
	for _, key := range keys {
		field += "." + key
	}
	return field
}
