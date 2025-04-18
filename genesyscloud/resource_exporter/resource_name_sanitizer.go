package resource_exporter

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"log"
	"strconv"
	"strings"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"

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
type sanitizerBCPOptimized struct{}

// NewSanitizerProvider returns a Sanitizer.
func NewSanitizerProvider() *SanitizerProvider {

	// Check if the Optimized Sanitizer environment variable is set
	optimizedExists := featureToggles.ExporterSanitizerOptimizedToggleExists()
	if optimizedExists {
		log.Print("Using the time optimized resource label sanitizer with transliteration")
		return &SanitizerProvider{
			S: &sanitizerOptimized{},
		}
	}

	// Check if the BCP Optimized Sanitizer environment variable is set
	bcpOptimizedExists := featureToggles.ExporterSanitizerBCPOptimizedToggleExists()
	if bcpOptimizedExists {
		log.Print("Using the BCP optimized resource label sanitizer")
		return &SanitizerProvider{
			S: &sanitizerBCPOptimized{},
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
				sanitizedLabel = sanitizedLabel + "_" + sod.SanitizeResourceHash(meta.BlockLabel)
			}
			if meta.OriginalLabel == "" {
				meta.OriginalLabel = meta.BlockLabel
			}
			meta.BlockLabel = sanitizedLabel
		}
	}
}

// SanitizeResourceBlockLabel sanitizes a single resource label
func (sod *sanitizerOriginal) SanitizeResourceBlockLabel(inputLabel string) string {
	if inputLabel == "" {
		return ""
	}
	label := unsafeLabelChars.ReplaceAllStringFunc(inputLabel, escapeRune)
	if len(label) == 0 {
		return ""
	}
	if unsafeLabelStartingChars.MatchString(string(rune(label[0]))) {
		// Terraform does not allow labels to begin with a number. Prefix with an underscore instead
		label = "_" + label
	}

	return label
}

func (sod *sanitizerOriginal) SanitizeResourceHash(originalBlockLabel string) string {
	algorithm := fnv.New32()
	algorithm.Write([]byte(originalBlockLabel))
	return strconv.FormatUint(uint64(algorithm.Sum32()), 10)
}

// Sanitize sanitizes all resource label using the time optimized algorithm
func (sod *sanitizerOptimized) Sanitize(idMetaMap ResourceIDMetaMap) {
	sanitizedLabels := make(map[string]int, len(idMetaMap))

	for _, meta := range idMetaMap {
		sanitizedLabel := sod.SanitizeResourceBlockLabel(meta.BlockLabel)

		if meta.OriginalLabel == "" {
			meta.OriginalLabel = meta.BlockLabel
		}

		if sanitizedLabel != meta.BlockLabel {
			if count, exists := sanitizedLabels[sanitizedLabel]; exists {
				// We've seen this sanitized label before
				sanitizedLabels[sanitizedLabel] = count + 1

				// Append a hash to ensure uniqueness
				hash := sod.SanitizeResourceHash(meta.BlockLabel)
				meta.BlockLabel = sanitizedLabel + "_" + hash
				if meta.OriginalLabel == "" {
					meta.OriginalLabel = meta.BlockLabel
				}

			} else {
				sanitizedLabels[sanitizedLabel] = 1
				meta.BlockLabel = sanitizedLabel
			}
		}
	}
}

// SanitizeResourceBlockLabel sanitizes a single resource label
func (sod *sanitizerOptimized) SanitizeResourceBlockLabel(inputLabel string) string {
	if inputLabel == "" {
		return ""
	}
	// Transliterate any non-latin-based characters to ASCII
	transliteratedLabel := strings.TrimSpace(unidecode.Unidecode(inputLabel))
	label := unsafeLabelChars.ReplaceAllStringFunc(transliteratedLabel, escapeRune)

	if len(label) == 0 {
		return ""
	}
	if unsafeLabelStartingChars.MatchString(string(rune(label[0]))) {
		// Terraform does not allow labels to begin with a number. Prefix with an underscore instead
		label = "_" + label
	}

	return label
}

