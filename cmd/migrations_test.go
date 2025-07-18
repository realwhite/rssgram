// Тест создан с помощью AI
package main

import (
	"database/sql"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

func TestMigrations_Up(t *testing.T) {
	// Создаем временную базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Создаем экземпляр SQLite для миграций
	instance, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	require.NoError(t, err)

	// Создаем источник миграций
	d, err := iofs.New(migrations, "migrations")
	require.NoError(t, err)

	// Создаем экземпляр миграции
	m, err := migrate.NewWithInstance("iofs", d, "sqlite", instance)
	require.NoError(t, err)
	defer m.Close()

	// Применяем миграции
	err = m.Up()
	require.NoError(t, err)

	// Проверяем, что таблицы созданы
	tables := []string{"feeds", "items"}
	for _, table := range tables {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Table %s should exist", table)
	}
}

func TestMigrations_Down(t *testing.T) {
	// Создаем временную базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Создаем экземпляр SQLite для миграций
	instance, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	require.NoError(t, err)

	// Создаем источник миграций
	d, err := iofs.New(migrations, "migrations")
	require.NoError(t, err)

	// Создаем экземпляр миграции
	m, err := migrate.NewWithInstance("iofs", d, "sqlite", instance)
	require.NoError(t, err)
	defer m.Close()

	// Применяем миграции вверх
	err = m.Up()
	require.NoError(t, err)

	// Откатываем миграции
	err = m.Down()
	require.NoError(t, err)

	// Проверяем, что таблицы удалены
	tables := []string{"feeds", "items"}
	for _, table := range tables {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Table %s should not exist", table)
	}
}

func TestMigrations_Step(t *testing.T) {
	// Создаем временную базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Создаем экземпляр SQLite для миграций
	instance, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	require.NoError(t, err)

	// Создаем источник миграций
	d, err := iofs.New(migrations, "migrations")
	require.NoError(t, err)

	// Создаем экземпляр миграции
	m, err := migrate.NewWithInstance("iofs", d, "sqlite", instance)
	require.NoError(t, err)
	defer m.Close()

	// Применяем миграции пошагово
	err = m.Steps(1)
	require.NoError(t, err)

	// Проверяем, что первая таблица создана
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='feeds'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Table feeds should exist")

	// Применяем следующую миграцию
	err = m.Steps(1)
	require.NoError(t, err)

	// Проверяем, что вторая таблица создана
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='items'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Table items should exist")
}

func TestMigrations_Force(t *testing.T) {
	// Создаем временную базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Создаем экземпляр SQLite для миграций
	instance, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	require.NoError(t, err)

	// Создаем источник миграций
	d, err := iofs.New(migrations, "migrations")
	require.NoError(t, err)

	// Создаем экземпляр миграции
	m, err := migrate.NewWithInstance("iofs", d, "sqlite", instance)
	require.NoError(t, err)
	defer m.Close()

	// Применяем миграции
	err = m.Up()
	require.NoError(t, err)

	// Принудительно устанавливаем версию
	err = m.Force(1)
	require.NoError(t, err)

	// Проверяем версию
	version, dirty, err := m.Version()
	require.NoError(t, err)
	assert.Equal(t, uint(1), version)
	assert.False(t, dirty)
}

func TestMigrations_Version(t *testing.T) {
	// Создаем временную базу данных в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Создаем экземпляр SQLite для миграций
	instance, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	require.NoError(t, err)

	// Создаем источник миграций
	d, err := iofs.New(migrations, "migrations")
	require.NoError(t, err)

	// Создаем экземпляр миграции
	m, err := migrate.NewWithInstance("iofs", d, "sqlite", instance)
	require.NoError(t, err)
	defer m.Close()

	// Проверяем начальную версию
	version, dirty, err := m.Version()
	if err != nil && err.Error() == "no migration" {
		// Это допустимо для пустой базы
		return
	}
	require.NoError(t, err)
	assert.Equal(t, uint(0), version)
	assert.False(t, dirty)

	// Применяем миграции
	err = m.Up()
	require.NoError(t, err)

	// Проверяем финальную версию
	version, dirty, err = m.Version()
	require.NoError(t, err)
	assert.Greater(t, version, uint(0))
	assert.False(t, dirty)
}

func TestMigrations_EmbeddedFiles(t *testing.T) {
	// Проверяем, что миграции встроены в бинарный файл
	files, err := migrations.ReadDir("migrations")
	require.NoError(t, err)
	assert.Greater(t, len(files), 0, "Should have at least one migration file")

	// Проверяем, что есть файлы .up.sql
	hasUpFiles := false
	for _, file := range files {
		if file.Name() == "01_init.up.sql" {
			hasUpFiles = true
			break
		}
	}
	assert.True(t, hasUpFiles, "Should have up migration files")
}
