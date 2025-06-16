// configloader.go

// Package configloader provee una forma flexible y robusta de cargar
// configuraciones desde archivos y variables de entorno usando Viper.
package configloader

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

/*
Funcionamiento:
EJ:
database:
  max_connections: 10
En viper.Unmarshal(&cfg), Viper ve el campo MaxConns,
   mira su tag mapstructure y busca una clave llamada max_connections dentro de la sección database de tu archivo YAML.
   Encuentra el valor 10 y lo asigna automáticamente al campo MaxConns.
*/

// --- ESTRUCTURAS DE CONFIGURACIÓN PÚBLICAS ---
// Todos los campos deben ser públicos (empezar con Mayúscula) para que Viper pueda llenarlos.
// Los tags `mapstructure` le dicen a Viper cómo mapear las claves del archivo YAML/JSON.

// Config es el struct principal que agrupa toda la configuración.
// Las claves aquí (application, database, etc.) DEBEN coincidir con las claves de nivel superior en el YAML.
type Config struct {
	App    AppConfig   `mapstructure:"application"`
	DB     DBConfig    `mapstructure:"database"`
	HTTP   HTTPConfig  `mapstructure:"http"`
	Redis  RedisConfig `mapstructure:"redis"`
	OAuth2 OAuthConfig `mapstructure:"google_oauth2"` // Coincide con la clave 'google_oauth2' en YAML
	Token  TokenConfig `mapstructure:"tokens"`        // Coincide con la clave 'tokens' en YAML
}

// AppConfig contiene la configuración de la aplicación.
type AppConfig struct {
	Name           string `mapstructure:"name"`
	Environment    string `mapstructure:"environment"`
	Port           string `mapstructure:"port"`
	Version        string `mapstructure:"version"`
	ProjectRoot    string `mapstructure:"project_root"`
	GenerationRoot string `mapstructure:"generation_root"`
}

// DBConfig contiene la configuración de la base de datos.
type DBConfig struct {
	User              string        `mapstructure:"user"`
	Password          string        `mapstructure:"password"`
	Host              string        `mapstructure:"host"`
	Port              string        `mapstructure:"port"`
	Name              string        `mapstructure:"name"`
	MaxConns          int32         `mapstructure:"max_connections"`
	MinConns          int32         `mapstructure:"min_connections"`
	MaxConnLifeTime   time.Duration `mapstructure:"max_connection_life_time"`
	MaxConnIdleTime   time.Duration `mapstructure:"max_connection_idle_time"`
	HealthCheckPeriod time.Duration `mapstructure:"health_check_period"`
}

// HTTPConfig contiene la configuración del servidor HTTP.
type HTTPConfig struct {
	Port           string `mapstructure:"port"`
	AllowedOrigins string `mapstructure:"allowed_origins"`
}

// RedisConfig contiene la configuración de Redis.
type RedisConfig struct {
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
}

// OAuthConfig contiene la configuración para OAuth2.
type OAuthConfig struct {
	GoogleClientID     string `mapstructure:"client_id"`
	GoogleClientSecret string `mapstructure:"client_secret"`
	GoogleRedirectURI  string `mapstructure:"redirect_uri"`
	SessionSecret      string `mapstructure:"session_secret"`
}

// TokenConfig contiene la configuración para la generación de tokens.
type TokenConfig struct {
	Duration time.Duration `mapstructure:"duration"`
}

// ---  OPCIONES DE CARGA ---

// Options permite al usuario de la librería personalizar el proceso de carga.
type Options struct {
	ConfigName  string   // ej: "config"
	ConfigType  string   // ej: "yaml", "json"
	ConfigPaths []string // ej: []string{".", "/etc/myapp"}
	EnvPrefix   string   // ej: "MYAPP"
}

// --- 3. FUNCIÓN DE CARGA PRINCIPAL ---

// Load busca, carga y decodifica la configuración en un struct Config.
// Devuelve un error si algo falla, permitiendo al programa principal manejarlo.
func Load(opts Options) (*Config, error) {
	v := viper.New()

	// Configurar Viper con las opciones proporcionadas por el usuario.
	v.SetConfigName(opts.ConfigName)
	v.SetConfigType(opts.ConfigType)
	for _, path := range opts.ConfigPaths {
		v.AddConfigPath(path)
	}

	// Configurar la lectura de variables de entorno.
	if opts.EnvPrefix != "" {
		v.SetEnvPrefix(opts.EnvPrefix)
	}
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Intentar leer el archivo de configuración (si existe).
	// No tratamos un archivo no encontrado como un error fatal.
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// El error es por otra cosa (ej: un archivo YAML malformado).
			return nil, fmt.Errorf("error al leer el archivo de configuración: %w", err)
		}
		// Si el archivo no se encuentra, no pasa nada.
	}

	// Decodificar (Unmarshal) toda la configuración en nuestro struct.
	// Esta es la "magia" que llena el struct automáticamente.
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error al decodificar la configuración: %w", err)
	}

	return &cfg, nil
}
