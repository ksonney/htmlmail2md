// Package converter turns HTML content into simplified Markdown.
package converter

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

// Convert parses an HTML document and returns a simplified Markdown
// representation of its body content.
func Convert(input []byte) (string, error) {
	doc, err := html.Parse(bytes.NewReader(input))
	if err != nil {
		return "", err
	}
	c := &converter{}
	var blocks []string
	c.collect(doc, &blocks)
	return strings.Join(blocks, "\n\n") + "\n", nil
}

type converter struct{}

// collect appends markdown blocks for the visual children of n to blocks.
// Elements that carry no markdown representation of their own are skipped.
func (c *converter) collect(n *html.Node, blocks *[]string) {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		switch child.Type {
		case html.ElementNode:
			if skipTags[child.Data] {
				continue
			}
			switch {
			case isListTag(child.Data):
				*blocks = append(*blocks, c.renderList(child, 0))
			case isBlockTag(child.Data):
				if b := c.renderBlock(child); b != "" {
					*blocks = append(*blocks, b)
				}
			case isContainerTag(child.Data):
				c.collect(child, blocks)
			default:
				if b := c.inlineText(child); b != "" {
					*blocks = append(*blocks, b)
				}
			}
		case html.TextNode:
			if t := strings.TrimSpace(child.Data); t != "" {
				*blocks = append(*blocks, t)
			}
		}
	}
}

// renderBlock renders a standalone block level element to Markdown.
func (c *converter) renderBlock(n *html.Node) string {
	switch n.Data {
	case "h1":
		return "# " + escapeText(c.inlineText(n))
	case "h2":
		return "## " + escapeText(c.inlineText(n))
	case "h3":
		return "### " + escapeText(c.inlineText(n))
	case "h4":
		return "#### " + escapeText(c.inlineText(n))
	case "h5":
		return "##### " + escapeText(c.inlineText(n))
	case "h6":
		return "###### " + escapeText(c.inlineText(n))
	case "p":
		return c.inlineText(n)
	case "blockquote":
		return c.renderBlockquote(n)
	case "pre":
		return c.renderPre(n)
	case "hr":
		return "---"
	}
	return ""
}

// renderBlockquote renders a blockquote, prefixing every line with "> ".
func (c *converter) renderBlockquote(n *html.Node) string {
	var blocks []string
	c.collect(n, &blocks)
	var lines []string
	for _, b := range blocks {
		for _, l := range strings.Split(b, "\n") {
			lines = append(lines, "> "+strings.TrimSpace(l))
		}
	}
	return strings.Join(lines, "\n")
}

// renderPre renders a pre element as a fenced code block.
func (c *converter) renderPre(n *html.Node) string {
	var sb strings.Builder
	sb.WriteString("```")
	if lang := codeLanguage(n); lang != "" {
		sb.WriteString(" " + lang)
	}
	sb.WriteString("\n")
	sb.WriteString(codeContent(n))
	sb.WriteString("\n```")
	return sb.String()
}

// renderList renders a ul or ol element as a Markdown list. Nested lists
// are indented relative to their parent item.
func (c *converter) renderList(list *html.Node, depth int) string {
	var prefix string
	if list.Data == "ol" {
		prefix = "1. "
	} else {
		prefix = "- "
	}
	prefix = strings.Repeat("  ", depth) + prefix

	var lines []string
	for li := list.FirstChild; li != nil; li = li.NextSibling {
		if li.Type != html.ElementNode || li.Data != "li" {
			continue
		}
		lines = append(lines, c.renderListItem(li, prefix, depth)...)
	}
	return strings.Join(lines, "\n")
}

// renderListItem renders a single list item, including any nested lists.
// It returns one or more markdown lines for the item.
func (c *converter) renderListItem(li *html.Node, prefix string, depth int) []string {
	var inline strings.Builder
	var nested []*html.Node

	for child := li.FirstChild; child != nil; child = child.NextSibling {
		switch {
		case child.Type == html.ElementNode && (child.Data == "ul" || child.Data == "ol"):
			nested = append(nested, child)
		case child.Type == html.TextNode:
			inline.WriteString(child.Data)
		case child.Type == html.ElementNode && !skipTags[child.Data]:
			c.writeInline(child, &inline)
		}
	}

	text := strings.TrimSpace(strings.Join(strings.Fields(inline.String()), " "))
	var lines []string
	if text != "" {
		lines = append(lines, prefix+text)
	}
	for _, sub := range nested {
		lines = append(lines, strings.Split(c.renderList(sub, depth+1), "\n")...)
	}
	return lines
}

// inlineText renders the inline content of n as a single line of Markdown
// text, collapsing whitespace and trimming surrounding space.
func (c *converter) inlineText(n *html.Node) string {
	var sb strings.Builder
	c.writeInline(n, &sb)
	return strings.TrimSpace(strings.Join(strings.Fields(sb.String()), " "))
}

