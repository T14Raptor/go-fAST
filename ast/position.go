package ast

// Positioned is implemented by structured errors that carry a source
// range. Concrete error types in scanner and parser implement it so
// callers can pull positions without knowing which layer produced the
// error:
//
//	var p ast.Positioned
//	if errors.As(err, &p) {
//	    start, _ := p.Pos()
//	    line, col, _ := ast.LineColumn(src, start)
//	    // ...
//	}
type Positioned interface {
	error
	Pos() (start, end Idx)
}

// LineColumn returns the 1-based line and column, together with the byte
// offset, for byte offset idx within src. An idx past the end of src is
// clamped to the end (and returned as offset).
//
// Line and column are 1-based. Column counts UTF-8 bytes from the start of
// the line, not Unicode code points or grapheme clusters; downstream
// formatters that need codepoint columns should re-derive them from the
// source. Line breaks are LF (\n) and CRLF (\r\n is one logical break); a
// bare CR is also treated as a line break for compatibility with older JS
// sources.
//
// LineColumn runs in O(idx). Callers that need to translate many indices
// against the same source should build their own index.
func LineColumn(src string, idx Idx) (line, column, offset int) {
	end := int(idx)
	if end > len(src) {
		end = len(src)
	}
	line = 1
	lineStart := 0
	for i := 0; i < end; i++ {
		switch src[i] {
		case '\n':
			line++
			lineStart = i + 1
		case '\r':
			line++
			if i+1 < end && src[i+1] == '\n' {
				i++
			}
			lineStart = i + 1
		}
	}
	return line, end - lineStart + 1, end
}
