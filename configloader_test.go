// configloader_test.go
package configloader_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	configloader "github.com/cafpleon/filingo-util-config"
)

func TestLoad_Success(t *testing.T) {
	// --- ARRANGE (Organizar) ---

	// 1. Define el contenido de un archivo de configuración YAML de prueba como un string.
	// Usamos valores diferentes a los de producción para asegurar que estamos leyendo este archivo.
	yamlContent := `
application:
  name: "Mi App de Prueba"
  environment: "testing"
  port: "9090"

database:
  host: "db-test-host"
  port: "5433"
  user: "testuser"
  max_connections: 20
  max_connection_life_time: "15m"

google_oauth2:
  client_id: "client-id-de-prueba"
  session_secret: "secreto-de-prueba"

tokens:
  duration: "1h30m"
`
	// 2. Crea un directorio temporal que se limpiará automáticamente después del test.
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	// 3. Escribe nuestro contenido YAML en el archivo temporal.
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err, "Falló la creación del archivo de configuración temporal")

	// --- ACT (Actuar) ---

	// 4. Llama a la función Load, apuntando a nuestro archivo y directorio temporales.
	opts := configloader.Options{
		ConfigName:  "test-config", // Nombre del archivo sin extensión
		ConfigType:  "yaml",
		ConfigPaths: []string{tempDir}, // Le decimos que busque SOLO en nuestro directorio temporal
	}
	cfg, err := configloader.Load(opts)

	// --- ASSERT (Afirmar) ---

	// 5. Verifica que no hubo errores y que la configuración se cargó.
	require.NoError(t, err, "La función Load() no debería devolver un error")
	require.NotNil(t, cfg, "El struct de configuración no debería ser nulo")

	// 6. Verifica que los valores específicos se cargaron correctamente.
	// Esto prueba que el mapeo de `mapstructure` está funcionando.
	assert.Equal(t, "Mi App de Prueba", cfg.App.Name)
	assert.Equal(t, "testing", cfg.App.Environment)
	assert.Equal(t, "9090", cfg.App.Port)

	assert.Equal(t, "db-test-host", cfg.DB.Host)
	assert.Equal(t, int32(20), cfg.DB.MaxConns)

	// Verifica que Viper decodificó correctamente la duración del string.
	expectedDuration, _ := time.ParseDuration("15m")
	assert.Equal(t, expectedDuration, cfg.DB.MaxConnLifeTime)

	assert.Equal(t, "client-id-de-prueba", cfg.OAuth2.GoogleClientID)

	// Verifica la duración del token
	expectedTokenDuration, _ := time.ParseDuration("1h30m")
	assert.Equal(t, expectedTokenDuration, cfg.Token.Duration)
}

func TestLoad_FileNotFound_NoError(t *testing.T) {
	// Arrange: Opciones que apuntan a un archivo que no existe.
	opts := configloader.Options{
		ConfigName:  "archivo-que-no-existe",
		ConfigPaths: []string{t.TempDir()}, // Un directorio temporal vacío
	}

	// Act
	cfg, err := configloader.Load(opts)

	// Assert
	// La función no debe devolver error si el archivo no se encuentra,
	// ya que esto es un comportamiento esperado (puede usar solo variables de entorno).
	require.NoError(t, err)
	require.NotNil(t, cfg, "Incluso sin archivo, se debe devolver un struct de config vacío")
}
