package link

import (
	"errors"
	"reflect"
	"testing"
)

func TestExtract(t *testing.T) {
	tests := []struct {
		name      string
		rawText   string
		wantLinks []Link
		wantErr   bool
	}{
		{
			name:      "no links",
			rawText:   "just plain text",
			wantLinks: []Link{},
			wantErr:   false,
		},
		{
			name:    "single direct link",
			rawText: "[[Target]]",
			wantLinks: []Link{
				{
					TargetFile:     "Target",
					TargetHeading:  "",
					DisplayText:    "",
					IsTransclusion: false,
					LinkType:       Direct,
				},
			},
			wantErr: false,
		},
		{
			name:    "single transclusion link",
			rawText: "![[Target]]",
			wantLinks: []Link{
				{
					TargetFile:     "Target",
					TargetHeading:  "",
					DisplayText:    "",
					IsTransclusion: true,
					LinkType:       Direct,
				},
			},
			wantErr: false,
		},
		{
			name:    "link with display text",
			rawText: "[[Target|Display]]",
			wantLinks: []Link{
				{
					TargetFile:     "Target",
					TargetHeading:  "",
					DisplayText:    "Display",
					IsTransclusion: false,
					LinkType:       Direct,
				},
			},
			wantErr: false,
		},
		{
			name:    "link with heading",
			rawText: "[[Target#Section]]",
			wantLinks: []Link{
				{
					TargetFile:     "Target",
					TargetHeading:  "Section",
					DisplayText:    "",
					IsTransclusion: false,
					LinkType:       Direct,
				},
			},
			wantErr: false,
		},
		{
			name:    "link with file, heading and display",
			rawText: "[[Target#Section|Display]]",
			wantLinks: []Link{
				{
					TargetFile:     "Target",
					TargetHeading:  "Section",
					DisplayText:    "Display",
					IsTransclusion: false,
					LinkType:       Direct,
				},
			},
			wantErr: false,
		},
		{
			name:    "backslash path separator normalized in full extraction",
			rawText: "[[Target\\nDisplay]]",
			wantLinks: []Link{
				{
					TargetFile:     "Target/nDisplay",
					TargetHeading:  "",
					DisplayText:    "",
					IsTransclusion: false,
					LinkType:       Direct,
				},
			},
			wantErr: false,
		},
		{
			name:    "multiple links",
			rawText: "prefix [[A]] middle ![[B#H|D]] suffix [[C|X]]",
			wantLinks: []Link{
				{
					TargetFile:     "A",
					TargetHeading:  "",
					DisplayText:    "",
					IsTransclusion: false,
					LinkType:       Direct,
				},
				{
					TargetFile:     "B",
					TargetHeading:  "H",
					DisplayText:    "D",
					IsTransclusion: true,
					LinkType:       Direct,
				},
				{
					TargetFile:     "C",
					TargetHeading:  "",
					DisplayText:    "X",
					IsTransclusion: false,
					LinkType:       Direct,
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid link produces error but returns valid ones",
			rawText: "[[Valid]] [[Invalid",
			wantLinks: []Link{
				{
					TargetFile:     "Valid",
					TargetHeading:  "",
					DisplayText:    "",
					IsTransclusion: false,
					LinkType:       Direct,
				},
			},
			wantErr: true,
		},
		{
			name:      "empty target produces error",
			rawText:   "[[]]",
			wantLinks: []Link{},
			wantErr:   true,
		},
		{
			name:    "spaces are trimmed",
			rawText: "[[  Target  #  Section  |  Display  ]]",
			wantLinks: []Link{
				{
					TargetFile:     "Target",
					TargetHeading:  "Section",
					DisplayText:    "Display",
					IsTransclusion: false,
					LinkType:       Direct,
				},
			},
			wantErr: false,
		},

		{
			name:      "display only link returns error",
			rawText:   "[[|aaa]]",
			wantLinks: []Link{},
			wantErr:   true,
		},
		{
			name:    "unicode and multiple spaces in heading",
			rawText: "[[Note  Тщеу#Section 1Hub f]]",
			wantLinks: []Link{
				{
					TargetFile:     "Note  Тщеу",
					TargetHeading:  "Section 1Hub f",
					DisplayText:    "",
					IsTransclusion: false,
					LinkType:       Direct,
				},
			},
			wantErr: false,
		},
		{
			name:    "nested brackets parse as file with leading bracket",
			rawText: "[[[fsm]]]",
			wantLinks: []Link{
				{
					TargetFile:     "[fsm",
					TargetHeading:  "",
					DisplayText:    "",
					IsTransclusion: false,
					LinkType:       Direct,
				},
			},
			wantErr: false,
		},
		{
			name:    "newline inside link preserved",
			rawText: "[[Target\nDisplay]]",
			wantLinks: []Link{
				{
					TargetFile:     "Target\nDisplay",
					TargetHeading:  "",
					DisplayText:    "",
					IsTransclusion: false,
					LinkType:       Direct,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLinks, err := Extract(tt.rawText)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Extract() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(gotLinks, tt.wantLinks) {
				t.Errorf("Extract() links = %#v, want %#v", gotLinks, tt.wantLinks)
			}
		})
	}
}

func TestParseLink(t *testing.T) {
	tests := []struct {
		name    string
		rawLink string
		want    Link
		wantErr error
	}{
		{
			name:    "direct link without display or heading",
			rawLink: "[[Target]]",
			want: Link{
				TargetFile:     "Target",
				TargetHeading:  "",
				DisplayText:    "",
				IsTransclusion: false,
				LinkType:       Direct,
			},
			wantErr: nil,
		},
		{
			name:    "backslash path separator normalized",
			rawLink: "[[Target\\nDisplay]]",
			want: Link{
				TargetFile:     "Target/nDisplay",
				TargetHeading:  "",
				DisplayText:    "",
				IsTransclusion: false,
				LinkType:       Direct,
			},
			wantErr: nil,
		},
		{
			name:    "transclusion link",
			rawLink: "![[Target]]",
			want: Link{
				TargetFile:     "Target",
				TargetHeading:  "",
				DisplayText:    "",
				IsTransclusion: true,
				LinkType:       Direct,
			},
			wantErr: nil,
		},
		{
			name:    "with display text",
			rawLink: "[[Target|Display]]",
			want: Link{
				TargetFile:     "Target",
				TargetHeading:  "",
				DisplayText:    "Display",
				IsTransclusion: false,
				LinkType:       Direct,
			},
			wantErr: nil,
		},
		{
			name:    "with heading",
			rawLink: "[[Target#Section]]",
			want: Link{
				TargetFile:     "Target",
				TargetHeading:  "Section",
				DisplayText:    "",
				IsTransclusion: false,
				LinkType:       Direct,
			},
			wantErr: nil,
		},
		{
			name:    "with file, heading and display",
			rawLink: "[[Target#Section|Display]]",
			want: Link{
				TargetFile:     "Target",
				TargetHeading:  "Section",
				DisplayText:    "Display",
				IsTransclusion: false,
				LinkType:       Direct,
			},
			wantErr: nil,
		},
		{
			name:    "only heading (no file)",
			rawLink: "[[#Section]]",
			want: Link{
				TargetFile:     "",
				TargetHeading:  "Section",
				DisplayText:    "",
				IsTransclusion: false,
				LinkType:       Direct,
			},
			wantErr: nil,
		},
		{
			name:    "invalid format missing brackets",
			rawLink: "Target",
			want:    Link{},
			wantErr: ErrInvalidWikilinkFormat,
		},
		{
			name:    "invalid format only opening brackets",
			rawLink: "[[Target",
			want:    Link{},
			wantErr: ErrInvalidWikilinkFormat,
		},
		{
			name:    "empty target",
			rawLink: "[[]]",
			want:    Link{},
			wantErr: ErrEmptyLinkTarget,
		},
		{
			name:    "spaces trimmed",
			rawLink: "[[  Target  #  Section  |  Display  ]]",
			want: Link{
				TargetFile:     "Target",
				TargetHeading:  "Section",
				DisplayText:    "Display",
				IsTransclusion: false,
				LinkType:       Direct,
			},
			wantErr: nil,
		},

		{
			name:    "display only returns empty target error",
			rawLink: "[[|aaa]]",
			want:    Link{},
			wantErr: ErrEmptyLinkTarget,
		},
		{
			name:    "unicode and multiple spaces in heading",
			rawLink: "[[Note  Тщеу#Section 1Hub f]]",
			want: Link{
				TargetFile:     "Note  Тщеу",
				TargetHeading:  "Section 1Hub f",
				DisplayText:    "",
				IsTransclusion: false,
				LinkType:       Direct,
			},
			wantErr: nil,
		},
		{
			name:    "nested brackets parse as file with leading bracket",
			rawLink: "[[[fsm]]]",
			want: Link{
				TargetFile:     "[fsm",
				TargetHeading:  "",
				DisplayText:    "",
				IsTransclusion: false,
				LinkType:       Direct,
			},
			wantErr: nil,
		},
		{
			name:    "newline inside link preserved",
			rawLink: "[[Target\nDisplay]]",
			want: Link{
				TargetFile:     "Target\nDisplay",
				TargetHeading:  "",
				DisplayText:    "",
				IsTransclusion: false,
				LinkType:       Direct,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLink(tt.rawLink)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("parseLink() error = %v, want %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseLink() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestExtractFromText(t *testing.T) {
	tests := []struct {
		name    string
		rawText string
		want    []string
	}{
		{
			name:    "no links",
			rawText: "plain text",
			want:    nil,
		},
		{
			name:    "single link",
			rawText: "[[A]]",
			want:    []string{"[[A]]"},
		},
		{
			name:    "multiple links and transclusion",
			rawText: "before [[A]] middle ![[B]] after [[C#D|E]]",
			want:    []string{"[[A]]", "![[B]]", "[[C#D|E]]"},
		},
		{
			name:    "nested brackets are matched",
			rawText: "[[[fsm]]]",
			want:    []string{"[[[fsm]]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFromText(tt.rawText)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractFromText() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestSplitDisplay(t *testing.T) {
	tests := []struct {
		name        string
		inner       string
		wantTarget  string
		wantDisplay string
	}{
		{"no pipe", "Target", "Target", ""},
		{"with pipe", "Target|Display", "Target", "Display"},
		{"pipe at start", "|Display", "", "Display"},
		{"pipe at end", "Target|", "Target", ""},
		{"multiple pipes use first", "Target|Display|Extra", "Target", "Display|Extra"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, display := splitDisplay(tt.inner)
			if target != tt.wantTarget || display != tt.wantDisplay {
				t.Errorf("splitDisplay(%q) = (%q, %q), want (%q, %q)",
					tt.inner, target, display, tt.wantTarget, tt.wantDisplay)
			}
		})
	}
}

func TestSplitHeading(t *testing.T) {
	tests := []struct {
		name            string
		target          string
		wantFilePart    string
		wantHeadingPart string
	}{
		{"no hash", "Target", "Target", ""},
		{"with hash", "Target#Section", "Target", "Section"},
		{"hash at start", "#Section", "", "Section"},
		{"hash at end", "Target#", "Target", ""},
		{"multiple hashes use first", "Target#Section#Sub", "Target", "Section#Sub"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePart, headingPart := splitHeading(tt.target)
			if filePart != tt.wantFilePart || headingPart != tt.wantHeadingPart {
				t.Errorf("splitHeading(%q) = (%q, %q), want (%q, %q)",
					tt.target, filePart, headingPart, tt.wantFilePart, tt.wantHeadingPart)
			}
		})
	}
}

func TestExtractInnerLinkText(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantInner string
		wantOK    bool
	}{
		{"valid", "[[inner]]", "inner", true},
		{"missing opening", "inner]]", "", false},
		{"missing closing", "[[inner", "", false},
		{"both missing", "inner", "", false},
		{"empty inner", "[[]]", "", true},
		{"nested brackets parsed to inner with leading bracket", "[[[fsm]]", "[fsm", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInner, gotOK := extractInnerLinkText(tt.text)
			if gotInner != tt.wantInner || gotOK != tt.wantOK {
				t.Errorf("extractInnerLinkText(%q) = (%q, %v), want (%q, %v)",
					tt.text, gotInner, gotOK, tt.wantInner, tt.wantOK)
			}
		})
	}
}

func TestLinkString(t *testing.T) {
	tests := []struct {
		name string
		link Link
		want string
	}{
		{
			name: "full link with long ids",
			link: Link{
				SourceFile:      "source.md",
				SourceSegmentID: "1234567890abcdef",
				TargetFile:      "target.md",
				TargetSegmentID: "abcdef1234567890",
				TargetHeading:   "Section",
				DisplayText:     "Display",
				IsTransclusion:  true,
				LinkType:        Direct,
			},
			want: "Link{Type:direct, Source:source.md|12345678..., Target:target.md|abcdef12... #Section, Display:\"Display\", Transclusion:true}",
		},
		{
			name: "minimal link",
			link: Link{
				SourceFile: "src.md",
				TargetFile: "dst.md",
				LinkType:   BackLink,
			},
			want: "Link{Type:backlink, Source:src.md|, Target:dst.md|, Display:\"\", Transclusion:false}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.link.String()
			if got != tt.want {
				t.Errorf("Link.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncateID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{"short id unchanged", "abc", "abc"},
		{"exact length unchanged", "12345678", "12345678"},
		{"long id truncated", "123456789", "12345678..."},
		{"empty id", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateID(tt.id)
			if got != tt.want {
				t.Errorf("truncateID(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}
