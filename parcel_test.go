package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Number:    randRange.Intn(10_000_000),
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	n, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, n)

	p, err := store.Get(n)
	require.NoError(t, err)
	require.Equal(t, p, parcel)

	err = store.Delete(n)
	require.NoError(t, err)
	_, err = store.Get(n)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	n, err := store.Add(parcel)
	require.NoError(t, err)

	newAddress := "new test address"
	err = store.SetAddress(n, newAddress)
	require.NoError(t, err)

	p, err := store.Get(n)
	require.NoError(t, err)
	assert.Equal(t, newAddress, p.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	n, err := store.Add(parcel)
	require.NoError(t, err)

	testStatus := ParcelStatusSent
	err = store.SetStatus(n, testStatus)
	require.NoError(t, err)

	p, err := store.Get(n)
	require.NoError(t, err)
	require.Equal(t, testStatus, p.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		if assert.NoError(t, err) {
			require.GreaterOrEqual(t, id, 0)
		}

		parcels[i].Number = id

		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	assert.Equal(t, len(storedParcels), len(parcels))

	for _, parcel := range storedParcels {
		id := parcel.Number

		assert.Contains(t, parcelMap, id)

		assert.Equal(t, parcelMap[id], parcel)
	}
}
