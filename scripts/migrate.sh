#!/bin/bash

# Controlla se un argomento è stato passato
if [ $# -lt 1 ]; then
    echo "Errore: Specifica il percorso del file SQL come argomento."
    echo "Esempio: ./run_migrations.sh ../db/migrations/0001_create_initial_tables.sql"
    exit 1
fi

# Percorso al file SQL passato come primo argomento
SQL_FILE="$1"

# Controlla se il file SQL esiste
if [ ! -f "$SQL_FILE" ]; then
    echo "Errore: Il file $SQL_FILE non esiste."
    exit 1
fi

# Percorso al file .env
ENV_FILE="../.env"

# Controlla se il file .env esiste
if [ ! -f "$ENV_FILE" ]; then
    echo "Errore: Il file $ENV_FILE non esiste."
    exit 1
fi

# Carica le variabili d'ambiente dal file .env
export $(grep -v '^#' "$ENV_FILE" | xargs)

# Variabili per il contenitore Docker
CONTAINER_NAME="mysql" 

# Controlla che le variabili d'ambiente necessarie siano definite
if [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ] || [ -z "$DB_NAME" ]; then
    echo "Errore: Le variabili DB_USER, DB_PASSWORD o DB_NAME non sono definite nel file .env."
    exit 1
fi

# Copia il file SQL nel contenitore
BASENAME=$(basename "$SQL_FILE")
docker cp "$SQL_FILE" "$CONTAINER_NAME:/tmp/$BASENAME"

# Esegui il file SQL nel contenitore
docker exec -i "$CONTAINER_NAME" sh -c "mysql -u$DB_USER -p$DB_PASSWORD $DB_NAME < /tmp/$BASENAME"

# Controllo dell'esito
if [ $? -eq 0 ]; then
    echo "Le tabelle sono state create con successo."
else
    echo "Errore durante l'esecuzione del file SQL."
fi
