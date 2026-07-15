package enconvert

import (
	"fmt"
	"sort"
	"strings"
)

// Format tables mirroring the gateway's CONVERTER_MAP (api/v1/convert.py).
//
// ImplementedConversions is the client-side gate: the gateway returns 503
// for any "{input}-to-{output}" endpoint not in its CONVERTER_MAP, so we
// reject unsupported pairs here with a useful message instead of paying a
// network round-trip for a guaranteed failure.

// ImplementedConversions lists the 44 implemented "{input}-to-{output}"
// conversion endpoints, copied verbatim from the Node SDK's formats.ts.
var ImplementedConversions = map[string]struct{}{
	// Structured text (13)
	"json-to-xml":      {},
	"xml-to-json":      {},
	"json-to-yaml":     {},
	"yaml-to-json":     {},
	"csv-to-json":      {},
	"json-to-csv":      {},
	"json-to-toml":     {},
	"toml-to-json":     {},
	"csv-to-xml":       {},
	"xml-to-csv":       {},
	"markdown-to-html": {},
	"markdown-to-pdf":  {},
	"html-to-pdf":      {},
	// Documents (10)
	"doc-to-pdf":     {},
	"excel-to-pdf":   {},
	"ppt-to-pdf":     {},
	"odt-to-pdf":     {},
	"ods-to-pdf":     {},
	"odp-to-pdf":     {},
	"ots-to-pdf":     {},
	"pages-to-pdf":   {},
	"numbers-to-pdf": {},
	"epub-to-pdf":    {},
	// Images (21)
	"jpeg-to-png":  {},
	"png-to-jpeg":  {},
	"jpeg-to-svg":  {},
	"svg-to-jpeg":  {},
	"jpeg-to-heic": {},
	"heic-to-jpeg": {},
	"jpeg-to-webp": {},
	"webp-to-jpeg": {},
	"png-to-svg":   {},
	"svg-to-png":   {},
	"png-to-heic":  {},
	"heic-to-png":  {},
	"png-to-webp":  {},
	"webp-to-png":  {},
	"svg-to-heic":  {},
	"heic-to-svg":  {},
	"svg-to-webp":  {},
	"webp-to-svg":  {},
	"heic-to-webp": {},
	"webp-to-heic": {},
	"pdf-to-jpeg":  {},
}

// imageFormats maps a filename extension to its API input format for
// convertImage inputs.
var imageFormats = map[string]string{
	".jpg":  "jpeg",
	".jpeg": "jpeg",
	".png":  "png",
	".svg":  "svg",
	".heic": "heic",
	".webp": "webp",
	// PDF is an image input solely for pdf-to-jpeg (rasterization).
	".pdf": "pdf",
}

// documentFormats maps a filename extension to its API input format for
// convertDocument inputs.
var documentFormats = map[string]string{
	".doc":      "doc",
	".docx":     "doc",
	".xls":      "excel",
	".xlsx":     "excel",
	".ppt":      "ppt",
	".pptx":     "ppt",
	".html":     "html",
	".htm":      "html",
	".odt":      "odt",
	".ods":      "ods",
	".odp":      "odp",
	".ots":      "ots",
	".pages":    "pages",
	".numbers":  "numbers",
	".epub":     "epub",
	".md":       "markdown",
	".markdown": "markdown",
	".csv":      "csv",
	".json":     "json",
	".xml":      "xml",
	".yaml":     "yaml",
	".yml":      "yaml",
	".toml":     "toml",
}

// mimeByExt maps a filename extension to its MIME type for multipart
// uploads.
var mimeByExt = map[string]string{
	".jpg":      "image/jpeg",
	".jpeg":     "image/jpeg",
	".png":      "image/png",
	".svg":      "image/svg+xml",
	".heic":     "image/heic",
	".webp":     "image/webp",
	".pdf":      "application/pdf",
	".doc":      "application/msword",
	".docx":     "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":      "application/vnd.ms-excel",
	".xlsx":     "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":      "application/vnd.ms-powerpoint",
	".pptx":     "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".html":     "text/html",
	".htm":      "text/html",
	".odt":      "application/vnd.oasis.opendocument.text",
	".ods":      "application/vnd.oasis.opendocument.spreadsheet",
	".odp":      "application/vnd.oasis.opendocument.presentation",
	".epub":     "application/epub+zip",
	".md":       "text/markdown",
	".markdown": "text/markdown",
	".csv":      "text/csv",
	".json":     "application/json",
	".xml":      "application/xml",
	".yaml":     "application/x-yaml",
	".yml":      "application/x-yaml",
	".toml":     "application/toml",
}

// outputFormatAliases holds common aliases users pass that differ from the
// API's canonical format names.
var outputFormatAliases = map[string]string{
	"jpg": "jpeg",
	"yml": "yaml",
	"htm": "html",
	"md":  "markdown",
}

// extOf returns the lowercased file extension (including the leading dot)
// of name, or "" if it has none.
func extOf(name string) string {
	i := strings.LastIndex(name, ".")
	if i == -1 {
		return ""
	}
	return strings.ToLower(name[i:])
}

// mimeFor returns the MIME type registered for name's extension, or
// "application/octet-stream" if unknown.
func mimeFor(name string) string {
	if mime, ok := mimeByExt[extOf(name)]; ok {
		return mime
	}
	return "application/octet-stream"
}

// resolveInputFormat maps a filename's extension to its API input format
// using the given table, or returns an error listing supported extensions.
func resolveInputFormat(name string, table map[string]string) (string, error) {
	ext := extOf(name)
	fmtName, ok := table[ext]
	if !ok {
		exts := make([]string, 0, len(table))
		for k := range table {
			exts = append(exts, k)
		}
		sort.Strings(exts)
		return "", fmt.Errorf("unsupported file extension '%s'. Supported: %s", ext, strings.Join(exts, ", "))
	}
	return fmtName, nil
}

// normalizeOutputFormat lowercases format, strips a leading dot, and
// resolves aliases (jpg, yml, htm, md).
func normalizeOutputFormat(format string) string {
	f := strings.ToLower(format)
	f = strings.TrimPrefix(f, ".")
	if alias, ok := outputFormatAliases[f]; ok {
		return alias
	}
	return f
}

// ValidOutputsFor lists the output formats the API implements for a given
// input format, sorted alphabetically. For example, ValidOutputsFor("json")
// returns ["csv", "toml", "xml", "yaml"].
func ValidOutputsFor(inputFormat string) []string {
	prefix := inputFormat + "-to-"
	outputs := make([]string, 0)
	for name := range ImplementedConversions {
		if strings.HasPrefix(name, prefix) {
			outputs = append(outputs, strings.TrimPrefix(name, prefix))
		}
	}
	sort.Strings(outputs)
	return outputs
}

// assertConversionImplemented asserts "{input}-to-{output}" is an
// implemented endpoint and returns its name. Returns an error listing the
// valid outputs for that input otherwise.
func assertConversionImplemented(inputFormat, outputFormat string) (string, error) {
	endpoint := inputFormat + "-to-" + outputFormat
	if _, ok := ImplementedConversions[endpoint]; !ok {
		outputs := ValidOutputsFor(inputFormat)
		var hint string
		if len(outputs) > 0 {
			hint = fmt.Sprintf("Supported outputs for '%s': %s", inputFormat, strings.Join(outputs, ", "))
		} else {
			hint = fmt.Sprintf("No conversions are available for input format '%s'", inputFormat)
		}
		return "", fmt.Errorf("conversion '%s' to '%s' is not supported. %s.", inputFormat, outputFormat, hint)
	}
	return endpoint, nil
}
