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
	SanitizeResourceBlockLabel(inputLabel string) string
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
		log.Print("Using the time optimized resource label sanitizer with transliteration")
		return &SanitizerProvider{
			S: &sanitizerOptimized{},
		}
	}

	log.Print("Using the original resource label sanitizer")
	return &SanitizerProvider{
		S: &sanitizerOriginal{},
	}
}

// Sanitize sanitizes all resource labels using the optimized algorithm
func (sod *sanitizerOriginal) Sanitize(idMetaMap ResourceIDMetaMap) {
	// Pull out all the original labels of the resources for reference later
	originalResourceLabels := make(map[string]string)
	for k, v := range idMetaMap {
		originalResourceLabels[k] = v.BlockLabel
	}

	// Iterate over the idMetaMap and sanitize the labels of each resource
	for _, meta := range idMetaMap {

		sanitizedLabel := sod.SanitizeResourceBlockLabel(meta.BlockLabel)

		// If there are more than one resource label that ends up with the same sanitized label,
		// append a hash of the original label to ensure uniqueness for labels to prevent duplicates
		if sanitizedLabel != meta.BlockLabel {
			numSeen := 0
			for _, originalLabel := range originalResourceLabels {
				originalSanitizedLabel := sod.SanitizeResourceBlockLabel(originalLabel)
				if sanitizedLabel == originalSanitizedLabel {
					numSeen++
				}
			}
			if numSeen > 1 {
				algorithm := fnv.New32()
				algorithm.Write([]byte(meta.BlockLabel))
				sanitizedLabel = sanitizedLabel + "_" + strconv.FormatUint(uint64(algorithm.Sum32()), 10)
			}
			meta.BlockLabel = sanitizedLabel
		}
	}
}

// SanitizeResourceBlockLabel sanitizes a single resource label
func (sod *sanitizerOriginal) SanitizeResourceBlockLabel(inputLabel string) string {
	label := unsafeLabelChars.ReplaceAllStringFunc(inputLabel, escapeRune)

	if unsafeLabelStartingChars.MatchString(string(rune(label[0]))) {
		// Terraform does not allow labels to begin with a number. Prefix with an underscore instead
		label = "_" + label
	}

	return label
}

// Sanitize sanitizes all resource label using the time optimized algorithm
func (sod *sanitizerOptimized) Sanitize(idMetaMap ResourceIDMetaMap) {
	sanitizedLabels := make(map[string]int, len(idMetaMap))

	for _, meta := range idMetaMap {
		sanitizedLabel := sod.SanitizeResourceBlockLabel(meta.BlockLabel)

		if sanitizedLabel != meta.BlockLabel {
			if count, exists := sanitizedLabels[sanitizedLabel]; exists {
				// We've seen this sanitized label before
				sanitizedLabels[sanitizedLabel] = count + 1

				// Append a hash to ensure uniqueness
				h := sha256.New()
				h.Write([]byte(meta.BlockLabel))
				hash := hex.EncodeToString(h.Sum(nil)[:10]) // Use first 10 characters of hash

				meta.BlockLabel = sanitizedLabel + "_" + hash
			} else {
				sanitizedLabels[sanitizedLabel] = 1
				meta.BlockLabel = sanitizedLabel
			}
		}
	}
}

// SanitizeResourceBlockLabel sanitizes a single resource label
func (sod *sanitizerOptimized) SanitizeResourceBlockLabel(inputLabel string) string {
	// Transliterate any non-latin-based characters to ASCII
	transliteratedLabel := strings.TrimSpace(unidecode.Unidecode(inputLabel))
	label := unsafeLabelChars.ReplaceAllStringFunc(transliteratedLabel, escapeRune)

	if unsafeLabelStartingChars.MatchString(string(rune(label[0]))) {
		// Terraform does not allow labels to begin with a number. Prefix with an underscore instead
		label = "_" + label
	}

	return label
}
