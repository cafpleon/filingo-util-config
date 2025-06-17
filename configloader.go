// configloader.go

// Package configloader provee una forma flexible y robusta de cargar
// configuraciones desde archivos y variables de entorno usando Viper.
package configloader

import (
	"context"
	"fmt"
	"strings"
	"sync"
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

// --- SINGLETON GLOBAL ---
var (
	// instance contendrá la única instancia de la configuración cargada.
	instance *Config
	// once asegura que la configuración se cargue una sola vez.
	once sync.Once
)

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
	Port           int32  `mapstructure:"port"`
	Version        string `mapstructure:"version"`
	ProjectRoot    string `mapstructure:"project_root"`
	GenerationRoot string `mapstructure:"generation_root"`
}

// DBConfig contiene la configuración de la base de datos.
type DBConfig struct {
	Driver            string        `mapstructure:"driver"`
	User              string        `mapstructure:"user"`
	Password          string        `mapstructure:"password"`
	Host              string        `mapstructure:"host"`
	Port              int32         `mapstructure:"port"`
	Name              string        `mapstructure:"name"`
	MaxConns          int32         `mapstructure:"max_connections"`
	MinConns          int32         `mapstructure:"min_connections"`
	MaxConnLifeTime   time.Duration `mapstructure:"max_connection_life_time"`
	MaxConnIdleTime   time.Duration `mapstructure:"max_connection_idle_time"`
	HealthCheckPeriod time.Duration `mapstructure:"health_check_period"`
}

// HTTPConfig contiene la configuración del servidor HTTP.
type HTTPConfig struct {
	Port           int32  `mapstructure:"port"`
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

// --- 3. FUNCIONES PÚBLICAS DE LA LIBRERÍA ---

// Init carga la configuración usando las opciones dadas y la almacena como un singleton.
// Debe ser llamada una sola vez al inicio de la aplicación. Es seguro llamarla múltiples veces.
func Init(opts Options) error {
	var err error
	once.Do(func() {
		// Llama a nuestra lógica de carga interna
		cfg, loadErr := load(opts)
		if loadErr != nil {
			err = loadErr
			return
		}
		instance = cfg
	})
	return err
}

// Get devuelve la instancia singleton de la configuración.
// Entrará en pánico si Init() no ha sido llamado exitosamente antes.
func Get() *Config {
	if instance == nil {
		panic("configloader: la configuración no ha sido inicializada. Llama a Init() primero.")
	}
	return instance
}

// configKey es un tipo privado para usar como clave en el contexto y evitar colisiones.
type configKey struct{}

// ToContext devuelve un nuevo contexto que contiene la configuración proporcionada.
func ToContext(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, configKey{}, cfg)
}

// FromContext extrae la configuración del contexto.
// Devuelve el puntero a la configuración y un booleano 'ok' que es 'true' si se encontró.
// Si no se encuentra, devuelve nil y false.
func FromContext(ctx context.Context) (*Config, bool) {
	cfg, ok := ctx.Value(configKey{}).(*Config)
	return cfg, ok
}

// --- LÓGICA DE CARGA INTERNA (NO PÚBLICA) ---

// load es la función interna que hace el trabajo pesado con Viper.
// Load busca, carga y decodifica la configuración en un struct Config.
// Devuelve un error si algo falla, permitiendo al programa principal manejarlo.
func load(opts Options) (*Config, error) {
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
