package whisper

// The AggregationMethod type describes how values are aggregated from one Whisper archive to another.
type AggregationMethod uint32

// Valid aggregation methods
const (
	AggregationUnknown = 0 // Unknown aggregation method
	AggregationAverage = 1 // Aggregate using averaging
	AggregationSum     = 2 // Aggregate using sum
	AggregationLast    = 3 // Aggregate using the last value
	AggregationMax     = 4 // Aggregate using the maximum value
	AggregationMin     = 5 // Aggregate using the minimum value
)

var (
	AggregationMethodsNtoV = map[string]AggregationMethod{
		"average": AggregationAverage,
		"sum":     AggregationSum,
		"last":    AggregationLast,
		"max":     AggregationMax,
		"min":     AggregationMin,
	}
	AggregationMethodsVtoN = reverseMethods(AggregationMethodsNtoV)
)

func reverseMethods(methods map[string]AggregationMethod) map[AggregationMethod]string {
	result := make(map[AggregationMethod]string)
	for k, v := range methods {
		result[v] = k
	}
	return result
}

func (m AggregationMethod) String() string {
	v, ok := AggregationMethodsVtoN[m]
	if !ok {
		return "unknown"
	}
	return v
}

func AggregationMethodByName(name string) AggregationMethod {
	v, ok := AggregationMethodsNtoV[name]
	if !ok {
		return AggregationUnknown
	}
	return v
}