func (sod *sanitizerOptimized) SanitizeResourceHash(originalBlockLabel string) string {
	h := sha256.New()
	h.Write([]byte(originalBlockLabel))
	return hex.EncodeToString(h.Sum(nil)[:10]) // Use first 10 characters of hash
}

// Sanitize sanitizes all resource label using the BCP specific algorithm which includes
// adding hashes to the end of all resource block labels to ensure consistent uniqueness
// See DEVTOOLING-1182 for details on why this sanitizer was necessary
func (sod *sanitizerBCPOptimized) Sanitize(idMetaMap ResourceIDMetaMap) {
	sanitizedLabels := make(map[string][]string)

	// First pass to generate sanitized labels and group them
	for id, meta := range idMetaMap {
		meta.OriginalLabel = meta.BlockLabel
		sanitizedLabel := sod.SanitizeResourceBlockLabel(meta.BlockLabel)
		hash := sod.SanitizeResourceHash(*meta)
		sanitizedLabel = sanitizedLabel + "__" + hash

		sanitizedLabels[sanitizedLabel] = append(sanitizedLabels[sanitizedLabel], id)
	}

	// Second pass to update labels
	for id, meta := range idMetaMap {
		sanitizedLabel := sod.SanitizeResourceBlockLabel(meta.BlockLabel)
		hash := sod.SanitizeResourceHash(*meta)
		sanitizedLabel = sanitizedLabel + "__" + hash

		// This should never/rarely happen as the sanitizer is specifically designed to prevent duplicates
		// through the BLH (Block Label Hash) and UFH (Unique Fields Hash) strategies.
		// If you encounter this, it most likely means the resource type needs a BlockHash (UFH) configured.
		// To resolve:
		// 1. Check if the resource type has a BlockHash configured in its exporter
		// 2. If not, add a BlockHash using unique identifying fields for the resource type. Details on this
		//    be found in the ResourceMeta.BlockHash documentation with directions to use util.QuickHashFields().
		// 3. If BlockHash is already configured, file a bug report with:
		//    - The original resource labels that caused the collision
		//    - The complete resource configurations
		//    - The generated hashes (both BLH and UFH)
		// Reference DEVTOOLING-1183 for more details on BlockHash implementation
		if len(sanitizedLabels[sanitizedLabel]) > 1 {

			instanceNum := 0
			for i, storedId := range sanitizedLabels[sanitizedLabel] {
				if storedId == id {
					instanceNum = i + 1
					break
				}
			}
			sanitizedLabel = fmt.Sprintf("%s_DUPLICATE_INSTANCE_PLEASE_REPORT_%d", sanitizedLabel, instanceNum)
		}

		meta.BlockLabel = sanitizedLabel
	}
}

// SanitizeResourceBlockLabel sanitizes a single resource label
func (sod *sanitizerBCPOptimized) SanitizeResourceBlockLabel(inputLabel string) string {
	if inputLabel == "" {
		return ""
	}
	// Transliterate any non-latin-based characters to ASCII
	transliteratedLabel := strings.TrimSpace(unidecode.Unidecode(inputLabel))
	label := unsafeLabelChars.ReplaceAllStringFunc(transliteratedLabel, escapeRune)
	if len(label) == 0 {
		return ""
	}
	if unsafeLabelStartingChars.MatchString(string(rune(label[0]))) {
		// Terraform does not allow labels to begin with a number. Prefix with an underscore instead
		label = "_" + label
	}

	return label
}

func (sod *sanitizerBCPOptimized) SanitizeResourceHash(meta ResourceMeta) string {
	h := sha256.New()
	h.Write([]byte(meta.OriginalLabel))
	labelHash := hex.EncodeToString(h.Sum(nil)[:10]) // Use first 10 characters of hash

	// BLH prefix = "block label hash"
	labelHashAppendage := "_BLH" + labelHash
	if meta.BlockHash != "" {
		// UFH prefix = "unique fields hash"
		labelHashAppendage = labelHashAppendage + "_UFH" + meta.BlockHash
	}
	return labelHashAppendage
}
