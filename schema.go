/*
antennaApp is a package for creating and curating blog, link blogs and social websites
Copyright (C) 2025 R. S. Doiel

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
*/
package antennaApp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// 3rd Party Package
	"gopkg.in/yaml.v3"
)

// AntennaApp configuration structure
type AppConfig struct {
	// Port holds the port number the preview web service will run on
	Port int `json:"port,omitempty" yaml:"port,omitempty"`

	// Host holds the IP address or machine name the preview service
	// will listen on. By default is is "localhost"
	Host string `json:"host,omitempty" yaml:"host,omitempty"`

	// Htdocs holds the path to directory that will recieve the generated content
	// It is also the directory used in the "preview" the static website.
	Htdocs string `json:"htdocs,omitempty" yaml:"htdocs,omitempty"`

	// UserAgent this holds a custom user agent string
	UserAgent string `json:"userAgent,omitempty" yaml:"userAgent,omitempty"`

	// BaseURL for the Antenna instance
	BaseURL string `json:"base_url,omitempty" yaml:"base_url,omitempty"`

	// Generator holds a YAML file that describes the HTML page structure.
	// This holds the default page generator description. Each collection can
	// use a custom one or the default one.
	Generator string `json:"generator,omitempty" yaml:"generator,omitempty"`

	// Collections holds a list of collections to curate
	Collections []*Collection `json:"collections,omitempty" yaml:"collections,omitempty"`

	// Sitemap settings, these should get sane defaults in the sitemap action
	ChunkSize   int
	DefaultFreq string
	DefaultPri  string
	FreqRules   map[string]string // outputPath prefix -> changefreq
	PriRules    map[string]string // outputPath prefix -> priority

}

// Collection describes the metadata about a collection of feeds.
// A collection can also be used to generate an RSS 2.0 feed of items
// harvested and related to the collection forming an aggregated item view
// of the collection of feeds.
//
// Some of the fields from the RSS 2.0 Channel can be set from the
// Markdown document's front matter.
//
// See https://cyber.harvard.edu/rss/rss.html#optionalChannelElements
type Collection struct {
	// Title of the collection
	Title string `json:"title,omitempty" yaml:"title,omitempty"`

	// Links holds the Link element used in the published RSS 2.0 output.
	Link string `json:"link,omitempty" yaml:"link,omitempty"`

	// Description holds the description of the collection
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// The language the collect is written in
	Language string `json:"language,omitempty" yaml:"language,omitempty"`

	// Copyright notice for content in the collection
	Copyright string `json:"copyright,omitempty" yaml:"copyright,omitempty"`

	// ManagingEditor holds an Email address for person responsible for editorial content
	// of the collection.
	ManagingEditor string `json:"managingEditor,omitempty" yaml:"managingEditor,omitempty"`

	// WebMaster holds an Email address for person responsible for technical issues relating to collection
	WebMaster string `json:"webMaster,omitempty" yaml:"webMaster,omitempty"`

	// PubDate holds the publication date for the content in the collection
	PubDate string `json:"pubDate,omitempty" yaml:"pubDate,omitempty"`

	// TTL is the time to live, the number of seconds to wait before trying a refresh
	TTL int `json:"ttl,omitempty" yaml:"ttl,omitempty"`

	// File holds the filepath to the Markdown document used to
	// define the collection.
	File string `json:"file,omitempty" yaml:"file,omitempty"`

	// Generator points to the YAML file to use when generating
	// a collection's HTML page.
	Generator string `json:"generator,omitempty" yaml:"generator,omitempty"`

	// Filters holds custom SQL that will be run against the Source to
	// determine which items to include and had off to the Generator.
	Filters []string `json:"filters,omitempty" yaml:"filters,omitempty"`

	// DbName holds the SQLite3 database filename
	DbName string `json:"dbName,omitempty" yaml:"dbName,omitempty"`
}

// Link represents a Markdown link with Label, URL, and optional Description.
type Link struct {
	// Label holds the text label that will be used when displaying the feed
	Label string `json:"label,omitempty" yaml:"label,omitempty"`
	// The URL holds the link text to the feed
	URL string `json:"url,omitempty" yaml:"url,omitempty"`
	// The optional description holds any description text associated with the link
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// LoadConfig process the AntennaApp YAML file and sets the
// attributes of the AntennaApp structure.
func (cfg *AppConfig) LoadConfig(cfgName string) error {
	src, err := os.ReadFile(cfgName)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(src, &cfg)
}

// SaveConfig save the current configuration of the AntennaApp
func (cfg *AppConfig) SaveConfig(cfgName string) error {
	src, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(cfgName, src, 0664); err != nil {
		return err
	}
	return nil
}

func (cfg *AppConfig) CollectionIndex(cName string) int {
	for i, col := range cfg.Collections {
		if filepath.Base(col.File) == filepath.Base(cName) {
			return i
		}
	}
	return -1
}

func (cfg *AppConfig) GetCollection(cName string) (*Collection, error) {
	i := cfg.CollectionIndex(cName)
	if i > -1 {
		return cfg.Collections[i], nil
	}
	return nil, fmt.Errorf("%s not in collection", cName)
}

func (collection *Collection) UpdateFrontMatter(frontMatter map[string]interface{}, cfg *AppConfig) error {
	rssFile := strings.TrimSuffix(collection.File, ".md") + ".xml"
	collection.Title = ""
	if title, ok := frontMatter["title"].(string); ok {
		collection.Title = title
	}
	if link, ok := frontMatter["link"].(string); ok {
		collection.Link = link
	} else if collection.Link == "" {
		if cfg.BaseURL != "" {
			collection.Link = fmt.Sprintf(`%s/%s`, cfg.BaseURL, rssFile)
		} else if cfg.Host != "" {
			collection.Link = fmt.Sprintf(`http://%s:%d/%s`, cfg.Host, cfg.Port, rssFile)
		} else {
			collection.Link = rssFile
		}

	}
	collection.Description = ""
	if description, ok := frontMatter["description"].(string); ok {
		collection.Description = description
	}
	collection.Language = ""
	if language, ok := frontMatter["language"].(string); ok {
		collection.Language = language
	}
	collection.Copyright = ""
	if copyright, ok := frontMatter["copyright"].(string); ok {
		collection.Copyright = copyright
	}
	collection.ManagingEditor = ""
	if managingEditor, ok := frontMatter["managingEditor"].(string); ok {
		collection.ManagingEditor = managingEditor
	}
	collection.WebMaster = ""
	if webMaster, ok := frontMatter["webMaster"].(string); ok {
		collection.WebMaster = webMaster
	}
	collection.PubDate = ""
	if pubDate, ok := frontMatter["pubDate"].(string); ok {
		collection.PubDate = pubDate
	}
	collection.TTL = 0
	if val, ok := frontMatter["ttl"].(int); ok {
		collection.TTL = val
	}
	collection.Generator = ""
	if generator, ok := frontMatter["generator"].(string); ok {
		collection.Generator = generator
	}
	collection.Filters = []string{}
	if filters, ok := frontMatter["filters"].([]string); ok {
		collection.Filters = append([]string{}, filters...)
	}
	collection.DbName = ""
	if dbName, ok := frontMatter["dbName"].(string); ok {
		collection.DbName = dbName
	}
	return nil
}
