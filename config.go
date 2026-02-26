package fblog

// Config holds the configuration read from file/defaults.
// This matches the Rust struct Config from config.rs
type Config struct {
	MessageKeys           []string          `json:"message_keys" toml:"message_keys"`
	TimeKeys             []string          `json:"time_keys" toml:"time_keys"`
	DumpAllExclude       []string          `json:"dump_all_exclude" toml:"dump_all_exclude"`
	AlwaysPrintFields    []string          `json:"always_print_fields" toml:"always_print_fields"`
	LevelKeys            []string          `json:"level_keys" toml:"level_keys"`
	LevelMap             map[string]string `json:"level_map" toml:"level_map"`
	MainLineFormat       string            `json:"main_line_format" toml:"main_line_format"`
	AdditionalValueFormat string            `json:"additional_value_format" toml:"additional_value_format"`
}

// LogSettings holds configuration specifically for processing a stream.
// Matches log_settings.rs
type LogSettings struct {
	MessageKeys      []string
	TimeKeys         []string
	LevelKeys        []string
	LevelMap         map[string]string
	AdditionalValues []string
	ExcludedValues   []string
	DumpAll          bool
	WithPrefix       bool
	Substitution     *Substitution
}

// NewDefaultConfig returns a configuration with default values
func NewDefaultConfig() Config {
	return Config{
		MessageKeys:           []string{"short_message", "msg", "message"},
		TimeKeys:              []string{"timestamp", "time", "@timestamp"},
		DumpAllExclude:        []string{},
		AlwaysPrintFields:     []string{},
		LevelKeys:             []string{"level", "severity", "log.level", "loglevel"},
		LevelMap:              map[string]string{},
		MainLineFormat:        "{{bold (fixed_size 19 .fblog_timestamp)}} {{level_style (uppercase (fixed_size 5 .fblog_level))}}:{{if .fblog_prefix}} {{bold (cyan .fblog_prefix)}}{{end}} {{.fblog_message}}",
		AdditionalValueFormat: "{{bold (color_rgb 150 150 150 (min_size 25 .key))}}: {{.value}}",
	}
}

// NewLogSettingsFromConfig initializes LogSettings based on the Config
func NewLogSettingsFromConfig(cfg Config) LogSettings {
	// Need to copy maps/slices so they can be mutated safely if needed
	levelMap := make(map[string]string, len(cfg.LevelMap))
	for k, v := range cfg.LevelMap {
		levelMap[k] = v
	}

	return LogSettings{
		MessageKeys:      append([]string(nil), cfg.MessageKeys...),
		TimeKeys:         append([]string(nil), cfg.TimeKeys...),
		LevelKeys:        append([]string(nil), cfg.LevelKeys...),
		LevelMap:         levelMap,
		AdditionalValues: append([]string(nil), cfg.AlwaysPrintFields...),
		ExcludedValues:   append([]string(nil), cfg.DumpAllExclude...),
		DumpAll:          false,
		WithPrefix:       false,
		Substitution:     nil,
	}
}

// NewDefaultLogSettings initializes LogSettings using the default config
func NewDefaultLogSettings() LogSettings {
	return NewLogSettingsFromConfig(NewDefaultConfig())
}

func (s *LogSettings) AddAdditionalValues(values []string) {
	s.AdditionalValues = append(s.AdditionalValues, values...)
}

func (s *LogSettings) AddMessageKeys(keys []string) {
	s.MessageKeys = append(keys, s.MessageKeys...) // prepend
}

func (s *LogSettings) AddTimeKeys(keys []string) {
	s.TimeKeys = append(keys, s.TimeKeys...) // prepend
}

func (s *LogSettings) AddLevelKeys(keys []string) {
	s.LevelKeys = append(keys, s.LevelKeys...) // prepend
}

func (s *LogSettings) AddLevelMap(m map[string]string) {
	if s.LevelMap == nil {
		s.LevelMap = make(map[string]string)
	}
	for k, v := range m {
		s.LevelMap[k] = v
	}
}

func (s *LogSettings) AddExcludedValues(values []string) {
	s.ExcludedValues = append(s.ExcludedValues, values...)
}

func (s *LogSettings) SetSubstitution(sub *Substitution) {
	s.Substitution = sub
}
