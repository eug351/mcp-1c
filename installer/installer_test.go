package installer

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestBuildDesignerArgs(t *testing.T) {
	tests := []struct {
		name       string
		dbPath     string
		serverMode bool
		dbUser     string
		dbPassword string
		logPath    string
		extraArgs  []string
		want       []string
	}{
		{
			name:   "file mode without credentials",
			dbPath: `C:\MyBase`, logPath: "log.txt",
			extraArgs: []string{"/LoadConfigFromFiles", "/tmp/ext"},
			want: []string{
				"DESIGNER", "/F", `C:\MyBase`,
				"/LoadConfigFromFiles", "/tmp/ext",
				"/Out", "log.txt", "/DisableStartupDialogs", "/DisableStartupMessages",
			},
		},
		{
			name:       "file mode with credentials",
			dbPath:     `C:\MyBase`,
			dbUser:     "Admin",
			dbPassword: "pass",
			logPath:    "log.txt",
			extraArgs:  []string{"/LoadConfigFromFiles", "/tmp/ext"},
			want: []string{
				"DESIGNER", "/F", `C:\MyBase`,
				"/N", "Admin", "/P", "pass",
				"/LoadConfigFromFiles", "/tmp/ext",
				"/Out", "log.txt", "/DisableStartupDialogs", "/DisableStartupMessages",
			},
		},
		{
			name:       "server mode without credentials",
			dbPath:     `server01\accounting`,
			serverMode: true,
			logPath:    "log.txt",
			extraArgs:  []string{"/UpdateDBCfg"},
			want: []string{
				"DESIGNER", "/S", `server01\accounting`,
				"/UpdateDBCfg",
				"/Out", "log.txt", "/DisableStartupDialogs", "/DisableStartupMessages",
			},
		},
		{
			name:       "server mode with credentials",
			dbPath:     `server01\accounting`,
			serverMode: true,
			dbUser:     "Admin",
			dbPassword: "secret",
			logPath:    "log.txt",
			want: []string{
				"DESIGNER", "/S", `server01\accounting`,
				"/N", "Admin", "/P", "secret",
				"/Out", "log.txt", "/DisableStartupDialogs", "/DisableStartupMessages",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildDesignerArgs(tc.dbPath, tc.serverMode, tc.dbUser, tc.dbPassword, tc.logPath, tc.extraArgs...)
			if !slices.Equal(got, tc.want) {
				t.Errorf("mismatch\ngot:  %v\nwant: %v", got, tc.want)
			}
		})
	}
}

func TestPlatformPatterns(t *testing.T) {
	patterns := platformPatterns()
	if len(patterns) == 0 {
		t.Fatalf("expected non-empty patterns for GOOS=%s", runtime.GOOS)
	}
	t.Logf("GOOS=%s, patterns: %v", runtime.GOOS, patterns)
}

func TestExtractXMLTag(t *testing.T) {
	xml := `<Properties>
		<CompatibilityMode>Version8_3_24</CompatibilityMode>
		<InterfaceCompatibilityMode>TaxiEnableVersion8_2</InterfaceCompatibilityMode>
	</Properties>`

	if got := extractXMLTag(xml, "CompatibilityMode"); got != "Version8_3_24" {
		t.Errorf("CompatibilityMode = %q, want Version8_3_24", got)
	}
	if got := extractXMLTag(xml, "InterfaceCompatibilityMode"); got != "TaxiEnableVersion8_2" {
		t.Errorf("InterfaceCompatibilityMode = %q, want TaxiEnableVersion8_2", got)
	}
	if got := extractXMLTag(xml, "Missing"); got != "" {
		t.Errorf("Missing = %q, want empty", got)
	}
}

func TestPatchExtensionXML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Configuration.xml")

	original := `<Properties>
			<ConfigurationExtensionCompatibilityMode>Version8_3_14</ConfigurationExtensionCompatibilityMode>
			<DefaultRunMode>ManagedApplication</DefaultRunMode>
		</Properties>`

	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := patchExtensionXML(path, "Version8_3_24", "TaxiEnableVersion8_2"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "<ConfigurationExtensionCompatibilityMode>Version8_3_24</ConfigurationExtensionCompatibilityMode>") {
		t.Error("CompatibilityMode not patched")
	}
	if !strings.Contains(content, "<InterfaceCompatibilityMode>TaxiEnableVersion8_2</InterfaceCompatibilityMode>") {
		t.Error("InterfaceCompatibilityMode not inserted")
	}
}

