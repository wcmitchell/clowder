package sizing

//Note on package:
//I didn't really want to pull this out into its own package
//I wanted this in database or providers but I ran into dependency cycle problems
//no matter what I did. So easiest and cleanest solution was just to pull it out
//that said maybe in the future we can extend sizing out to other stuff in which case
//a sizing package will be helpful

//Note on naming:
//Naming is hard. In the context of this API "sizes" are
//the t shirt sizes (small, medium, etc) whereas "capacities"
//are the values k8s uses like Gi, M, m, etc. This distinction is
//important because most of what we're doing here is
//converting sizes to capacities, so I enforce the distinction
//strictly. The method names are long, but more importantly, accurate

import (
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	//The nature of the beast in this code is we got a lot of magic strings
	XSMALL string = "x-small"
	SMALL  string = "small"
	MEDIUM string = "medium"
	LARGE  string = "large"
	XLARGE string = "x-large"
	//We need to define default sizes because if a ClowdApp doesn't provide
	//volume or ram/cpu capacities we just get an empty string, so we need
	//defaults to plug in there
	DEFAULT_SIZE_VOL     string = XSMALL
	DEFAULT_SIZE_CPU_RAM string = SMALL
)

// Public methods

func GetDefaultResourceRequirements() core.ResourceRequirements {
	return GetResourceRequirementsForSize(GetDefaultSizeCPURAM())
}

//Gets the default size for CPU and RAM
func GetDefaultSizeCPURAM() string {
	return DEFAULT_SIZE_CPU_RAM
}

//Gets the default vol size
func GetDefaultSizeVol() string {
	return DEFAULT_SIZE_VOL
}

//Get the default volume size, for use if none is provided
func GetDefaultVolCapacity() string {
	return getVolSizeToCapacityMap()[GetDefaultSizeVol()]
}

//Get the default database resource requirements
func GetResourceRequirementsForSize(tShirtSize string) core.ResourceRequirements {
	requestSize := useDefaultIfEmptySize(tShirtSize, GetDefaultSizeCPURAM())
	cpu := getCPUSizeToCapacityMap()
	ram := getRAMSizeToCapacityMap()
	limitSize := getLimitSizeForRequestSize(requestSize)
	return core.ResourceRequirements{
		Limits: core.ResourceList{
			"memory": resource.MustParse(ram[limitSize]),
			"cpu":    resource.MustParse(cpu[limitSize]),
		},
		Requests: core.ResourceList{
			"memory": resource.MustParse(ram[requestSize]),
			"cpu":    resource.MustParse(cpu[requestSize]),
		},
	}
}

//For a givin vol size get the capacity
//If "" is provided you'll get DEFAULT_SIZE_VOL
func GetVolCapacityForSize(size string) string {
	requestSize := useDefaultIfEmptySize(size, GetDefaultSizeVol())
	return getVolSizeToCapacityMap()[requestSize]
}

//Sometimes we need to know if one size is larger than another
func IsCapacityLarger(capacityA string, capacityB string) bool {
	capacities := map[string]int{
		XSMALL: 0,
		SMALL:  1,
		MEDIUM: 2,
		LARGE:  3,
		XLARGE: 4,
	}
	return capacities[capacityA] > capacities[capacityB]
}

// Private methods

//Get a map of CPU T-Shirt sizes to capacities
func getCPUSizeToCapacityMap() map[string]string {
	return map[string]string{
		SMALL:  "600m",
		MEDIUM: "1200m",
		LARGE:  "1800m",
		//Why x-large? For CPU and RAM we have a request and a limit. The limit needs to be
		//larger than the request. Therefore, if large is requested we need an x-large as a
		//limit. x-large can't be requested - it isn't part of the config enum valid value set
		XLARGE: "2400m",
	}
}

//For any given size get the next size up
//Allows for size to limit mapping without conditionality
func getLimitSizeForRequestSize(tShirtSize string) string {
	sizeMap := map[string]string{
		XSMALL: SMALL,
		SMALL:  MEDIUM,
		MEDIUM: LARGE,
		LARGE:  XLARGE,
	}
	return sizeMap[tShirtSize]
}

//Get a map of RAM T-Shirt sizes to capacities
func getRAMSizeToCapacityMap() map[string]string {
	return map[string]string{
		SMALL:  "512Mi",
		MEDIUM: "1Gi",
		LARGE:  "2Gi",
		XLARGE: "3Gi",
	}
}

//Get a map of volume T-Shirt size to capacities
func getVolSizeToCapacityMap() map[string]string {
	return map[string]string{
		//x-small is because volume t shirt sizes pre-exist this implementation and there
		//we shipped a default smaller than small. I'm just leaving that pattern intact
		//In real life no one requests x-small, they request "" and get x-small
		XSMALL: "1Gi",
		SMALL:  "2Gi",
		MEDIUM: "3Gi",
		LARGE:  "5Gi",
	}
}

//Often we have to sanitize a size such that "" == whatever the default is
func useDefaultIfEmptySize(size string, def string) string {
	if size == "" {
		return def
	}
	return size
}
