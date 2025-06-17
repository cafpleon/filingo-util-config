// configloader_test.go
package configloader

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitAndGet_Success(t *testing.T) {
	// --- CAMBIO 2: Añadimos una función de limpieza ---
	// t.Cleanup agenda esta función para que se ejecute AUTOMÁTICAMENTE
	// cuando este test termine. Así nos aseguramos de que el siguiente test
	// empiece con un estado limpio.
	t.Cleanup(func() {
		instance = nil
		once = sync.Once{}
	})

	// --- ARRANGE (Organizar) ---
	yamlContent := `
application:
  name: "Mi App de Prueba"
  environment: "testing"
  port: 9090
database:
  host: "db-test-host"
  max_connections: 20
  max_connection_life_time: "15m"
google_oauth2:
  client_id: "client-id-de-prueba"
`
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err, "Falló la creación del archivo de configuración temporal")

	// --- ACT (Actuar) ---
	// 4. Ahora llamamos a Init(), que es la función pública de nuestra librería.
	opts := Options{
		ConfigName:  "test-config",
		ConfigType:  "yaml",
		ConfigPaths: []string{tempDir},
	}
	initErr := Init(opts)

	// --- ASSERT (Afirmar) ---
	// 5. Verificamos que la inicialización fue exitosa.
	require.NoError(t, initErr, "Init() no debería devolver un error")

	// 6. Obtenemos la configuración usando Get() y verificamos que no sea nula.
	cfg := Get()
	require.NotNil(t, cfg, "Get() debería devolver un struct de configuración")

	// 7. Verificamos que los valores específicos se cargaron correctamente.
	assert.Equal(t, "Mi App de Prueba", cfg.App.Name)
	assert.Equal(t, "testing", cfg.App.Environment)
	assert.Equal(t, "db-test-host", cfg.DB.Host)
	assert.Equal(t, int32(20), cfg.DB.MaxConns)
	expectedDuration, _ := time.ParseDuration("15m")
	assert.Equal(t, expectedDuration, cfg.DB.MaxConnLifeTime)
	assert.Equal(t, "client-id-de-prueba", cfg.OAuth2.GoogleClientID)
}

func TestInit_ErrorOnMalformedFile(t *testing.T) {
	// Limpiamos el estado del singleton para este test también.
	t.Cleanup(func() {
		instance = nil
		once = sync.Once{}
	})

	// Arrange: Creamos un archivo YAML inválido.
	invalidYamlContent := `
application:
  name: "App Rota"
  port: 9090 : otrovalor # Sintaxis YAML incorrecta
`
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "bad-config.yaml")
	err := os.WriteFile(configPath, []byte(invalidYamlContent), 0644)
	require.NoError(t, err)

	// Act
	opts := Options{
		ConfigName:  "bad-config",
		ConfigType:  "yaml",
		ConfigPaths: []string{tempDir},
	}
	initErr := Init(opts)

	// Assert
	// Verificamos que Init() devuelve un error, como se esperaba.
	require.Error(t, initErr, "Init() debería devolver un error con un archivo malformado")
}

func TestGet_PanicsIfNotInitialized(t *testing.T) {
	// Limpiamos por si acaso algún test anterior falló antes de su cleanup.
	instance = nil
	once = sync.Once{}

	// Assert: Verificamos que llamar a Get() antes de Init() causa un pánico.
	// Esto confirma que nuestra guarda de seguridad funciona.
	assert.Panics(t, func() {
		Get()
	}, "Get() debería entrar en pánico si no se ha llamado a Init()")
}
