// Package enconvert is the Go SDK for the Enconvert file conversion API
// (https://enconvert.com).
//
// Convert URLs to PDFs, capture screenshots, extract Markdown, crawl whole
// websites, transform images, and convert documents — all with a single API
// call. The V2 namespace (Client.V2) additionally renders pages into
// agent-ready artifacts, runs web search, extracts structured data, ingests
// sites into RAG-ready JSONL, and watches pages for changes.
//
// # Quick start
//
//	client, err := enconvert.New("sk_...")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	result, err := client.ConvertURLToPDF(ctx, "https://example.com", enconvert.URLToPDFOptions{
//		SaveTo: "page.pdf",
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(result.PresignedURL)
//
// # Errors
//
// Failed requests return *enconvert.APIError. Use IsAuthenticationError,
// IsQuotaError, and IsRateLimitError to check for the conditions the API
// signals with dedicated status codes (401/403, 402, 429 respectively).
package enconvert

// Version is the SDK's release version.
const Version = "0.0.1"