// writeInline writes the inline content of n into sb, transforming inline
// HTML elements into their Markdown equivalents.
func (c *converter) writeInline(n *html.Node, sb *strings.Builder) {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		switch child.Type {
		case html.TextNode:
			sb.WriteString(child.Data)
		case html.ElementNode:
			switch child.Data {
			case "a":
				inner := c.inlineText(child)
				if href := attrValue(child, "href"); href != "" {
					sb.WriteString("[" + escapeText(inner) + "](" + escapeURL(href) + ")")
				} else {
					sb.WriteString(inner)
				}
			case "img":
				sb.WriteString("![" + escapeText(attrValue(child, "alt")) + "](" + escapeURL(attrValue(child, "src")) + ")")
			case "br":
				sb.WriteString("\n")
			case "strong", "b":
				var inner strings.Builder
				c.writeInline(child, &inner)
				sb.WriteString("**")
				sb.WriteString(strings.TrimSpace(inner.String()))
				sb.WriteString("**")
			case "em", "i":
				var inner strings.Builder
				c.writeInline(child, &inner)
				sb.WriteString("_")
				sb.WriteString(strings.TrimSpace(inner.String()))
				sb.WriteString("_")
			case "code":
				sb.WriteString("`")
				sb.WriteString(codeSpan(child))
				sb.WriteString("`")
			default:
				c.writeInline(child, sb)
			}
		}
	}
}

// codeContent returns the text content of a pre/code element.
func codeContent(n *html.Node) string {
	var sb strings.Builder
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			switch child.Type {
			case html.TextNode:
				sb.WriteString(child.Data)
			case html.ElementNode:
				walk(child)
			}
		}
	}
	walk(n)
	return strings.TrimSpace(sb.String())
}

// codeLanguage returns the language identifier from a code element's class.
func codeLanguage(n *html.Node) string {
	var lang string
	var visit func(n *html.Node)
	visit = func(n *html.Node) {
		for _, a := range n.Attr {
			if strings.EqualFold(a.Key, "class") {
				for _, tok := range strings.Fields(a.Val) {
					if strings.HasPrefix(tok, "language-") {
						lang = strings.TrimPrefix(tok, "language-")
						return
					}
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode {
				visit(child)
			}
		}
	}
	visit(n)
	return lang
}

// codeSpan renders a code element as an inline code span.
func codeSpan(n *html.Node) string {
	s := codeContent(n)
	if strings.Contains(s, "`") {
		return strings.ReplaceAll(s, "`", "``")
	}
	return s
}

// attrValue returns an HTML attribute value by name.
func attrValue(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, name) {
			return a.Val
		}
	}
	return ""
}

// escapeText escapes characters that are significant in Markdown text.
func escapeText(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "`", "\\`")
	s = strings.ReplaceAll(s, "*", "\\*")
	s = strings.ReplaceAll(s, "_", "\\_")
	s = strings.ReplaceAll(s, "[", "\\[")
	s = strings.ReplaceAll(s, "]", "\\]")
	s = strings.ReplaceAll(s, "#", "\\#")
	s = strings.ReplaceAll(s, ">", "\\>")
	return s
}

// escapeURL escapes characters that are not allowed unescaped in a link
// destination.
func escapeURL(s string) string {
	s = strings.ReplaceAll(s, "(", "%28")
	s = strings.ReplaceAll(s, ")", "%29")
	s = strings.ReplaceAll(s, " ", "%20")
	return s
}

// isListTag reports whether the tag renders as a Markdown list.
func isListTag(tag string) bool {
	return tag == "ul" || tag == "ol"
}

// isBlockTag reports whether the tag renders as a standalone Markdown block.
func isBlockTag(tag string) bool {
	switch tag {
	case "h1", "h2", "h3", "h4", "h5", "h6", "p", "blockquote", "pre", "hr":
		return true
	}
	return false
}

// isContainerTag reports whether the element wraps other blocks.
func isContainerTag(tag string) bool {
	switch tag {
	case "div", "section", "article", "header", "footer", "nav", "main", "aside",
		"figure", "figcaption", "address", "dl", "table", "tr", "td", "th",
		"tbody", "thead", "tfoot", "body", "html":
		return true
	}
	return false
}

// skipTags are elements that carry no Markdown representation of their own.
var skipTags = map[string]bool{
	"head": true, "title": true, "meta": true, "link": true, "style": true,
	"script": true, "base": true, "template": true, "iframe": true, "svg": true,
	"noscript": true, "area": true, "param": true, "source": true, "track": true,
	"col": true, "colgroup": true,
}
