package validator

import (
	"fmt"
	"main/internal/shared"
	"os"
)

const (
	TokenEndOfStream = iota
	TokenError
	TokenOpenBrace
	TokenOpenBracket
	TokenCloseBrace
	TokenCloseBracket
	TokenComma
	TokenColon
	TokenSemiColon
	TokenStringLiteral
	TokenNumber
	TokenTrue
	TokenFalse
	TokenNull
	TokenCount
)

type Token struct {
	Type   int
	Start  int
	Length int
}

type Element struct {
	Label           string
	Value           string
	FirstSubElement *Element
	NextSibling     *Element
}

type Parser struct {
	Source   []byte
	At       int
	HadError bool
}

func IsJSONDigit(source []byte, at int) bool {
	result := false
	if IsInBounds(source, at) {
		val := source[at]
		result = '0' <= val && val <= '9'
	}
	return result
}

func IsJSONWhitespace(source []byte, at int) bool {
	result := false
	if IsInBounds(source, at) {
		val := source[at]
		result = val == ' ' || val == '\t' || val == '\n' || val == '\r'
	}
	return result
}

func (p *Parser) IsParsing() bool {
	return !p.HadError && IsInBounds(p.Source, p.At)
}

func (p *Parser) Error(token Token, message string) {
	p.HadError = true
	fmt.Fprintf(os.Stderr, "ERROR: \"%s\" - %s\n", p.ExtractTokenValue(token), message)
}

func (p *Parser) ParseKeyword(keyword string, result *Token) {
	keywordBytes := []byte(keyword)
	if (len(p.Source) - p.At) >= len(keywordBytes) {
		check := string(p.Source[p.At : p.At+len(keywordBytes)])
		if check == keyword {
			result.Type = keywordToTokenType(keyword)
			result.Start = p.At
			result.Length = len(keywordBytes)
			p.At += len(keywordBytes)
		}
	}
}

func keywordToTokenType(keyword string) int {
	switch keyword {
	case "false":
		return TokenFalse
	case "null":
		return TokenNull
	case "true":
		return TokenTrue
	default:
		return TokenError //Shouldn't reach here
	}
}

func (p *Parser) GetJSONToken() Token {
	var result Token
	source := p.Source
	at := p.At

	for p.IsParsing() && IsJSONWhitespace(source, at) {
		at++
		p.At = at
	}

	if p.IsParsing() {
		result.Type = TokenError
		result.Start = at
		result.Length = 1

		switch source[at] {
		case '{':
			result.Type = TokenOpenBrace
			at++
		case '[':
			result.Type = TokenOpenBracket
			at++
		case '}':
			result.Type = TokenCloseBrace
			at++
		case ']':
			result.Type = TokenCloseBracket
			at++
		case ',':
			result.Type = TokenComma
			at++
		case ':':
			result.Type = TokenColon
			at++
		case ';':
			result.Type = TokenSemiColon
			at++
		case 'f':
			p.ParseKeyword("false", &result)
		case 't':
			p.ParseKeyword("true", &result)
		case 'n':
			p.ParseKeyword("null", &result)
		case '"':
			result.Type = TokenStringLiteral
			start := at + 1
			at++ //skip opening quote

			for p.IsParsing() && source[at] != '"' {
				if at+1 < len(source) && source[at] == '\\' && source[at+1] == '"' {
					at++ //skip escaped quote
				}
				at++
			}
			result.Start = start
			result.Length = at - start
			if p.IsParsing() && source[at] == '"' {
				at++ //skip closing quote
			}
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			start := at
			result.Type = TokenNumber

			if source[at] == '-' {
				at++
			}

			//parse before decimal
			for p.IsParsing() && IsJSONDigit(source, at) {
				at++
			}

			//Handle decimal point and digits after
			if p.IsParsing() && source[at] == '.' {
				at++
				for p.IsParsing() && IsJSONDigit(source, at) {
					at++
				}
			}

			if p.IsParsing() && (source[at] == 'e' || source[at] == 'E') {
				at++
				if p.IsParsing() && (source[at] == '+' || source[at] == '-') {
					at++
				}
				for p.IsParsing() && IsJSONDigit(source, at) {
					at++
				}
			}
			result.Start = start
			result.Length = at - start
		default:
			// Leave Type as TokenError, Start as at, Length as 1
		}
		p.At = at
	}
	return result
}

func (p *Parser) ExtractTokenValue(token Token) string {
	if !IsInBounds(p.Source, token.Start) || token.Length <= 0 {
		return "" //Error neg start location
	}
	if token.Start+token.Length > len(p.Source) {
		return "" //Token is out of bounds
	}
	return string(p.Source[token.Start : token.Start+token.Length])
}

