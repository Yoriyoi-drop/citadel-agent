// backend/go.test.sh
#!/bin/bash

# Script untuk menjalankan semua test Go dengan coverage

set -e

echo "Menjalankan Go tests..."

# Buat direktori untuk hasil coverage
mkdir -p coverage

# Jalankan tests dengan coverage
go test -v -race -coverprofile=coverage/coverage.out -covermode=atomic ./...

# Konversi coverage ke format fungsi
go tool cover -func=coverage/coverage.out -o coverage/coverage.func

# Tampilkan ringkasan coverage
echo "Coverage summary:"
go tool cover -func=coverage/coverage.out | grep total

# Jika coverage kurang dari 70%, keluarkan error (opsional)
MIN_COVERAGE=70
COVERAGE=$(go tool cover -func=coverage/coverage.out | grep total | grep -oE '[0-9]+\.[0-9]+')
if (( $(echo "$COVERAGE < $MIN_COVERAGE" | bc -l) )); then
    echo "Coverage $COVERAGE% is below minimum required $MIN_COVERAGE%"
    exit 1
fi

echo "Tests selesai. Lihat hasil coverage di coverage/coverage.out"

# Opsi: buat coverage report dalam format HTML
go tool cover -html=coverage/coverage.out -o coverage/coverage.html