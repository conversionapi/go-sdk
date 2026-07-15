package enconvert

import "time"

// ConversionResult is the result of a synchronous (or job-polled) file or
// URL conversion.
type ConversionResult struct {
	PresignedURL          string
	ObjectKey             string
	Filename              string
	FileSize              *int64
	ConversionTimeSeconds *float64
	JobID                 string
}

// JobStatusValue is the lifecycle state of an async conversion job.
type JobStatusValue string

const (
	JobStatusProcessing JobStatusValue = "processing"
	JobStatusSuccess    JobStatusValue = "success"
	JobStatusFailed     JobStatusValue = "failed"
)

// JobStatus is the response from GetJobStatus.
type JobStatus struct {
	Status       JobStatusValue
	PresignedURL string
	ObjectKey    string
	Error        string
}

// PDFMargins are page margins in the PDF's native units (points).
type PDFMargins struct {
	Top    *float64 `json:"top,omitempty"`
	Bottom *float64 `json:"bottom,omitempty"`
	Left   *float64 `json:"left,omitempty"`
	Right  *float64 `json:"right,omitempty"`
}

// PDFHeaderFooter is a header or footer block rendered on each PDF page.
type PDFHeaderFooter struct {
	// Content is text content, max 2000 characters.
	Content string `json:"content,omitempty"`
	// Height is the block height.
	Height *float64 `json:"height,omitempty"`
}

// PDFOptions configures PDF rendering. Only fields explicitly set are sent
// to the API; use pointers for optional numeric/boolean fields so the zero
// value can be distinguished from "unset".
type PDFOptions struct {
	PageSize string
	// PageWidth overrides PageSize when set together with PageHeight.
	PageWidth *float64
	// PageHeight overrides PageSize when set together with PageWidth.
	PageHeight  *float64
	Orientation string // "portrait" | "landscape"
	Margins     *PDFMargins
	Scale       *float64
	Grayscale   *bool
	Header      *PDFHeaderFooter
	Footer      *PDFHeaderFooter
}

// HTTPBasicAuth carries HTTP Basic Auth credentials for pages behind a
// login (plan-gated). Field names match the wire format exactly, since the
// Node SDK passes this object through untransformed.
type HTTPBasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// BrowserCookie is a cookie injected into the browser context before
// rendering (plan-gated). The API requires Name, Value, and either Domain
// or URL. When Domain is set without Path, the API defaults Path to "/".
// Field names match the wire format exactly (including HTTPOnly and
// SameSite), since the Node SDK passes cookie objects through untransformed.
type BrowserCookie struct {
	Name     string   `json:"name"`
	Value    string   `json:"value"`
	Domain   string   `json:"domain,omitempty"`
	URL      string   `json:"url,omitempty"`
	Path     string   `json:"path,omitempty"`
	Expires  *float64 `json:"expires,omitempty"`
	HTTPOnly *bool    `json:"httpOnly,omitempty"`
	Secure   *bool    `json:"secure,omitempty"`
	SameSite string   `json:"sameSite,omitempty"` // "Strict" | "Lax" | "None"
}

// URLRenderOptions holds options shared by all URL-based conversions
// (single page and website).
type URLRenderOptions struct {
	ViewportWidth  *int
	ViewportHeight *int
	LoadMedia      *bool
	EnableScroll   *bool
	OutputFilename string
	// Auth is HTTP Basic Auth for protected pages (plan-gated).
	Auth *HTTPBasicAuth
	// Cookies injected before rendering, max 50 (plan-gated).
	Cookies []BrowserCookie
	// Headers are extra request headers, max 20; hop-by-hop headers
	// rejected (plan-gated).
	Headers map[string]string
}

// URLToPDFOptions configures ConvertURLToPDF.
type URLToPDFOptions struct {
	URLRenderOptions
	SaveTo     string
	SinglePage *bool
	PDFOptions *PDFOptions
}

// URLToScreenshotOptions configures ConvertURLToScreenshot.
type URLToScreenshotOptions struct {
	URLRenderOptions
	SaveTo string
}

// URLToMarkdownOptions configures ConvertURLToMarkdown.
type URLToMarkdownOptions struct {
	URLRenderOptions
	SaveTo string
}

// ConvertImageOptions configures ConvertImage.
type ConvertImageOptions struct {
	// OutputFormat is required.
	OutputFormat   string
	SaveTo         string
	OutputFilename string
}