func TestPatchExtensionXML_InsertBoth(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Configuration.xml")

	original := `<Properties>
			<DefaultRunMode>ManagedApplication</DefaultRunMode>
		</Properties>`

	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := patchExtensionXML(path, "Version8_3_20", "Taxi"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "Version8_3_20") {
		t.Error("CompatibilityMode not inserted")
	}
	if !strings.Contains(content, "<InterfaceCompatibilityMode>Taxi</InterfaceCompatibilityMode>") {
		t.Error("InterfaceCompatibilityMode not inserted")
	}
}

func TestReplaceOrInsertXMLTag(t *testing.T) {
	// Replace existing tag.
	content := `<Foo>old</Foo>`
	got := replaceOrInsertXMLTag(content, "Foo", "new")
	if !strings.Contains(got, "<Foo>new</Foo>") {
		t.Errorf("replace failed: %s", got)
	}

	// Insert missing tag.
	content = `<Properties>
		</Properties>`
	got = replaceOrInsertXMLTag(content, "Bar", "val")
	if !strings.Contains(got, "<Bar>val</Bar>") {
		t.Errorf("insert failed: %s", got)
	}
}

func TestFindPlatform(t *testing.T) {
	path, err := FindPlatform()
	if err != nil {
		t.Logf("1C not installed (expected on CI): %v", err)
		return
	}
	t.Logf("Found 1C at: %s", path)
}

func TestExtractPlatformMinor(t *testing.T) {
	tests := []struct {
		path      string
		wantMajor int
		wantMinor int
		wantOK    bool
	}{
		{`C:\Program Files\1cv8\8.3.27.1859\bin\1cv8.exe`, 3, 27, true},
		{`C:\Program Files\1cv8\8.3.14.2000\bin\1cv8.exe`, 3, 14, true},
		{`/opt/1cv8/x86_64/8.3.22.1709/1cv8`, 3, 22, true},
		{`/opt/1cv8/x86_64/8.5.1.100/1cv8`, 5, 1, true},
		{`/Applications/1cv8.localized/8.3.25.1000/1cv8.app/Contents/MacOS/1cv8`, 3, 25, true},
		{`/usr/bin/some-tool`, 0, 0, false},
		{``, 0, 0, false},
	}

	for _, tc := range tests {
		major, minor, ok := extractPlatformMinor(tc.path)
		if ok != tc.wantOK || major != tc.wantMajor || minor != tc.wantMinor {
			t.Errorf("extractPlatformMinor(%q) = (%d, %d, %v), want (%d, %d, %v)",
				tc.path, major, minor, ok, tc.wantMajor, tc.wantMinor, tc.wantOK)
		}
	}
}

