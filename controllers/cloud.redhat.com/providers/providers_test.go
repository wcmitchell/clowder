package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestGetLimitForRequestSize(t *testing.T) {
	assert.Equal(t, GetLimitForRequestSize("small"), "medium")
	assert.Equal(t, GetLimitForRequestSize("medium"), "large")
	assert.Equal(t, GetLimitForRequestSize("large"), "x-large")
}

func TestGetDBVolSizes(t *testing.T) {
	s := GetDBVolSizes()
	assert.Equal(t, s[""], "1Gi")
	assert.Equal(t, s["small"], "2Gi")
	assert.Equal(t, s["medium"], "3Gi")
	assert.Equal(t, s["large"], "5Gi")
}

func TestGetDBCPUSizes(t *testing.T) {
	c := GetDBCPUSizes()
	assert.Equal(t, c["small"], "600M")
	assert.Equal(t, c["medium"], "1200M")
	assert.Equal(t, c["large"], "2400M")
	assert.Equal(t, c["x-large"], "3200M")
}

func TestGetDBRAMSizes(t *testing.T) {
	r := GetDBRAMSizes()
	assert.Equal(t, r["small"], "1024Mi")
	assert.Equal(t, r["medium"], "2048Mi")
	assert.Equal(t, r["large"], "4096Mi")
	assert.Equal(t, r["x-large"], "6144Mi")
}

func TestGetDBDefaultResourceRequirements(t *testing.T) {
	reqs := GetDBDefaultResourceRequirements()

	ramSmall := GetDBRAMSizes()["small"]
	cpuSmall := GetDBCPUSizes()["small"]
	ramMed := GetDBRAMSizes()["medium"]
	cpuMed := GetDBCPUSizes()["medium"]

	assert.Equal(t, reqs.Limits["memory"], resource.MustParse(ramMed))
	assert.Equal(t, reqs.Limits["cpu"], resource.MustParse(cpuMed))
	assert.Equal(t, reqs.Requests["memory"], resource.MustParse(ramSmall))
	assert.Equal(t, reqs.Requests["cpu"], resource.MustParse(cpuSmall))
}

func TestDBDResourceRequirements(t *testing.T) {
	reqs := GetDBResourceRequirements("medium")

	ramLarge := GetDBRAMSizes()["large"]
	cpuLarge := GetDBCPUSizes()["large"]
	ramMed := GetDBRAMSizes()["medium"]
	cpuMed := GetDBCPUSizes()["medium"]

	assert.Equal(t, reqs.Limits["memory"], resource.MustParse(ramLarge))
	assert.Equal(t, reqs.Limits["cpu"], resource.MustParse(cpuLarge))
	assert.Equal(t, reqs.Requests["memory"], resource.MustParse(ramMed))
	assert.Equal(t, reqs.Requests["cpu"], resource.MustParse(cpuMed))
}

func TestGetDBDefaultVolSize(t *testing.T) {
	d := GetDBDefaultVolSize()
	dd := GetDBVolSizes()[""]
	assert.Equal(t, d, dd)
}
