package main

import (
    "context"
    "encoding/csv"
    "fmt"
    "io"
    "log"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

func main() {
    connString := "postgres://user:password@localhost:5433/social_network"
    pool, err := pgxpool.New(context.Background(), connString)
    if err != nil {
	log.Fatalf("Не удалось подключиться к БД: %v", err)
    }
    defer pool.Close()

    fileName := "people.v2.csv"
    file, err := os.Open(fileName)
    if err != nil {
	log.Fatalf("Не удалось открыть файл %s: %v\nСначала скачайте файл: wget https://raw.githubusercontent.com/OtusTeam/highload/master/homework/people.v2.csv", fileName, err)
    }
    defer file.Close()

    reader := csv.NewReader(file)
    reader.Comma = ','
    reader.FieldsPerRecord = -1

    // Пропускаем заголовок
    if _, err := reader.Read(); err != nil {
	log.Fatalf("Ошибка чтения заголовка CSV: %v", err)
    }

    fmt.Println("Начало импорта данных...")
    start := time.Now()

    const batchSize = 10000
    batch := make([][]interface{}, 0, batchSize)

    totalInserted := 0
    lineNumber := 1

    processLine := func(record []string, idx int) {
	if len(record) < 3 {
	    return
	}

	nameParts := strings.SplitN(strings.TrimSpace(record[0]), " ", 2)
	if len(nameParts) < 2 {
	    return
	}
	lastName := strings.TrimSpace(nameParts[0])
	firstName := strings.TrimSpace(nameParts[1])

	birthDate := strings.TrimSpace(record[1])
	city := strings.TrimSpace(record[2])

	// Пустые значения для gender и interests
	gender := ""
	interests := ""

	login := strings.ToLower(firstName + "." + lastName + strconv.Itoa(idx))
	passwordHash := "$2a$10$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36WQoeG6Lruj3vjPGfa3ZlW"

	batch = append(batch, []interface{}{
	    login, firstName, lastName, birthDate, gender, interests, city, passwordHash,
	})
    }

    for {
	record, err := reader.Read()
	if err == io.EOF {
	    break
	}
	if err != nil {
	    continue
	}

	processLine(record, lineNumber)
	lineNumber++

	if len(batch) >= batchSize {
	    if err := insertBatch(context.Background(), pool, batch); err != nil {
		log.Printf("Ошибка вставки батча: %v", err)
	    }
	    totalInserted += len(batch)
	    batch = batch[:0]

	    if totalInserted%50000 == 0 {
		fmt.Printf("⏳ Вставлено записей: %d\n", totalInserted)
	    }
	}

	if totalInserted >= 1000000 {
	    break
	}
    }

    if len(batch) > 0 {
	if err := insertBatch(context.Background(), pool, batch); err != nil {
	    log.Printf("Ошибка вставки последнего батча: %v", err)
	}
	totalInserted += len(batch)
    }

    // Если в файле меньше 1 000 000 строк — дублируем данные
    if totalInserted < 1000000 {
	fmt.Printf("📊 В файле только %d строк. Генерируем дополнительные записи...\n", totalInserted)

	file.Seek(0, 0)
	reader = csv.NewReader(file)
	reader.Comma = ','
	reader.FieldsPerRecord = -1
	reader.Read()

	duplicateRound := 1
	for totalInserted < 1000000 {
	    record, err := reader.Read()
	    if err == io.EOF {
		file.Seek(0, 0)
		reader = csv.NewReader(file)
		reader.Comma = ','
		reader.FieldsPerRecord = -1
		reader.Read()
		duplicateRound++
		continue
	    }
	    if err != nil {
		continue
	    }

	    if len(record) < 3 {
		continue
	    }

	    nameParts := strings.SplitN(strings.TrimSpace(record[0]), " ", 2)
	    if len(nameParts) < 2 {
		continue
	    }
	    lastName := strings.TrimSpace(nameParts[0]) + strconv.Itoa(duplicateRound)
	    firstName := strings.TrimSpace(nameParts[1])
	    birthDate := strings.TrimSpace(record[1])
	    city := strings.TrimSpace(record[2])

	    gender := ""
	    interests := ""

	    login := strings.ToLower(firstName + "." + lastName + strconv.Itoa(totalInserted))
	    passwordHash := "$2a$10$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36WQoeG6Lruj3vjPGfa3ZlW"

	    batch = append(batch, []interface{}{
		login, firstName, lastName, birthDate, gender, interests, city, passwordHash,
	    })

	    if len(batch) >= batchSize {
		if err := insertBatch(context.Background(), pool, batch); err != nil {
		    log.Printf("Ошибка вставки батча: %v", err)
		}
		totalInserted += len(batch)
		batch = batch[:0]

		if totalInserted%50000 == 0 {
		    fmt.Printf("⏳ Вставлено записей: %d\n", totalInserted)
		}
	    }

	    if totalInserted >= 1000000 {
		break
	    }
	}

	if len(batch) > 0 {
	    if err := insertBatch(context.Background(), pool, batch); err != nil {
		log.Printf("Ошибка вставки последнего батча: %v", err)
	    }
	    totalInserted += len(batch)
	}
    }

    fmt.Printf("✅ Импорт завершен! Всего записей: %d. Время: %v\n", totalInserted, time.Since(start))
}

func insertBatch(ctx context.Context, pool *pgxpool.Pool, batch [][]interface{}) error {
    query := "INSERT INTO users (login, first_name, last_name, birth_date, gender, interests, city, password_hash) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
    _, err := pool.CopyFrom(
	ctx,
	pgx.Identifier{"users"},
	[]string{"login", "first_name", "last_name", "birth_date", "gender", "interests", "city", "password_hash"},
	pgx.CopyFromRows(batch),
    )
    _ = query // на случай, если понадобится для отладки
    return err
}