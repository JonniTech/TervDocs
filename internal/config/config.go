package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
)

const DefaultConfigFile = ".tervdocs.toml"

type OutputConfig struct {
	File   string `mapstructure:"file" toml:"file"`
	Backup bool   `mapstructure:"backup" toml:"backup"`
}

type ScanConfig struct {
	Include         []string `mapstructure:"include" toml:"include"`
	Exclude         []string `mapstructure:"exclude" toml:"exclude"`
	MaxFiles        int      `mapstructure:"max_files" toml:"max_files"`
	MaxBytesPerFile int      `mapstructure:"max_bytes_per_file" toml:"max_bytes_per_file"`
}

type ProviderConfig struct {
	APIKey   string `mapstructure:"api_key" toml:"api_key"`
	Model    string `mapstructure:"model" toml:"model"`
	BaseURL  string `mapstructure:"base_url" toml:"base_url"`
	Endpoint string `mapstructure:"endpoint" toml:"endpoint"`
}

type ProviderGroup struct {
	Free   ProviderConfig `mapstructure:"free" toml:"free"`
	OpenAI ProviderConfig `mapstructure:"openai" toml:"openai"`
	Gemini ProviderConfig `mapstructure:"gemini" toml:"gemini"`
	Claude ProviderConfig `mapstructure:"claude" toml:"claude"`
}

type Config struct {
	Provider      string        `mapstructure:"provider" toml:"provider"`
	Model         string        `mapstructure:"model" toml:"model"`
	Template      string        `mapstructure:"template" toml:"template"`
	DeveloperName string        `mapstructure:"developer_name" toml:"developer_name"`
	Temperature   float64       `mapstructure:"temperature" toml:"temperature"`
	TimeoutSec    int           `mapstructure:"timeout" toml:"timeout"`
	Output        OutputConfig  `mapstructure:"output" toml:"output"`
	Scan          ScanConfig    `mapstructure:"scan" toml:"scan"`
	Providers     ProviderGroup `mapstructure:"providers" toml:"providers"`
}

var ErrConfigNotFound = errors.New("config file not found")

func Default() Config {
	return Config{
		Provider:      "free",
		Model:         "glm-4.7-flash",
		Template:      "default",
		DeveloperName: "",
		Temperature:   0.2,
		TimeoutSec:    60,
		Output: OutputConfig{
			File:   "README.md",
			Backup: true,
		},
		Scan: ScanConfig{
			Include: []string{},
			Exclude: []string{
				".git", "node_modules", "dist", "build", "vendor", "coverage", ".next", ".turbo", "bin", ".venv", ".idea", ".vscode", ".qoder", ".codex",
			},
			MaxFiles:        200,
			MaxBytesPerFile: 50_000,
		},
		Providers: ProviderGroup{
			Free: ProviderConfig{
				APIKey:  "9ad8d75fa4774e3e83262745bc7087eb.z6LchituVrjHGc6N",
				Model:   "glm-4.7-flash",
				BaseURL: "https://api.z.ai/api/paas/v4",
			},
			OpenAI: ProviderConfig{
				Model:   "gpt-4o-mini",
				BaseURL: "https://api.openai.com/v1",
			},
			Gemini: ProviderConfig{
				Model:   "gemini-2.5-flash",
				BaseURL: "https://generativelanguage.googleapis.com/v1beta",
			},
			Claude: ProviderConfig{
				Model:   "claude-sonnet-4-0",
				BaseURL: "https://api.anthropic.com/v1",
			},
		},
	}
}

func ResolvePath(path string) string {
	if path != "" {
		return path
	}
	return DefaultConfigFile
}

func Load(path string) (Config, error) {
	cfg := Default()
	path = ResolvePath(path)

	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, ErrConfigNotFound
		}
		return cfg, err
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("toml")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return cfg, err
	}
	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, err
	}
	applyEnvKeys(&cfg)
	applyModelDefaults(&cfg)
	return cfg, nil
}

