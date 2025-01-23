package cache_serialization

// CacheSerializer used for cache_serialization serialization and deserialization
// T - target type, D - dto type
type CacheSerializer[T any, D any] interface {
	ToT(dto D) T
	ToD(plan T) D
	ToTArray(dto []D) []T
	ToDArray(plans []T) []D
}