func (p *Parser) _ParseJSONList(startingToken Token, endType int, hasLabels bool) *Element {
	var firstElement *Element
	var lastElement *Element

	for p.IsParsing() {
		var label string
		valueToken := p.GetJSONToken()

		if hasLabels {
			if valueToken.Type == TokenStringLiteral {
				label = p.ExtractTokenValue(valueToken)
				colon := p.GetJSONToken()
				if colon.Type == TokenColon {
					valueToken = p.GetJSONToken()
				} else {
					p.Error(colon, "Expected colon after field name")
				}
			} else if valueToken.Type != endType {
				p.Error(valueToken, "Unexpected token in JSON")
			}
		}

		element := p._ParseJSONElement(label, valueToken)

		if element != nil {
			if lastElement == nil {
				firstElement = element
				lastElement = element
			} else {
				lastElement.NextSibling = element
				lastElement = element
			}
		} else if valueToken.Type == endType {
			break
		} else {
			p.Error(valueToken, "Unexpected token in JSON")
		}

		comma := p.GetJSONToken()
		if comma.Type == endType {
			break
		} else if comma.Type != TokenComma {
			p.Error(comma, "Unexpected token in JSON. Expected ,")
		}

	}
	return firstElement
}

func (p *Parser) _ParseJSONElement(label string, valueToken Token) *Element {
	valid := true
	var subElement *Element
	valueType := valueToken.Type

	if valueType == TokenOpenBracket {
		subElement = p._ParseJSONList(valueToken, TokenCloseBracket, false)
	} else if valueType == TokenOpenBrace {
		subElement = p._ParseJSONList(valueToken, TokenCloseBrace, true)
	} else if valueType == TokenStringLiteral || valueType == TokenTrue || valueType == TokenFalse || valueType == TokenNull || valueType == TokenNumber {
		// Nothing to do here, the value is a literal; store index and length

	} else {
		valid = false
	}

	var result *Element

	if valid {
		result = new(Element)
		result.Label = label
		result.Value = p.ExtractTokenValue(valueToken)
		result.FirstSubElement = subElement
		result.NextSibling = nil
	}

	return result
}

func (p *Parser) ParseJSON(inputsJSON []byte) *Element {
	p.Source = inputsJSON
	p.At = 0
	p.HadError = false
	token := p.GetJSONToken()
	return p._ParseJSONElement("", token)
}

func ConvertElementToFloat64(object *Element, elementName string) float64 {
	result := 0.0
	element := LookupElement(object, elementName)

	if element != nil {
		source := element.Value
		at := 0

		sign := 1.0
		if len(source) > at && source[at] == '-' {
			sign = -1.0
			at++
		}

		//Convert number
		number := 0.0
		for len(source) > at && IsJSONDigit([]byte(source), at) {
			char := float64(source[at] - '0') // converts ASCII values to numerical value
			number = 10.0*number + char
			at++
		}

		//Handle decimal
		if len(source) > at && source[at] == '.' {
			at++
			c := 1.0 / 10.0
			for len(source) > at && IsJSONDigit([]byte(source), at) {
				char := source[at] - '0'
				number = number + c*float64(char)
				c *= 1.0 / 10.0
				at++
			}
		}

		//Handle scientific notation
		if len(source) > at && (source[at] == 'e' || source[at] == 'E') {
			at++
			exponentSign := 1.0
			if len(source) > at && (source[at] == '+' || source[at] == '-') {
				if source[at] == '-' {
					exponentSign = -1.0
				}
				at++
			}

			exponent := 0.0
			for len(source) > at && IsJSONDigit([]byte(source), at) {
				char := source[at] - '0'
				exponent = 10.0*exponent + float64(char)
				at++
			}

			number *= pow(10.0, exponentSign*exponent)
		}
		result = sign * number
	}
	return result
}

func pow(x, y float64) float64 {
	result := 1.0
	for range int(y) {
		result *= x
	}
	return result
}

func LookupElement(object *Element, elementName string) *Element {
	result := (*Element)(nil)

	if object != nil {
		for search := object.FirstSubElement; search != nil; search = search.NextSibling {
			if search.Label == elementName {
				result = search
				break
			}
		}
	}
	return result
}

func IsInBounds(source []byte, at int) bool {
	return at >= 0 && at < len(source)
}

func ParseHaversinePairs(inputJSON []byte, maxPairCount int, pairs []shared.HaversinePair) int {
	parser := Parser{Source: inputJSON}

	JSON := parser.ParseJSON(inputJSON)

	pairCount := 0
	pairsArray := LookupElement(JSON, "pairs")

	if pairsArray != nil {
		for element := pairsArray.FirstSubElement; element != nil && pairCount < maxPairCount; element = element.NextSibling {
			pair := &pairs[pairCount]
			pair.X0 = ConvertElementToFloat64(element, "X0")
			pair.Y0 = ConvertElementToFloat64(element, "Y0")
			pair.X1 = ConvertElementToFloat64(element, "X1")
			pair.Y1 = ConvertElementToFloat64(element, "Y1")
			pairCount++
		}
	}
	return pairCount
}
