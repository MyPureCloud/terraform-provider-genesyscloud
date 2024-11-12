package resource_exporter

import (
	"crypto/sha256"
	"encoding/hex"
	"hash/fnv"
	"log"
	"strconv"
	"strings"
	feature_toggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"

	unidecode "github.com/mozillazg/go-unidecode"
)

type SanitizerProvider struct {
	S Sanitizer
}

type Sanitizer interface {
	Sanitize(idMetaMap ResourceIDMetaMap)
	SanitizeResourceName(inputName string) string
}

// Two different Sanitizer structs one with the original algorithm
type sanitizerOriginal struct{}
type sanitizerOptimized struct{}

// NewSanitizerProvider returns a Sanitizer. Without a GENESYS_SANITIZER_LEGACY environment variable set it will always use the optimized Sanitizer
func NewSanitizerProvider() *SanitizerProvider {

	// Check if the environment variable is set
	optimizedExists := feature_toggles.ExporterSanitizerOptimizedToggleExists()

	//If the GENESYS_SANITIZER_TIME_OPTIMIZED is set use the updated time optimized sanitizer
	if optimizedExists {
		log.Print("Using the time optimized resource name sanitizer with transliteration")
		return &SanitizerProvider{
			S: &sanitizerOptimized{},
		}
	}

	log.Print("Using the original resource name sanitizer")
	return &SanitizerProvider{
		S: &sanitizerOriginal{},
	}
}

// Sanitize sanitizes all resource name using the optimized algorithm
func (sod *sanitizerOriginal) Sanitize(idMetaMap ResourceIDMetaMap) {
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
func (sod *sanitizerOriginal) SanitizeResourceName(inputName string) string {
	name := unsafeNameChars.ReplaceAllStringFunc(inputName, escapeRune)

	if unsafeNameStartingChars.MatchString(string(rune(name[0]))) {
		// Terraform does not allow names to begin with a number. Prefix with an underscore instead
		name = "_" + name
	}

	return name
}

// Sanitize sanitizes all resource name using the time optimized algorithm
func (sod *sanitizerOptimized) Sanitize(idMetaMap ResourceIDMetaMap) {
	sanitizedNames := make(map[string]int, len(idMetaMap))

	for _, meta := range idMetaMap {
		sanitizedName := sod.SanitizeResourceName(meta.Name)

		if sanitizedName != meta.Name {
			if count, exists := sanitizedNames[sanitizedName]; exists {
				// We've seen this sanitized name before
				sanitizedNames[sanitizedName] = count + 1

				// Append a hash to ensure uniqueness
				h := sha256.New()
				h.Write([]byte(meta.Name))
				hash := hex.EncodeToString(h.Sum(nil)[:10]) // Use first 10 characters of hash

				meta.Name = sanitizedName + "_" + hash
			} else {
				sanitizedNames[sanitizedName] = 1
				meta.Name = sanitizedName
			}
		}
	}
}

// SanitizeResourceName sanitizes a single resource name
func (sod *sanitizerOptimized) SanitizeResourceName(inputName string) string {
	// Transliterate any non-latin-based characters to ASCII
	transliteratedName := strings.TrimSpace(unidecode.Unidecode(inputName))
	name := unsafeNameChars.ReplaceAllStringFunc(transliteratedName, escapeRune)

	if unsafeNameStartingChars.MatchString(string(rune(name[0]))) {
		// Terraform does not allow names to begin with a number. Prefix with an underscore instead
		name = "_" + name
	}

	return name
}
