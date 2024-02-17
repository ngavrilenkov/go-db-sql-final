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
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	n, err := store.Add(parcel)
	if assert.NoError(t, err) {
		require.GreaterOrEqual(t, n, 0)
	}

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	p, err := store.Get(n)
	if assert.NoError(t, err) {
		require.Equal(t, p.Number, n)
	}

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	err = store.Delete(n)
	if err != nil {
		require.NoError(t, err)
	}

	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(n)
	if err != nil {
		require.Error(t, err)
	}
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	n, err := store.Add(parcel)
	if assert.NoError(t, err) {
		require.GreaterOrEqual(t, n, 0)
	}

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(n, newAddress)
	if err != nil {
		require.NoError(t, err)
	}

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	p, err := store.Get(n)
	if assert.NoError(t, err) {
		assert.Equal(t, p.Address, newAddress)
	}
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	n, err := store.Add(parcel)
	if assert.NoError(t, err) {
		require.GreaterOrEqual(t, n, 0)
	}

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	testStatus := ParcelStatusSent
	err = store.SetStatus(n, testStatus)
	if err != nil {
		require.NoError(t, err)
	}

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	p, err := store.Get(n)
	if assert.NoError(t, err) {
		assert.Equal(t, p.Status, testStatus)
	}
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		if assert.NoError(t, err) {
			require.GreaterOrEqual(t, id, 0)
		}

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	// убедитесь в отсутствии ошибки
	if err != nil {
		require.NoError(t, err)
	}
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	assert.Equal(t, storedParcels, parcels)

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		id := parcel.Number

		// убедитесь, что все посылки из storedParcels есть в parcelMap
		assert.Contains(t, parcelMap, id)

		// убедитесь, что значения полей полученных посылок заполнены верно
		assert.Equal(t, parcelMap[id], parcel)
	}
}
