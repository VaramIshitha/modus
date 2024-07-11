package utils

import (
	"errors"
	"fmt"
	"math"

	"github.com/chewxy/math32"
)

const (
	Euclidian            = "euclidian"
	Cosine               = "cosine"
	DotProd              = "dotproduct"
	plError              = "\nerror fetching posting list for data key: "
	dataError            = "\nerror fetching data for data key: "
	VecKeyword           = "__vector_"
	visitedVectorsLevel  = "visited_vectors_level_"
	distanceComputations = "vector_distance_computations"
	searchTime           = "vector_search_time"
	VecEntry             = "__vector_entry"
	VecDead              = "__vector_dead"
	VectorIndexMaxLevels = 5
	EfConstruction       = 16
	EfSearch             = 12
	numEdgesConst        = 2
	// ByteData indicates the key stores data.
	ByteData = byte(0x00)
	// DefaultPrefix is the prefix used for data, index and reverse keys so that relative
	DefaultPrefix = byte(0x00)
	// NsSeparator is the separator between the namespace and attribute.
	NsSeparator = "-"
)

func IsBetterScoreForSimilarity(a, b float64) bool {
	return a > b
}

func Normalize(v []float32) ([]float32, error) {
	norm := norm(v)
	if norm == 0 {
		return nil, errors.New("can not normalize vector with zero norm")
	}
	for i := range v {
		v[i] /= norm
	}
	return v, nil
}

func norm(v []float32) float32 {
	vectorNorm, _ := DotProduct(v, v)
	return math32.Sqrt(vectorNorm)
}

func DotProduct(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, errors.New("can not compute dot product on vectors of different lengths")
	}
	var dotProduct float32
	for i := range a {
		dotProduct += a[i] * b[i]
	}
	return dotProduct, nil
}

// assume normalization for vectors
func CosineSimilarity(a, b []float32) (float64, error) {
	dotProd, err := DotProduct(a, b)
	if err != nil {
		return 0, err
	}

	return float64(dotProd), nil
}

func ConcatStrings(strs ...string) string {
	total := ""
	for _, s := range strs {
		total += s
	}
	return total
}

func ConvertToFloat32_2DArray(result any) ([][]float32, error) {
	resultArr, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("error converting type to float32: %v", result)
	}

	textVecs := make([][]float32, len(resultArr))
	for i, res := range resultArr {

		subArr, ok := res.([]interface{})
		if !ok {
			return nil, fmt.Errorf("error converting type to float32: %v", res)
		}

		textVecs[i] = make([]float32, len(subArr))
		for j, val := range subArr {
			if v, ok := val.(float64); ok {
				textVecs[i][j] = float32(v)
			} else if v, ok := val.(float32); ok {
				textVecs[i][j] = v
			} else {
				return nil, fmt.Errorf("error converting type to float32: %v", val)
			}
		}
	}
	return textVecs, nil
}

func EqualFloat32Slices(a, b []float32) bool {
	const epsilon = 1e-9
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.Abs(float64(a[i]-b[i])) > epsilon {
			return false
		}
	}
	return true
}