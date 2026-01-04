package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDistanceKm(t *testing.T) {
	d := DistanceKm(12.97, 77.59, 12.97, 77.59)
	assert.InDelta(t, 0, d, 0.01)
}

func TestGeoCell(t *testing.T) {
	cell := GeoCell(12.9876, 77.5946)
	assert.Equal(t, "13.0:77.6", cell)
}
