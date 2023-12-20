package resource_exporter

import (
	"hash/fnv"
	"log"
	"os"
	"strconv"
)

type SanitizerProvider struct {
	S Sanitizer
}

type Sanitizer interface {
	Sanitize(idMetaMap ResourceIDMetaMap)
	SanitizeResourceName(inputName string) string
}

// Two different Sanitizer structs one with the original algorithmn
type sanitizerOriginal struct{}
type sanitizerOptimized struct{}

// NewSanitizierProvider returns a Sanitizer. Without a GENESYS_SANITIZER_LEGACY environment variable set it will always use the optimized Sanitizer
func NewSanitizerProvider() *SanitizerProvider {
	// Check if the environment variable is set
	_, exists := os.LookupEnv("GENESYS_SANITIZER_LEGACY")

	//If the GENESYS_SANITIZER_LEGACY is set use the original name sanitizer
	if exists {
		log.Print("Using the original resource name sanitizer")
		return &SanitizerProvider{
			S: &sanitizerOriginal{},
		}
	}

	log.Print("Using the optimized resource name sanitizer")
	return &SanitizerProvider{
		S: &sanitizerOptimized{},
	}

}

// Sanitize sanitizes all the resource names using the original algorithm
func (so *sanitizerOriginal) Sanitize(idMetaMap ResourceIDMetaMap) {
	for _, meta := range idMetaMap {
		meta.Name = so.SanitizeResourceName(meta.Name)
	}
}

// SanitizeResourceName sanitizes a single resource name using  the original resource name sanitizer
func (so *sanitizerOriginal) SanitizeResourceName(inputName string) string {
	name := unsafeNameChars.ReplaceAllStringFunc(inputName, escapeRune)
	if name != inputName {
		// Append a hash of the original name to ensure uniqueness for similar names
		// and that equivalent names are consistent across orgs
		algorithm := fnv.New32()
		algorithm.Write([]byte(inputName))
		name = name + "_" + strconv.FormatUint(uint64(algorithm.Sum32()), 10)
	}
	if unsafeNameStartingChars.MatchString(string(rune(name[0]))) {
		// Terraform does not allow names to begin with a number. Prefix with an underscore instead
		name = "_" + name
	}

	return name
}

// Sanitize sanitizes all resource name using the optimized algorithm
func (sod *sanitizerOptimized) Sanitize(idMetaMap ResourceIDMetaMap) {
	// Pull out all the original names of the resources for reference later
	originalResourceNames := make(map[string]string)
	for k, v := range idMetaMap {
		originalResourceNames[k] = v.Name
	}

	// Iterate over the idMetaMap and sanitize the names of each resource
	for _, meta := range idMetaMap {

		sanitizedName := sod.SanitizeResourceName(meta.Name)

		// If there are more than one resource name that ends up with the same sanitized name,
		// append a hash of the original name to ensure uniqueness for names to prevent duplicates
		if sanitizedName != meta.Name {
			numSeen := 0
			for _, originalName := range originalResourceNames {
				originalSanitizedName := sod.SanitizeResourceName(originalName)
				if sanitizedName == originalSanitizedName {
					numSeen++
				}
			}
			if numSeen > 1 {
				algorithm := fnv.New32()
				algorithm.Write([]byte(meta.Name))
				sanitizedName = sanitizedName + "_" + strconv.FormatUint(uint64(algorithm.Sum32()), 10)
			}
			meta.Name = sanitizedName
		}
	}
}

// SanitizeResourceName sanitizes a single resource name
func (sod *sanitizerOptimized) SanitizeResourceName(inputName string) string {
	name := unsafeNameChars.ReplaceAllStringFunc(inputName, escapeRune)

	if unsafeNameStartingChars.MatchString(string(rune(name[0]))) {
		// Terraform does not allow names to begin with a number. Prefix with an underscore instead
		name = "_" + name
	}

	return name
}
