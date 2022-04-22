package sizing

import (
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	DEFAULT_SIZE_VOL     string = "x-small"
	DEFAULT_SIZE_CPU_RAM string = "small"
)

func GetDefaultVolSize() string {
	return DEFAULT_SIZE_VOL
}

func GetDefaultSizeCPURAM() string {
	return DEFAULT_SIZE_CPU_RAM
}

//Naming is hard. In the context of this API "sizes" are
//the t shirt sizes (small, medium, etc) whereas "capacities"
//are the values k8s uses like Gi, M, m, etc. This distinction is
//important because most of what we're doing here is
//converting sizes to capacities, so I enforce the distinction
//strictly. The method names are long, but more importantly, accurate

func IsCapacityLarger(capacityA string, capacityB string) bool {
	capacities := map[string]int{
		"x-small": 0,
		"small":   1,
		"medium":  2,
		"large":   3,
		"x-large": 4,
	}
	return capacities[capacityA] > capacities[capacityB]
}

//For any given size get the next size up
//Allows for size to limit mapping without conditionality
func GetLimitSizeForRequestSize(tShirtSize string) string {
	sizeMap := map[string]string{
		"x-small": "small",
		"small":   "medium",
		"medium":  "large",
		"large":   "x-large",
	}
	return sizeMap[tShirtSize]
}

//Get a map of volume T-Shirt sizes
func GetVolSizeToCapacityMap() map[string]string {
	return map[string]string{
		//x-small is because volume t shirt sizes pre-exist this implementation and there
		//we shipped a default smaller than small. I'm just leaving that pattern intact
		"x-small": "1Gi",
		"small":   "2Gi",
		"medium":  "3Gi",
		"large":   "5Gi",
	}
}

//Get a map of CPU T-Shirt sizes
func GetCPUSizeToCapacityMap() map[string]string {
	return map[string]string{
		"small":  "600m",
		"medium": "1200m",
		"large":  "1800m",
		//Why x-large? For CPU and RAM we have a request and a limit. The limit needs to be
		//larger than the request. Therefore, if large is requested we need an x-large as a
		//limit. x-large can't be requested - it isn't part of the config enum valid value set
		"x-large": "2400m",
	}
}

//Get a map of RAM T-Shirt sizes
func GetRAMSizeToCapacityMap() map[string]string {
	return map[string]string{
		"small":   "512Mi",
		"medium":  "1Gi",
		"large":   "2Gi",
		"x-large": "3Gi",
	}
}

//For a givin vol size get the capacity
func GetVolCapacityForSize(size string) string {
	requestSize := size
	//Oh golang... my kingdom for a ternary operator
	if requestSize == "" {
		requestSize = DEFAULT_SIZE_VOL
	}
	return GetVolSizeToCapacityMap()[requestSize]
}

//Get the default volume size, for use if none is provided
func GetDefaultVolCapacity() string {
	return GetVolSizeToCapacityMap()[DEFAULT_SIZE_VOL]
}

//Get the default database resource requirements
func GetResourceRequirementsForSize(tShirtSize string) core.ResourceRequirements {
	cpu := GetCPUSizeToCapacityMap()
	ram := GetRAMSizeToCapacityMap()
	limitSize := GetLimitSizeForRequestSize(tShirtSize)
	return core.ResourceRequirements{
		Limits: core.ResourceList{
			"memory": resource.MustParse(ram[limitSize]),
			"cpu":    resource.MustParse(cpu[limitSize]),
		},
		Requests: core.ResourceList{
			"memory": resource.MustParse(ram[tShirtSize]),
			"cpu":    resource.MustParse(cpu[tShirtSize]),
		},
	}
}

func GetDefaultResourceRequirements() core.ResourceRequirements {
	return GetResourceRequirementsForSize(DEFAULT_SIZE_CPU_RAM)
}