func TestFormatVersionForPlatform(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		// Each platform version gets its correct format.
		{`C:\Program Files\1cv8\8.3.14.2000\bin\1cv8.exe`, "2.8"},
		{`C:\Program Files\1cv8\8.3.15.2000\bin\1cv8.exe`, "2.9"},
		{`C:\Program Files\1cv8\8.3.16.2000\bin\1cv8.exe`, "2.9.1"},
		{`C:\Program Files\1cv8\8.3.17.2000\bin\1cv8.exe`, "2.10"},
		{`C:\Program Files\1cv8\8.3.18.2000\bin\1cv8.exe`, "2.11"},
		{`C:\Program Files\1cv8\8.3.19.2000\bin\1cv8.exe`, "2.12"},
		{`C:\Program Files\1cv8\8.3.20.2000\bin\1cv8.exe`, "2.13"},
		{`C:\Program Files\1cv8\8.3.21.2000\bin\1cv8.exe`, "2.14"},
		{`C:\Program Files\1cv8\8.3.22.2000\bin\1cv8.exe`, "2.15"},
		{`C:\Program Files\1cv8\8.3.23.2000\bin\1cv8.exe`, "2.16"},
		{`C:\Program Files\1cv8\8.3.24.2000\bin\1cv8.exe`, "2.17"},
		{`C:\Program Files\1cv8\8.3.25.2000\bin\1cv8.exe`, "2.18"},
		{`C:\Program Files\1cv8\8.3.26.2000\bin\1cv8.exe`, "2.19"},
		{`C:\Program Files\1cv8\8.3.27.1859\bin\1cv8.exe`, "2.20"},
		// Platform 8.5 gets format 2.21.
		{`/opt/1cv8/x86_64/8.5.1.100/1cv8`, platform85FormatVersion},
		// Unknown path falls back to default.
		{`/usr/bin/unknown`, defaultFormatVersion},
		{``, defaultFormatVersion},
	}

	for _, tc := range tests {
		got := formatVersionForPlatform(tc.path)
		if got != tc.want {
			t.Errorf("formatVersionForPlatform(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

func TestPatchFormatVersion(t *testing.T) {
	dir := t.TempDir()

	// Create Configuration.xml with version="2.21".
	cfgXML := `<?xml version="1.0" encoding="UTF-8"?>
<MetaDataObject xmlns="http://v8.1c.ru/8.3/MDClasses" version="2.21">
	<Configuration uuid="test-uuid">
		<Properties>
			<Name>Test</Name>
		</Properties>
	</Configuration>
</MetaDataObject>`
	cfgPath := filepath.Join(dir, "Configuration.xml")
	if err := os.WriteFile(cfgPath, []byte(cfgXML), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create ConfigDumpInfo.xml with version="2.21".
	dumpXML := `<?xml version="1.0" encoding="UTF-8"?>
<ConfigDumpInfo xmlns="http://v8.1c.ru/8.3/xcf/dumpinfo" format="Hierarchical" version="2.21">
	<ConfigVersions/>
</ConfigDumpInfo>`
	dumpPath := filepath.Join(dir, "ConfigDumpInfo.xml")
	if err := os.WriteFile(dumpPath, []byte(dumpXML), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a subdirectory with another XML file.
	subDir := filepath.Join(dir, "HTTPServices")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	svcXML := `<?xml version="1.0" encoding="UTF-8"?>
<MetaDataObject xmlns="http://v8.1c.ru/8.3/MDClasses" version="2.21">
	<HTTPService uuid="svc-uuid"/>
</MetaDataObject>`
	svcPath := filepath.Join(subDir, "MCPService.xml")
	if err := os.WriteFile(svcPath, []byte(svcXML), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a non-XML file that should NOT be patched.
	txtPath := filepath.Join(dir, "readme.txt")
	txtContent := `version="2.21" should not be patched`
	if err := os.WriteFile(txtPath, []byte(txtContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Patch to version 2.18.
	if err := patchFormatVersion(dir, "2.18"); err != nil {
		t.Fatal(err)
	}

	// Verify Configuration.xml was patched.
	data, _ := os.ReadFile(cfgPath)
	if !strings.Contains(string(data), `version="2.18"`) {
		t.Errorf("Configuration.xml not patched:\n%s", data)
	}
	if strings.Contains(string(data), `version="2.21"`) {
		t.Errorf("Configuration.xml still contains old version:\n%s", data)
	}
	// Verify XML declaration was NOT touched.
	if !strings.Contains(string(data), `<?xml version="1.0"`) {
		t.Errorf("XML declaration version was incorrectly modified:\n%s", data)
	}

	// Verify ConfigDumpInfo.xml was patched.
	data, _ = os.ReadFile(dumpPath)
	if !strings.Contains(string(data), `version="2.18"`) {
		t.Errorf("ConfigDumpInfo.xml not patched:\n%s", data)
	}

	// Verify subdirectory XML was patched.
	data, _ = os.ReadFile(svcPath)
	if !strings.Contains(string(data), `version="2.18"`) {
		t.Errorf("HTTPServices/MCPService.xml not patched:\n%s", data)
	}

	// Verify non-XML file was NOT patched.
	data, _ = os.ReadFile(txtPath)
	if string(data) != txtContent {
		t.Errorf("non-XML file was modified: %s", data)
	}
}

func TestPatchFormatVersion_NoVersionAttr(t *testing.T) {
	dir := t.TempDir()

	// XML file without version= attribute should not be modified.
	original := `<?xml version="1.0" encoding="UTF-8"?>
<Root><Child>value</Child></Root>`
	path := filepath.Join(dir, "test.xml")
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := patchFormatVersion(dir, "2.16"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	if string(data) != original {
		t.Errorf("file without version attr was modified:\n%s", data)
	}
}

func TestPatchFormatVersion_AlreadyCorrect(t *testing.T) {
	dir := t.TempDir()

	// File already has the target version - should not be rewritten.
	original := `<?xml version="1.0" encoding="UTF-8"?>
<MetaDataObject version="2.16"><Data/></MetaDataObject>`
	path := filepath.Join(dir, "test.xml")
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := patchFormatVersion(dir, "2.16"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	if string(data) != original {
		t.Errorf("file with matching version was modified:\n%s", data)
	}
}
