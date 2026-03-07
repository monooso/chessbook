package pgn

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// Game represents a parsed PGN game: tag pairs, move list, and result.
// Moves are stored as raw SAN strings; no move validation is performed.
type Game struct {
	Tags   map[string]string
	Moves  []string
	Result string
}

// Parse parses a PGN stream into a slice of games.
// Comments, recursive annotation variations (RAVs), and numeric annotation
// glyphs (NAGs) are discarded. Move numbers are stripped.
func Parse(r io.Reader) ([]Game, error) {
	scanner := bufio.NewScanner(r)
	var games []Game
	var currentTags map[string]string
	var movetext strings.Builder
	inMovetext := false

	for scanner.Scan() {
		line := scanner.Text()

		// Strip semicolon comments (rest of line after ;).
		if idx := strings.IndexByte(line, ';'); idx >= 0 {
			line = line[:idx]
		}

		trimmed := strings.TrimSpace(line)

		// Tag pair line.
		if strings.HasPrefix(trimmed, "[") {
			if inMovetext {
				// We've hit a new game's tags. Finish the current game.
				g, err := parseGame(currentTags, movetext.String())
				if err != nil {
					return nil, err
				}
				games = append(games, g)
				currentTags = nil
				movetext.Reset()
				inMovetext = false
			}
			if currentTags == nil {
				currentTags = make(map[string]string)
			}
			key, value, err := parseTag(trimmed)
			if err != nil {
				return nil, err
			}
			currentTags[key] = value
			continue
		}

		// Empty line: separator between tags and movetext, or between games.
		if trimmed == "" {
			continue
		}

		// Movetext line.
		inMovetext = true
		movetext.WriteByte(' ')
		movetext.WriteString(trimmed)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("pgn: read error: %w", err)
	}

	// Finish last game.
	if currentTags != nil || movetext.Len() > 0 {
		g, err := parseGame(currentTags, movetext.String())
		if err != nil {
			return nil, err
		}
		games = append(games, g)
	}

	return games, nil
}

func parseTag(line string) (string, string, error) {
	// Expected format: [Key "Value"]
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "[") || !strings.HasSuffix(line, "]") {
		return "", "", fmt.Errorf("pgn: invalid tag pair: %q", line)
	}

	inner := line[1 : len(line)-1]
	before, after, ok := strings.Cut(inner, " ")
	if !ok {
		return "", "", fmt.Errorf("pgn: invalid tag pair: %q", line)
	}

	key := before
	valueStr := strings.TrimSpace(after)

	// Remove surrounding quotes.
	if len(valueStr) < 2 || valueStr[0] != '"' || valueStr[len(valueStr)-1] != '"' {
		return "", "", fmt.Errorf("pgn: invalid tag value in %q", line)
	}
	value := valueStr[1 : len(valueStr)-1]

	return key, value, nil
}

func parseGame(tags map[string]string, movetext string) (Game, error) {
	if tags == nil {
		tags = make(map[string]string)
	}

	moves, result := tokenizeMovetext(movetext)

	return Game{
		Tags:   tags,
		Moves:  moves,
		Result: result,
	}, nil
}

// result tokens.
func isResult(s string) bool {
	return s == "1-0" || s == "0-1" || s == "1/2-1/2" || s == "*"
}

// isMoveNumber returns true if the token is a move number (e.g. "1." or "1...").
func isMoveNumber(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Must start with a digit and end with one or more dots.
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 {
		return false
	}
	for i < len(s) && s[i] == '.' {
		i++
	}
	return i == len(s)
}

// tokenizeMovetext extracts SAN moves and the result from movetext.
// Comments ({...}), RAVs ((...)), NAGs ($N), and move numbers are stripped.
func tokenizeMovetext(s string) ([]string, string) {
	var moves []string
	var result string
	i := 0

	for i < len(s) {
		ch := s[i]

		// Skip whitespace.
		if unicode.IsSpace(rune(ch)) {
			i++
			continue
		}

		// Skip brace comments.
		if ch == '{' {
			depth := 1
			i++
			for i < len(s) && depth > 0 {
				if s[i] == '{' {
					depth++
				} else if s[i] == '}' {
					depth--
				}
				i++
			}
			continue
		}

		// Skip RAVs.
		if ch == '(' {
			depth := 1
			i++
			for i < len(s) && depth > 0 {
				if s[i] == '(' {
					depth++
				} else if s[i] == ')' {
					depth--
				}
				i++
			}
			continue
		}

		// Skip NAGs.
		if ch == '$' {
			i++
			for i < len(s) && s[i] >= '0' && s[i] <= '9' {
				i++
			}
			continue
		}

		// Read a token (non-whitespace sequence).
		start := i
		for i < len(s) && !unicode.IsSpace(rune(s[i])) && s[i] != '{' && s[i] != '(' && s[i] != ')' {
			i++
		}
		token := s[start:i]

		if isResult(token) {
			result = token
			continue
		}

		if isMoveNumber(token) {
			continue
		}

		if token != "" {
			moves = append(moves, token)
		}
	}

	return moves, result
}
