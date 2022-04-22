package sizing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestGetLimitSizeForRequestSize(t *testing.T) {
	assert.Equal(t, getLimitSizeForRequestSize("small"), "medium")
	assert.Equal(t, getLimitSizeForRequestSize("medium"), "large")
	assert.Equal(t, getLimitSizeForRequestSize("large"), "x-large")
}

func TestGetVolSizeToCapacityMap(t *testing.T) {
	s := getVolSizeToCapacityMap()
	assert.Equal(t, s["x-small"], "1Gi")
	assert.Equal(t, s["small"], "2Gi")
	assert.Equal(t, s["medium"], "3Gi")
	assert.Equal(t, s["large"], "5Gi")
}

func TestGetCPUSizeToCapacityMap(t *testing.T) {
	c := getCPUSizeToCapacityMap()
	assert.Equal(t, c["small"], "600m")
	assert.Equal(t, c["medium"], "1200m")
	assert.Equal(t, c["large"], "1800m")
	assert.Equal(t, c["x-large"], "2400m")
}

func TestGetRAMSizeToCapacityMap(t *testing.T) {
	r := getRAMSizeToCapacityMap()
	assert.Equal(t, r["small"], "512Mi")
	assert.Equal(t, r["medium"], "1Gi")
	assert.Equal(t, r["large"], "2Gi")
	assert.Equal(t, r["x-large"], "3Gi")
}

func TestGetDefaultResourceRequirements(t *testing.T) {
	reqs := GetDefaultResourceRequirements()

	ramSmall := getRAMSizeToCapacityMap()["small"]
	cpuSmall := getCPUSizeToCapacityMap()["small"]
	ramMed := getRAMSizeToCapacityMap()["medium"]
	cpuMed := getCPUSizeToCapacityMap()["medium"]

	assert.Equal(t, reqs.Limits["memory"], resource.MustParse(ramMed))
	assert.Equal(t, reqs.Limits["cpu"], resource.MustParse(cpuMed))
	assert.Equal(t, reqs.Requests["memory"], resource.MustParse(ramSmall))
	assert.Equal(t, reqs.Requests["cpu"], resource.MustParse(cpuSmall))
}

func TestDBDResourceRequirements(t *testing.T) {
	reqs := GetResourceRequirementsForSize("medium")

	ramLarge := getRAMSizeToCapacityMap()["large"]
	cpuLarge := getCPUSizeToCapacityMap()["large"]
	ramMed := getRAMSizeToCapacityMap()["medium"]
	cpuMed := getCPUSizeToCapacityMap()["medium"]

	assert.Equal(t, reqs.Limits["memory"], resource.MustParse(ramLarge))
	assert.Equal(t, reqs.Limits["cpu"], resource.MustParse(cpuLarge))
	assert.Equal(t, reqs.Requests["memory"], resource.MustParse(ramMed))
	assert.Equal(t, reqs.Requests["cpu"], resource.MustParse(cpuMed))
}

func TestGetDefaultVolCapacity(t *testing.T) {
	d := GetDefaultVolCapacity()
	dd := getVolSizeToCapacityMap()[GetDefaultSizeVol()]
	assert.Equal(t, d, dd)
}
