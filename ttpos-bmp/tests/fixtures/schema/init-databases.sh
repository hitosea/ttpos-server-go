#!/bin/bash
# Initialize BMP test databases and apply all migrations in timestamp order.
# Executed by MySQL Docker entrypoint on first startup.

set -e

mysql -u root -p"${MYSQL_ROOT_PASSWORD}" <<-EOSQL
  CREATE DATABASE IF NOT EXISTS erp       CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
  CREATE DATABASE IF NOT EXISTS takeout   CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
  CREATE DATABASE IF NOT EXISTS message   CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
  CREATE DATABASE IF NOT EXISTS websocket CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

  GRANT ALL PRIVILEGES ON erp.*       TO '${MYSQL_USER}'@'%';
  GRANT ALL PRIVILEGES ON takeout.*   TO '${MYSQL_USER}'@'%';
  GRANT ALL PRIVILEGES ON message.*   TO '${MYSQL_USER}'@'%';
  GRANT ALL PRIVILEGES ON websocket.* TO '${MYSQL_USER}'@'%';
  FLUSH PRIVILEGES;
EOSQL

echo "Applying ERP migrations..."
for f in $(ls /migrations/erp/*.up.sql 2>/dev/null | sort); do
  echo "  -> $f"
  mysql -u root -p"${MYSQL_ROOT_PASSWORD}" erp < "$f"
done

echo "Applying Takeout migrations..."
for f in $(ls /migrations/takeout/*.up.sql 2>/dev/null | sort); do
  echo "  -> $f"
  mysql -u root -p"${MYSQL_ROOT_PASSWORD}" takeout < "$f"
done

echo "Applying Message migrations..."
for f in $(ls /migrations/message/*.up.sql 2>/dev/null | sort); do
  echo "  -> $f"
  mysql -u root -p"${MYSQL_ROOT_PASSWORD}" message < "$f"
done

echo "Applying WebSocket migrations..."
for f in $(ls /migrations/websocket/*.up.sql 2>/dev/null | sort); do
  echo "  -> $f"
  mysql -u root -p"${MYSQL_ROOT_PASSWORD}" websocket < "$f"
done

echo "BMP database initialization complete."
