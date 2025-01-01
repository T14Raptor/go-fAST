package scanner

func isLineTerminator(chr rune) bool {
	switch chr {
	case '\u000a', '\u000d', '\u2028', '\u2029':
		return true
	}
	return false
}

func (s *Scanner) skipSingleLineComment() {
	for {
		p, ok := s.PeekRune()
		if !ok {
			break
		}
		if isLineTerminator(p) {
			continue
		}
		break
	}
}

func (s *Scanner) skipMultiLineComment() (hasLineTerminator bool) {
	var p rune
	for {
		var ok bool
		p, ok = s.NextRune()
		if !ok {
			break
		}

		if p == '\r' || p == '\n' || p == '\u2028' || p == '\u2029' {
			hasLineTerminator = true
			break
		}

		next := s.ConsumeRune()
		if p == '*' && next == '/' {
			s.ConsumeRune()
			return
		}
	}
	for p >= 0 {
		oldChr := p
		chr := s.ConsumeRune()
		if oldChr == '*' && chr == '/' {
			s.ConsumeRune()
			return
		}
	}

	//p.errorUnexpected(p.chr)
	return
}
