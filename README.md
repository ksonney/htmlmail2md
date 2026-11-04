# htmlmail2md

Converts HTML email content to Markdown.

## Usage

```
htmlmail2md <input.html> [-o output.md]
```

- Reads HTML from the file given as the first argument, or from standard
  input when the argument is `-` or omitted.
- Writes the Markdown result to the file given with `-o`/`--output`, or to
  standard output.

## Example

```console
$ htmlmail2md newsletter.html -o newsletter.md
```

## Conversion notes

The converter walks the HTML document and emits Markdown blocks for:

- Headings (`#` through `######`)
- Paragraphs
- Unordered (`-`) and ordered (`1.`) lists, including nested lists
- Blockquotes (`>`)
- Fenced code blocks (```` ``` ````) for `<pre>`/`<code>` elements, using the
  language from the element's `class` attribute when present
- Thematic breaks (`---`) for `<hr>`
- Links (`[text](url)`) and images (`![alt](src)`)
- Emphasis via bold (`**`) and italics (`_`)

Headings, plain text, and link text are escaped where Markdown-significant
characters would otherwise be interpreted. `<head>` content such as styles and
titles is skipped.

## Development

Requires Go 1.26 or newer.

```console
go build ./...
go test ./...
```
