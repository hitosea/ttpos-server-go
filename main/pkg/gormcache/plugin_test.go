package gormcache

import (
	"testing"
)

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"User", "user"},
		{"UserProfile", "user_profile"},
		{"HTTPServer", "h_t_t_p_server"},
		{"ID", "i_d"},
		{"", ""},
		{"a", "a"},
		{"ABC", "a_b_c"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValueToString(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"nil", nil, "nil"},
		{"int", 42, "42"},
		{"string", "hello", "hello"},
		{"bool", true, "true"},
		{"slice", []int{1, 2, 3}, "[1,2,3]"},
		{"empty slice", []int{}, "[]"},
		{"map", map[string]int{"a": 1}, "{a:1}"},
		{"empty map", map[string]int{}, "{}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valueToString(tt.input)
			if result != tt.expected {
				t.Errorf("valueToString(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValuesToString(t *testing.T) {
	tests := []struct {
		name     string
		input    []any
		expected string
	}{
		{"empty", []any{}, ""},
		{"single", []any{1}, "1"},
		{"multiple", []any{1, "hello", true}, "1,hello,true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valuesToString(tt.input)
			if result != tt.expected {
				t.Errorf("valuesToString(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	conf := DefaultConfig()

	if !conf.Easer {
		t.Error("Default Easer should be true")
	}

	if conf.TTL != 5*60*1000000000 { // 5 minutes in nanoseconds
		t.Errorf("Default TTL should be 5 minutes, got %v", conf.TTL)
	}

	if conf.MaxRows != 1000 {
		t.Errorf("Default MaxRows should be 1000, got %d", conf.MaxRows)
	}

	if conf.InvalidateOnWrite == nil || !*conf.InvalidateOnWrite {
		t.Error("Default InvalidateOnWrite should be true")
	}
}

func TestPluginName(t *testing.T) {
	plugin := New(nil)
	if plugin.Name() != "gorm:table-cache" {
		t.Errorf("Plugin name should be 'gorm:table-cache', got %q", plugin.Name())
	}
}

func TestShouldCacheTable(t *testing.T) {
	// 测试白名单模式
	plugin := New(&Config{
		Tables: []string{"users", "products"},
	})

	if !plugin.shouldCacheTable("users") {
		t.Error("users should be cached (in whitelist)")
	}

	if plugin.shouldCacheTable("orders") {
		t.Error("orders should not be cached (not in whitelist)")
	}

	// 测试黑名单模式
	plugin2 := New(&Config{
		ExcludeTables: []string{"orders", "logs"},
	})

	if !plugin2.shouldCacheTable("users") {
		t.Error("users should be cached (not in blacklist)")
	}

	if plugin2.shouldCacheTable("orders") {
		t.Error("orders should not be cached (in blacklist)")
	}
}

func TestConfigOptions(t *testing.T) {
	conf := DefaultConfig()

	WithEaser(false)(conf)
	if conf.Easer {
		t.Error("WithEaser(false) should set Easer to false")
	}

	WithTTL(10 * 60 * 1000000000)(conf) // 10 minutes
	if conf.TTL != 10*60*1000000000 {
		t.Error("WithTTL should set TTL")
	}

	WithMaxRows(500)(conf)
	if conf.MaxRows != 500 {
		t.Error("WithMaxRows should set MaxRows")
	}

	WithTables("a", "b", "c")(conf)
	if len(conf.Tables) != 3 {
		t.Error("WithTables should set Tables")
	}

	WithExcludeTables("x", "y")(conf)
	if len(conf.ExcludeTables) != 2 {
		t.Error("WithExcludeTables should set ExcludeTables")
	}

	WithDebug(true)(conf)
	if !conf.Debug {
		t.Error("WithDebug(true) should set Debug to true")
	}
}