// ConvertDocumentOptions configures ConvertDocument.
type ConvertDocumentOptions struct {
	// OutputFormat defaults to "pdf" when empty.
	OutputFormat   string
	SaveTo         string
	OutputFilename string
	PDFOptions     *PDFOptions
}

// CrawlMode is the URL discovery strategy for website conversions.
type CrawlMode string

const (
	// CrawlModeAuto uses the highest mode the plan allows (default).
	CrawlModeAuto CrawlMode = "auto"
	// CrawlModeSitemap uses sitemap.xml only (Starter and above).
	CrawlModeSitemap CrawlMode = "sitemap"
	// CrawlModeFull uses sitemap plus BFS crawl (Pro/Business).
	CrawlModeFull CrawlMode = "full"
)

// WebsiteConversionOptions holds options shared by ConvertWebsiteToPDF and
// ConvertWebsiteToScreenshot.
type WebsiteConversionOptions struct {
	URLRenderOptions
	CrawlMode CrawlMode
	// IncludePatterns only crawls URLs matching these patterns (full crawl
	// mode).
	IncludePatterns []string
	// ExcludePatterns skips URLs matching these patterns (full crawl mode).
	ExcludePatterns []string
	// NotificationEmail is the email notified on completion. Defaults to
	// the project owner's email.
	NotificationEmail string
	// CallbackURL is a webhook POSTed when the batch finishes (plan-gated).
	CallbackURL string
}

// WebsiteToPDFOptions configures ConvertWebsiteToPDF.
type WebsiteToPDFOptions struct {
	WebsiteConversionOptions
	SinglePage *bool
	PDFOptions *PDFOptions
}

// WebsiteToScreenshotOptions configures ConvertWebsiteToScreenshot.
type WebsiteToScreenshotOptions struct {
	WebsiteConversionOptions
}

// BatchSubmission is the 202 response from an async batch submission
// (website conversions).
type BatchSubmission struct {
	BatchID string
	// Status is always "processing" on submission.
	Status string
	// URLCount is the number of pages queued for conversion.
	URLCount int
	// TotalDiscovered is the total URLs found during discovery (before
	// plan limits applied).
	TotalDiscovered *int
	// DiscoveryMethod is how URLs were discovered: "sitemap" or
	// "full_crawl".
	DiscoveryMethod string
	// OutputFormat is the output packaging, "zip" for website conversions.
	OutputFormat string
}

// BatchStatusValue is the lifecycle state of a batch conversion.
type BatchStatusValue string

const (
	BatchStatusProcessing BatchStatusValue = "processing"
	BatchStatusCompleted  BatchStatusValue = "completed"
	BatchStatusPartial    BatchStatusValue = "partial"
	BatchStatusFailed     BatchStatusValue = "failed"
)

// BatchItem is a per-URL entry in a batch status response.
type BatchItem struct {
	SourceURL string
	// Status is the raw activity status: "In Progress", "Success", or
	// "Failed".
	Status         string
	DownloadURL    string
	OutputFileSize *int64
	Duration       string
}

// BatchStatus is the response from GetBatchStatus and WaitForBatch.
type BatchStatus struct {
	BatchID    string
	Status     BatchStatusValue
	Total      int
	Completed  int
	Failed     int
	InProgress int
	OutputMode string // "zip" | "individual"
	// ZipDownloadURL is the presigned URL of the bundled ZIP when
	// OutputMode is "zip".
	ZipDownloadURL string
	Items          []BatchItem
}

// WaitForBatchOptions configures WaitForBatch. Zero values fall back to the
// documented defaults (5s interval, 30m timeout).
type WaitForBatchOptions struct {
	// Interval between polls. Defaults to 5 seconds.
	Interval time.Duration
	// Timeout gives up after this duration. Defaults to 30 minutes.
	Timeout time.Duration
	// SaveTo saves the batch ZIP to this local path once available.
	SaveTo string
}

// Pointer helpers for the many optional (pointer-typed) option fields —
// e.g. enconvert.Int(2), enconvert.Bool(true) — so callers don't need a
// throwaway local variable to take an address.

// Bool returns a pointer to v.
func Bool(v bool) *bool { return &v }

// Int returns a pointer to v.
func Int(v int) *int { return &v }

// Float64 returns a pointer to v.
func Float64(v float64) *float64 { return &v }

// String returns a pointer to v. Mainly useful for WatcherUpdate.WebhookURL,
// where a pointer to "" explicitly clears the webhook.
func String(v string) *string { return &v }
