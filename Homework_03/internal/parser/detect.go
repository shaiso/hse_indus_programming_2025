package parser

import (
	"path/filepath"
	"strings"
)

const (
	LangGo         = "go"
	LangPython     = "python"
	LangJavaScript = "javascript"
	LangTypeScript = "typescript"
	LangUnknown    = "unknown"
)

const (
	FrameworkGin     = "gin"
	FrameworkNetHTTP = "net/http"
	FrameworkChi     = "chi"
	FrameworkEcho    = "echo"
	FrameworkFiber   = "fiber"
	FrameworkFastAPI = "fastapi"
	FrameworkFlask   = "flask"
	FrameworkDjango  = "django"
	FrameworkExpress = "express"
	FrameworkKoa     = "koa"
	FrameworkNest    = "nest"
	FrameworkUnknown = ""
)

func DetectLanguage(files []File) string {
	counts := map[string]int{}
	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f.Name))
		switch ext {
		case ".go":
			counts[LangGo]++
		case ".py":
			counts[LangPython]++
		case ".js", ".mjs", ".cjs":
			counts[LangJavaScript]++
		case ".ts":
			counts[LangTypeScript]++
		}
	}
	if len(counts) == 0 {
		return LangUnknown
	}
	best, max := LangUnknown, 0
	for lang, c := range counts {
		if c > max {
			best, max = lang, c
		}
	}
	return best
}

type frameworkRule struct {
	framework string
	markers   []string
}

var frameworkRules = map[string][]frameworkRule{
	LangGo: {
		{FrameworkGin, []string{"github.com/gin-gonic/gin", "gin.Default(", "gin.New("}},
		{FrameworkChi, []string{"go-chi/chi"}},
		{FrameworkEcho, []string{"labstack/echo"}},
		{FrameworkFiber, []string{"gofiber/fiber"}},
		{FrameworkNetHTTP, []string{"net/http", "http.HandleFunc", "http.ListenAndServe"}},
	},
	LangPython: {
		{FrameworkFastAPI, []string{"from fastapi", "import fastapi", "FastAPI("}},
		{FrameworkFlask, []string{"from flask", "import flask", "Flask(__name__)"}},
		{FrameworkDjango, []string{"django.urls", "from django"}},
	},
	LangJavaScript: jsRules,
	LangTypeScript: jsRules,
}

var jsRules = []frameworkRule{
	{FrameworkExpress, []string{`require('express')`, `require("express")`, `from 'express'`, `from "express"`, "express()"}},
	{FrameworkKoa, []string{`require('koa')`, "new Koa()"}},
	{FrameworkNest, []string{"@nestjs/", "NestFactory.create"}},
}

func DetectFramework(language string, files []File) string {
	rules, ok := frameworkRules[language]
	if !ok {
		return FrameworkUnknown
	}
	hits := map[string]int{}
	for _, f := range files {
		for _, rule := range rules {
			if anyMarker(f.Content, rule.markers) {
				hits[rule.framework]++
			}
		}
	}
	best, top := FrameworkUnknown, 0
	for fw, c := range hits {
		if c > top {
			best, top = fw, c
		}
	}
	return best
}

func anyMarker(content string, markers []string) bool {
	for _, m := range markers {
		if strings.Contains(content, m) {
			return true
		}
	}
	return false
}
