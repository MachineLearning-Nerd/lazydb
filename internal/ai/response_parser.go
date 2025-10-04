package ai

import (
	"regexp"
	"strings"
)

// ResponseSection represents a parsed section of an AI response
type ResponseSection struct {
	Type    string // "code", "text", "list", "header", "query"
	Content string
	Number  int
	Title   string // For headers or inferred titles
}

// ParseResponse intelligently splits an AI response into sections
func ParseResponse(response string) []ResponseSection {
	if response == "" {
		return nil
	}

	sections := []ResponseSection{}
	sectionNum := 1

	// Split by code blocks first
	parts := splitByCodeBlocks(response)

	for _, part := range parts {
		if part.isCode {
			// This is a code block
			lang := extractCodeLanguage(part.content)
			title := "Code"
			if lang != "" {
				title = strings.ToUpper(lang) + " Code"
			}

			sections = append(sections, ResponseSection{
				Type:    "code",
				Content: part.content,
				Number:  sectionNum,
				Title:   title,
			})
			sectionNum++
		} else {
			// This is regular text - might have headers, lists, etc.
			textSections := splitTextIntoSections(part.content, sectionNum)
			sections = append(sections, textSections...)
			sectionNum += len(textSections)
		}
	}

	return sections
}

// codePart represents a part that is either code or text
type codePart struct {
	content string
	isCode  bool
}

// splitByCodeBlocks splits content by code blocks (```...```)
func splitByCodeBlocks(content string) []codePart {
	parts := []codePart{}

	// Regex to match code blocks
	codeBlockRegex := regexp.MustCompile("(?s)```[\\w]*\\n(.*?)```")

	lastIndex := 0
	matches := codeBlockRegex.FindAllStringSubmatchIndex(content, -1)

	for _, match := range matches {
		// Add text before code block
		if match[0] > lastIndex {
			textContent := strings.TrimSpace(content[lastIndex:match[0]])
			if textContent != "" {
				parts = append(parts, codePart{
					content: textContent,
					isCode:  false,
				})
			}
		}

		// Add code block content (group 1 is the content inside ```)
		codeContent := content[match[2]:match[3]]
		parts = append(parts, codePart{
			content: strings.TrimSpace(codeContent),
			isCode:  true,
		})

		lastIndex = match[1]
	}

	// Add remaining text after last code block
	if lastIndex < len(content) {
		textContent := strings.TrimSpace(content[lastIndex:])
		if textContent != "" {
			parts = append(parts, codePart{
				content: textContent,
				isCode:  false,
			})
		}
	}

	// If no code blocks found, return entire content as text
	if len(parts) == 0 {
		parts = append(parts, codePart{
			content: strings.TrimSpace(content),
			isCode:  false,
		})
	}

	return parts
}

// extractCodeLanguage extracts the language from a code block fence
func extractCodeLanguage(codeBlock string) string {
	lines := strings.Split(codeBlock, "\n")
	if len(lines) == 0 {
		return ""
	}

	firstLine := strings.TrimSpace(lines[0])
	// Check if it looks like SQL
	if strings.Contains(strings.ToLower(codeBlock), "select ") ||
	   strings.Contains(strings.ToLower(codeBlock), "insert ") ||
	   strings.Contains(strings.ToLower(codeBlock), "update ") {
		return "sql"
	}

	return firstLine
}

// splitTextIntoSections splits text content into logical sections
func splitTextIntoSections(text string, startNum int) []ResponseSection {
	sections := []ResponseSection{}

	// Split by headers (## or **)
	headerRegex := regexp.MustCompile(`(?m)^#{1,3}\s+(.+)$`)
	boldHeaderRegex := regexp.MustCompile(`(?m)^\*\*(.+?)\*\*:?\s*$`)

	lines := strings.Split(text, "\n")
	currentSection := []string{}
	currentTitle := ""
	sectionNum := startNum

	for _, line := range lines {
		// Check if this is a header
		if headerMatch := headerRegex.FindStringSubmatch(line); headerMatch != nil {
			// Save previous section
			if len(currentSection) > 0 {
				content := strings.TrimSpace(strings.Join(currentSection, "\n"))
				if content != "" {
					sections = append(sections, ResponseSection{
						Type:    inferSectionType(content),
						Content: content,
						Number:  sectionNum,
						Title:   currentTitle,
					})
					sectionNum++
				}
			}

			// Start new section
			currentTitle = headerMatch[1]
			currentSection = []string{}
		} else if boldMatch := boldHeaderRegex.FindStringSubmatch(line); boldMatch != nil {
			// Bold header
			if len(currentSection) > 0 {
				content := strings.TrimSpace(strings.Join(currentSection, "\n"))
				if content != "" {
					sections = append(sections, ResponseSection{
						Type:    inferSectionType(content),
						Content: content,
						Number:  sectionNum,
						Title:   currentTitle,
					})
					sectionNum++
				}
			}

			currentTitle = boldMatch[1]
			currentSection = []string{}
		} else {
			// Regular line
			currentSection = append(currentSection, line)
		}
	}

	// Save final section
	if len(currentSection) > 0 {
		content := strings.TrimSpace(strings.Join(currentSection, "\n"))
		if content != "" {
			sections = append(sections, ResponseSection{
				Type:    inferSectionType(content),
				Content: content,
				Number:  sectionNum,
				Title:   currentTitle,
			})
		}
	}

	// If no sections were created, create one big text section
	if len(sections) == 0 {
		sections = append(sections, ResponseSection{
			Type:    "text",
			Content: strings.TrimSpace(text),
			Number:  startNum,
			Title:   "Response",
		})
	}

	return sections
}

// inferSectionType attempts to infer the type of a text section
func inferSectionType(content string) string {
	// Check if it's a list (starts with -, *, or numbers)
	listRegex := regexp.MustCompile(`(?m)^\s*[-*]\s+|^\s*\d+\.\s+`)
	if listRegex.MatchString(content) {
		return "list"
	}

	// Check if it looks like a SQL query (but not in code block)
	sqlKeywords := []string{"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "ALTER", "DROP"}
	upperContent := strings.ToUpper(content)
	for _, keyword := range sqlKeywords {
		if strings.Contains(upperContent, keyword) {
			return "query"
		}
	}

	return "text"
}

// FormatSectionTitle formats a section title with its number
func FormatSectionTitle(section ResponseSection) string {
	if section.Title != "" {
		return section.Title
	}

	// Generate title based on type
	switch section.Type {
	case "code":
		return "Code Block"
	case "query":
		return "SQL Query"
	case "list":
		return "List"
	default:
		return "Text"
	}
}
