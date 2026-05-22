package parser

import "testing"

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name  string
		files []File
		want  string
	}{
		{
			name:  "go files",
			files: []File{{Name: "main.go", Content: "package main"}, {Name: "router.go", Content: "package main"}},
			want:  LangGo,
		},
		{
			name:  "python files",
			files: []File{{Name: "app.py", Content: "from fastapi import FastAPI"}},
			want:  LangPython,
		},
		{
			name:  "javascript",
			files: []File{{Name: "server.js", Content: "const x = 1"}},
			want:  LangJavaScript,
		},
		{
			name:  "typescript",
			files: []File{{Name: "server.ts", Content: "const x: number = 1"}},
			want:  LangTypeScript,
		},
		{
			name:  "unknown",
			files: []File{{Name: "README.md", Content: "# hi"}},
			want:  LangUnknown,
		},
		{
			name: "majority wins",
			files: []File{
				{Name: "a.go", Content: ""},
				{Name: "b.go", Content: ""},
				{Name: "c.py", Content: ""},
			},
			want: LangGo,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectLanguage(tt.files)
			if got != tt.want {
				t.Fatalf("DetectLanguage = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectFramework(t *testing.T) {
	tests := []struct {
		name  string
		lang  string
		files []File
		want  string
	}{
		{
			name:  "gin via import",
			lang:  LangGo,
			files: []File{{Name: "main.go", Content: `import "github.com/gin-gonic/gin"`}},
			want:  FrameworkGin,
		},
		{
			name:  "fastapi via import",
			lang:  LangPython,
			files: []File{{Name: "app.py", Content: "from fastapi import FastAPI"}},
			want:  FrameworkFastAPI,
		},
		{
			name:  "flask via Flask(__name__)",
			lang:  LangPython,
			files: []File{{Name: "app.py", Content: "from flask import Flask\napp = Flask(__name__)"}},
			want:  FrameworkFlask,
		},
		{
			name:  "express via require",
			lang:  LangJavaScript,
			files: []File{{Name: "server.js", Content: `const express = require('express'); const app = express();`}},
			want:  FrameworkExpress,
		},
		{
			name:  "express via es-module import",
			lang:  LangTypeScript,
			files: []File{{Name: "server.ts", Content: `import express from "express";`}},
			want:  FrameworkExpress,
		},
		{
			name:  "unknown",
			lang:  LangGo,
			files: []File{{Name: "main.go", Content: "package main"}},
			want:  FrameworkUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectFramework(tt.lang, tt.files)
			if got != tt.want {
				t.Fatalf("DetectFramework = %q, want %q", got, tt.want)
			}
		})
	}
}
