# Data Sources

**yams** is only able to reliably simulate identities and resources it knows about, so providing
accurate and complete data is crucial.

To understand **Sources** in the abstract, please see [Concepts > Sources](./concepts.md).

**Sources** can have various schemas, formats, and locations, which are inferred based on a string
shorthand.

### **Schemas**

* AWS Config (default)

### **Formats**

* JSON (`.json` suffix)
* JSON-L (`.jsonl` suffix)

### **Locations**

* Local file (default, or via `file://` prefix)
* S3 object (`s3://` prefix)

Additionally, compressed **Source** files are supported for all formats and typically offer improved
load performance for larger environments:

* gzip-compressed files (`.gz` suffix)

### Examples

| Source string            | Explanation |
| ------------------------ | ----------- |
| `resources.json`         | A local file with name `resources.json`; formatted as a JSON array
| `file://loadme.jsonl.gz` | A gzip-compressed local file with name `loadme.jsonl.gz`; formatted as newline-separated JSON objects
| `s3://mybucket/resources.json.gz` | A gzip-compressed object in the S3 bucket `mybucket` with key `resources.json.gz`; formatted as a JSON array