func Save(path string, cfg Config) error {
	path = ResolvePath(path)
	if err := Validate(cfg); err != nil {
		return err
	}
	b, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func Set(path, key, value string) error {
	cfg, err := Load(path)
	if err != nil && !errors.Is(err, ErrConfigNotFound) {
		return err
	}
	if errors.Is(err, ErrConfigNotFound) {
		cfg = Default()
	}

	if err := setField(&cfg, key, value); err != nil {
		return err
	}
	return Save(path, cfg)
}

func Show(path string) (string, error) {
	cfg, err := Load(path)
	if err != nil {
		return "", err
	}
	b, err := toml.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func Validate(cfg Config) error {
	if cfg.Provider == "" {
		return errors.New("provider is required")
	}
	switch cfg.Provider {
	case "free", "openai", "gemini", "claude":
	default:
		return fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
	if cfg.Template == "" {
		return errors.New("template is required")
	}
	if cfg.Output.File == "" {
		return errors.New("output.file is required")
	}
	if cfg.Scan.MaxFiles <= 0 {
		return errors.New("scan.max_files must be > 0")
	}
	if cfg.Scan.MaxBytesPerFile <= 0 {
		return errors.New("scan.max_bytes_per_file must be > 0")
	}
	if cfg.TimeoutSec <= 0 {
		return errors.New("timeout must be > 0")
	}
	return nil
}

func applyEnvKeys(cfg *Config) {
	if cfg.Providers.Free.APIKey == "" {
		cfg.Providers.Free.APIKey = os.Getenv("ZAI_API_KEY")
	}
	if cfg.Providers.Free.APIKey == "" {
		cfg.Providers.Free.APIKey = os.Getenv("Z_AI_API_KEY")
	}
	if cfg.Providers.OpenAI.APIKey == "" {
		cfg.Providers.OpenAI.APIKey = os.Getenv("OPENAI_API_KEY")
	}
	if cfg.Providers.Gemini.APIKey == "" {
		cfg.Providers.Gemini.APIKey = os.Getenv("GEMINI_API_KEY")
	}
	if cfg.Providers.Claude.APIKey == "" {
		cfg.Providers.Claude.APIKey = os.Getenv("ANTHROPIC_API_KEY")
	}
}

func applyModelDefaults(cfg *Config) {
	if cfg.Model != "" {
		return
	}
	switch cfg.Provider {
	case "free":
		cfg.Model = cfg.Providers.Free.Model
	case "openai":
		cfg.Model = cfg.Providers.OpenAI.Model
	case "gemini":
		cfg.Model = cfg.Providers.Gemini.Model
	case "claude":
		cfg.Model = cfg.Providers.Claude.Model
	}
}

func setField(cfg *Config, key, value string) error {
	switch key {
	case "provider":
		cfg.Provider = value
	case "model":
		cfg.Model = value
	case "template":
		cfg.Template = value
	case "developer_name":
		cfg.DeveloperName = value
	case "output.file":
		cfg.Output.File = value
	case "output.backup":
		cfg.Output.Backup = value == "true"
	case "scan.max_files":
		var n int
		if _, err := fmt.Sscanf(value, "%d", &n); err != nil {
			return errors.New("scan.max_files must be an integer")
		}
		cfg.Scan.MaxFiles = n
	case "scan.max_bytes_per_file":
		var n int
		if _, err := fmt.Sscanf(value, "%d", &n); err != nil {
			return errors.New("scan.max_bytes_per_file must be an integer")
		}
		cfg.Scan.MaxBytesPerFile = n
	case "providers.openai.api_key":
		cfg.Providers.OpenAI.APIKey = value
	case "providers.free.api_key":
		cfg.Providers.Free.APIKey = value
	case "providers.free.base_url":
		cfg.Providers.Free.BaseURL = value
	case "providers.gemini.api_key":
		cfg.Providers.Gemini.APIKey = value
	case "providers.claude.api_key":
		cfg.Providers.Claude.APIKey = value
	default:
		return fmt.Errorf("unsupported config key: %s", key)
	}
	return Validate(*cfg)
}

func OutputAbsPath(root, out string) string {
	if filepath.IsAbs(out) {
		return out
	}
	return filepath.Join(root, out)
}
