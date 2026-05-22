package parser

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrEmptyInput   = errors.New("no files provided")
	ErrTooManyFiles = errors.New("too many files")
	ErrTooLarge     = errors.New("input exceeds size limit")
	ErrNotSupport   = errors.New("language is not supported")
)

type File struct {
	Name    string
	Content string
}

type ParsedInput struct {
	Files     []File
	Language  string
	Framework string
	TotalSize int
}

type Service struct{}

func New() *Service { return &Service{} }

type ParseOptions struct {
	HintLanguage    string
	MaxFiles        int
	MaxRequestBytes int64
}

func (s *Service) Parse(files []File, opts ParseOptions) (*ParsedInput, error) {
	if len(files) == 0 {
		return nil, ErrEmptyInput
	}
	if opts.MaxFiles > 0 && len(files) > opts.MaxFiles {
		return nil, fmt.Errorf("%w: %d > %d", ErrTooManyFiles, len(files), opts.MaxFiles)
	}

	var totalSize int64
	for _, f := range files {
		totalSize += int64(len(f.Content))
	}
	if opts.MaxRequestBytes > 0 && totalSize > opts.MaxRequestBytes {
		return nil, fmt.Errorf("%w: %d bytes > %d bytes", ErrTooLarge, totalSize, opts.MaxRequestBytes)
	}

	relevant := FilterRelevant(files)
	if len(relevant) == 0 {
		relevant = files
	}

	lang := strings.ToLower(strings.TrimSpace(opts.HintLanguage))
	if lang == "" || lang == "auto" {
		lang = DetectLanguage(relevant)
	} else {
		switch lang {
		case LangGo, LangPython, LangTypeScript, LangJavaScript:
		default:
			return nil, fmt.Errorf("%w: %s", ErrNotSupport, lang)
		}
	}
	framework := DetectFramework(lang, relevant)

	return &ParsedInput{
		Files:     relevant,
		Language:  lang,
		Framework: framework,
		TotalSize: int(totalSize),
	}, nil
}
