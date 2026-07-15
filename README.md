# enconvert

Go SDK for the [Enconvert](https://enconvert.com) file conversion API.

Convert URLs to PDFs, capture screenshots, extract Markdown, crawl whole websites, transform images, and convert documents — all with a single API call. Requires Go 1.21+. No third-party dependencies — built entirely on the standard library (`net/http`, `encoding/json`, `mime/multipart`).

## Install

```bash
go get github.com/conversionapi/go-sdk
```

## Quick Start

```go
import (
	"context"

	enconvert "github.com/conversionapi/go-sdk"
)

client, err := enconvert.New("sk_...")
if err != nil {
	log.Fatal(err)
}
```

`New` accepts functional options to override defaults:

```go
client, err := enconvert.New("sk_...",
	enconvert.WithTimeout(60*time.Second),
	enconvert.WithBaseURL("https://api.enconvert.com"),
)
```

Every method that talks to the API takes a `context.Context` as its first argument, so callers control cancellation and deadlines on top of the client's own timeout.

### URL to PDF

```go
result, err := client.ConvertURLToPDF(ctx, "https://example.com", enconvert.URLToPDFOptions{
	SaveTo: "page.pdf",
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(result.PresignedURL)
```

### URL to Screenshot

```go
result, err := client.ConvertURLToScreenshot(ctx, "https://example.com", enconvert.URLToScreenshotOptions{
	URLRenderOptions: enconvert.URLRenderOptions{
		ViewportWidth: enconvert.Int(1440),
	},
	SaveTo: "screenshot.png",
})
```

### URL to Markdown

Extract clean GitHub-Flavored Markdown from any URL — strips nav/footer/ads/scripts, keeps the main article content, and adds YAML frontmatter (title, description, url, links, images).

```go
result, err := client.ConvertURLToMarkdown(ctx, "https://example.com/article", enconvert.URLToMarkdownOptions{
	SaveTo: "article.md",
})
```

### Website to PDF / Screenshot (whole-site batch)

Discover every page of a website (via sitemap, or full crawl on Pro/Business plans), convert each one in the background, and receive a single ZIP. Requires a private API key with crawl access.

```go
batch, err := client.ConvertWebsiteToPDF(ctx, "https://example.com", enconvert.WebsiteToPDFOptions{
	WebsiteConversionOptions: enconvert.WebsiteConversionOptions{
		CrawlMode:       enconvert.CrawlModeSitemap, // Auto (default) | Sitemap | Full
		ExcludePatterns: []string{"/blog/tag/"},     // full crawl mode only
	},
})
fmt.Println(batch.BatchID, batch.URLCount, batch.DiscoveryMethod)

// Block until done and save the ZIP:
status, err := client.WaitForBatch(ctx, batch.BatchID, enconvert.WaitForBatchOptions{SaveTo: "site.zip"})
fmt.Println(status.Completed, "of", status.Total, "pages converted")

// Or poll yourself:
s, err := client.GetBatchStatus(ctx, batch.BatchID)
if s.Status != enconvert.BatchStatusProcessing {
	fmt.Println(s.ZipDownloadURL)
}
```

`ConvertWebsiteToScreenshot` works the same way and produces a ZIP of PNGs.

### Image Conversion

`ConvertImage` and `ConvertDocument` accept any `FileSource`: a `FilePath` (filesystem path), `FileBytes` (raw bytes, uploaded as `upload.bin`), or `FileInput` (raw bytes with an explicit filename/content type).

```go
result, err := client.ConvertImage(ctx, enconvert.FilePath("photo.heic"), enconvert.ConvertImageOptions{
	OutputFormat: "webp",
	SaveTo:       "photo.webp",
})
```

Any pair among `jpeg`, `png`, `svg`, `heic`, `webp` — plus PDF rasterization:

```go
client.ConvertImage(ctx, enconvert.FilePath("scan.pdf"), enconvert.ConvertImageOptions{
	OutputFormat: "jpeg",
	SaveTo:       "scan.jpeg",
})
```

Passing bytes you already have in memory instead of a path:

```go
client.ConvertImage(ctx, enconvert.FileInput{
	Data:     imageBytes,
	Filename: "photo.heic",
}, enconvert.ConvertImageOptions{OutputFormat: "webp", SaveTo: "photo.webp"})
```

### Document Conversion

```go
client.ConvertDocument(ctx, enconvert.FilePath("report.docx"), enconvert.ConvertDocumentOptions{SaveTo: "report.pdf"})
client.ConvertDocument(ctx, enconvert.FilePath("data.json"), enconvert.ConvertDocumentOptions{OutputFormat: "yaml", SaveTo: "data.yaml"})
client.ConvertDocument(ctx, enconvert.FilePath("notes.md"), enconvert.ConvertDocumentOptions{OutputFormat: "html", SaveTo: "notes.html"})
```

Supported inputs: `doc`/`docx`, `xls`/`xlsx`, `ppt`/`pptx`, `odt`, `ods`, `odp`, `ots`, `pages`, `numbers`, `epub`, `html`, `markdown`, `csv`, `json`, `xml`, `yaml`, `toml`

The SDK validates every `{input}-to-{output}` pair against the conversions the API actually implements and returns an error immediately — with the list of valid outputs for that input — instead of sending a doomed request. Introspect programmatically:

```go
enconvert.ValidOutputsFor("json") // ["csv", "toml", "xml", "yaml"]
enconvert.ValidOutputsFor("pdf")  // ["jpeg"]
```

### Supported conversions

| Input | Outputs |
|-------|---------|
| json | csv, toml, xml, yaml |
| xml | csv, json |
| yaml | json |
| csv | json, xml |
| toml | json |
| markdown | html, pdf |
| html | pdf |
| doc, excel, ppt, odt, ods, odp, ots, pages, numbers, epub | pdf |
| jpeg, png, svg, heic, webp | each other (all 20 pairs) |
| pdf | jpeg |

### Job Status (async polling)

```go
status, err := client.GetJobStatus(ctx, "job_abc123")
if status.Status == enconvert.JobStatusSuccess {
	fmt.Println(status.PresignedURL)
}
```

## PDF Options

```go
result, err := client.ConvertURLToPDF(ctx, "https://example.com", enconvert.URLToPDFOptions{
	PDFOptions: &enconvert.PDFOptions{
		PageSize:    "A4", // or custom dimensions via PageWidth + PageHeight
		Orientation: "landscape",
		Margins: &enconvert.PDFMargins{
			Top: enconvert.Float64(10), Bottom: enconvert.Float64(10),
			Left: enconvert.Float64(15), Right: enconvert.Float64(15),
		},
		Header: &enconvert.PDFHeaderFooter{Content: "Quarterly Report", Height: enconvert.Float64(15)},
		Footer: &enconvert.PDFHeaderFooter{Content: "Confidential", Height: enconvert.Float64(12)},
	},
	SaveTo: "report.pdf",
})
```

## Authenticated Pages (plan-gated)

All URL and website conversions accept HTTP Basic Auth, cookies, and custom headers for pages behind a login:

```go
client.ConvertURLToPDF(ctx, "https://internal.example.com/report", enconvert.URLToPDFOptions{
	URLRenderOptions: enconvert.URLRenderOptions{
		Auth: &enconvert.HTTPBasicAuth{Username: "user", Password: "pass"},
		// or cookies / headers:
		Cookies: []enconvert.BrowserCookie{{Name: "session", Value: "abc123", Domain: "internal.example.com"}},
		Headers: map[string]string{"X-Tenant": "acme"},
	},
	SaveTo: "report.pdf",
})
```

Do not combine `Auth` with an `Authorization` header — the API rejects the conflict.

## Error Handling

Failed requests return `*enconvert.APIError`, whose `Error()` string is `"[<status>] <message>"`. Use the `IsXxxError` helpers to check for the conditions the API signals with dedicated status codes:

```go
result, err := client.ConvertURLToPDF(ctx, "https://example.com", enconvert.URLToPDFOptions{})
if err != nil {
	switch {
	case enconvert.IsAuthenticationError(err):
		log.Println("invalid API key")
	case enconvert.IsRateLimitError(err):
		log.Println("too many requests — slow down")
	case enconvert.IsQuotaError(err):
		log.Println("plan feature not enabled or quota exhausted")
	default:
		var apiErr *enconvert.APIError
		if errors.As(err, &apiErr) {
			log.Printf("API error [%d]: %s", apiErr.StatusCode, apiErr.Message)
		} else {
			log.Println(err) // network error, context cancellation, etc.
		}
	}
}
```

## Configuration

```go
client, err := enconvert.New("sk_...",
	enconvert.WithTimeout(300*time.Second), // default
	enconvert.WithBaseURL("https://api.enconvert.com"), // default
)
```

## V2 API (`client.V2`)

The V2 namespace turns web pages into agent-ready data: render, search, extract, ingest, and monitor. All V2 endpoints require a **private API key** and are plan-gated — a disabled feature or exhausted monthly quota returns an error where `IsQuotaError(err)` is true (HTTP 402).

### Perceive — render a URL into artifacts

```go
op, err := client.V2.Perceive(ctx, "https://example.com", enconvert.PerceiveOptions{
	Outputs: []enconvert.PerceiveOutputName{
		enconvert.PerceiveOutputMarkdown,
		enconvert.PerceiveOutputScreenshot,
		enconvert.PerceiveOutputStructured,
	},
	Extract: []enconvert.PerceiveExtractName{enconvert.PerceiveExtractTables, enconvert.PerceiveExtractMetadata},
})
fmt.Println(op.Outputs["markdown"].URL) // 15-min signed URL
fmt.Println(op.Structured)

// Re-sign artifact URLs later:
again, err := client.V2.GetPerceiveOperation(ctx, op.OperationID)

// Batch (<=10 URLs runs inline; larger returns "queued" — poll):
batch, err := client.V2.PerceiveBatch(ctx, []string{"https://a.com", "https://b.com"}, enconvert.PerceiveBatchOptions{
	PerceiveOptions: enconvert.PerceiveOptions{Outputs: []enconvert.PerceiveOutputName{enconvert.PerceiveOutputMarkdown}},
	OutputMode:      enconvert.PerceiveBatchOutputZip,
})
done, err := client.V2.GetPerceiveBatch(ctx, batch.JobID)
```

### Discover — enumerate a site's URLs (no rendering)

```go
found, err := client.V2.Discover(ctx, "https://example.com", enconvert.DiscoverOptions{
	Mode:            enconvert.DiscoverModeHybrid, // Sitemap | Crawl | Hybrid
	MaxURLs:         enconvert.Int(200),
	ExcludePatterns: []string{"/tag/"},
})
fmt.Println(found.Total, found.URLs)
```

### Lookup — web search with optional auto-perceive

```go
search, err := client.V2.Lookup(ctx, "best static site generators", enconvert.LookupOptions{
	Category:    enconvert.LookupCategoryWeb, // web | news | images | scholar | patents | maps
	NumResults:  enconvert.Int(10),
	PerceiveTop: enconvert.Int(3), // auto-render top 3 results (uses perceive quota)
})
for _, hit := range search.Results {
	fmt.Println(hit.Title, hit.URL)
	if hit.Perceive != nil {
		fmt.Println(hit.Perceive.Outputs)
	}
}
```

### Distill — schema-driven structured extraction

```go
extraction, err := client.V2.Distill(ctx, enconvert.DistillOptions{
	URLs:   []string{"https://example.com/pricing"},
	Schema: map[string]any{"plans": "list of plan names with monthly prices"},
	CSSSchema: &enconvert.CSSSchema{ // optional free CSS pass before the LLM tier
		BaseSelector: ".plan-card",
		Fields: []enconvert.CSSField{
			{Name: "name", Type: enconvert.CSSFieldText, Selector: "h3"},
			{Name: "price", Type: enconvert.CSSFieldText, Selector: ".price"},
		},
	},
})
fmt.Println(extraction.Results[0].Data, extraction.Results[0].ExtractionTier)

// Or discover-then-distill:
client.V2.Distill(ctx, enconvert.DistillOptions{
	DiscoverFrom: &enconvert.DistillDiscoverFrom{URL: "https://example.com", Mode: enconvert.DiscoverModeSitemap, MaxPages: enconvert.Int(10)},
	Schema:       map[string]any{"title": "page title", "summary": "one-line summary"},
})
```

### Ingest — site to RAG-ready JSONL (always async)

```go
job, err := client.V2.Ingest(ctx, enconvert.IngestOptions{
	Mode:     enconvert.IngestModeSitemap,
	URL:      "https://docs.example.com",
	MaxPages: enconvert.Int(100),
	Chunk:    &enconvert.IngestChunkOptions{MaxWords: enconvert.Int(512), SentenceOverlap: enconvert.Int(1)},
	WebhookURL: "https://my.app/hooks/enconvert",
})

status, err := client.V2.GetIngestJob(ctx, job.JobID) // poll
if status.Status == enconvert.IngestStatusCompleted {
	fmt.Println(status.OutputURL) // JSONL
}

client.V2.ListIngestJobs(ctx, enconvert.V2ListOptions{Limit: enconvert.Int(20)})
client.V2.CancelIngestJob(ctx, job.JobID) // idempotent

// Webhook signing (HMAC):
secret, err := client.V2.GetWebhookSecret(ctx)
fmt.Println(secret.Secret, secret.SignatureHeader)
client.V2.RotateWebhookSecret(ctx)             // invalidates old secret
client.V2.RetryIngestWebhook(ctx, job.JobID)   // re-deliver
```

### Watch — recurring change monitoring

```go
watcher, err := client.V2.CreateWatcher(ctx, "https://example.com/pricing", enconvert.WatchCreateOptions{
	FrequencyMinutes: enconvert.Int(60),               // hourly floor
	DiffMode:         enconvert.WatchDiffAuto,         // auto | text | structured | tables | metadata
	WebhookURL:       "https://my.app/hooks/changes",
	NotifyEmail:      enconvert.Bool(true),
})

client.V2.ListWatchers(ctx, enconvert.V2ListOptions{})
client.V2.GetWatcher(ctx, watcher.WatcherID)
client.V2.GetWatcherSnapshots(ctx, watcher.WatcherID, enconvert.SnapshotListOptions{Limit: enconvert.Int(10)})
client.V2.UpdateWatcher(ctx, watcher.WatcherID, enconvert.WatcherUpdate{Status: enconvert.WatcherStatusPaused})
client.V2.UpdateWatcher(ctx, watcher.WatcherID, enconvert.WatcherUpdate{WebhookURL: enconvert.String("")}) // clears webhook
client.V2.DeleteWatcher(ctx, watcher.WatcherID) // soft-delete, idempotent
```

### V2 error handling

```go
_, err := client.V2.Ingest(ctx, enconvert.IngestOptions{URLs: []string{"https://example.com"}})
if err != nil {
	if enconvert.IsQuotaError(err) {
		log.Println("upgrade plan or wait for quota reset")
	} else {
		log.Fatal(err)
	}
}
```

## Get an API Key

Sign up at [enconvert.com](https://enconvert.com) to get your API key.

## License

MIT
