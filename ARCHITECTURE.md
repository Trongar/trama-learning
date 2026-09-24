# Trama — arquitectura actual

## Servidor

- PocketBase 0.40.4 embebido en un binario Go 1.27; SQLite y migraciones propias.
- `main.go`: inicializa PocketBase, sirve la web estática, crea una ruta inicial al registrar cada cuenta y expone evaluación demo autenticada.
- `migrations/`: define las reglas y los campos de `learning_goals`; reusa la colección Auth `users` incluida por PocketBase.
- `TRAMA_DATA_DIR`: directorio de datos. Localmente `./pb_data`; contenedor `/data`.
- `web/public/`: interfaz HTML/CSS/JavaScript servida por el propio PocketBase.

## Identidad y separación de datos

- El registro/inicio de sesión usa los endpoints Auth integrados de PocketBase. El bearer token se guarda por pestaña en `sessionStorage`; la página lo renueva con `auth-refresh`.
- Cada registro `learning_goals` relaciona el propietario `user`; la ruta inicial, pasos, respuestas y mastery se guardan en ese registro.
- Las reglas PB limitan listar/ver/editar/borrar a `user = @request.auth.id`. La creación exige que el propietario enviado sea el usuario autenticado; la regla de edición impide transferir la propiedad.
- Una cuenta nueva recibe una copia independiente de la ruta inicial de tres pasos. La creación por hook es verificada; el cliente también crea la ruta si no existe.
- `POST /api/trama/goals/{id}/attempts` exige Auth y compara el propietario antes de actualizar la ruta dentro de una transacción. Un ID de otra persona devuelve 404.
- Para invitar testers, comparte el enlace público y pide a cada persona crear su propia cuenta. Si comparten una cuenta, compartirán sus datos.

## Rutas y API

- `GET /`: aplicación web estática con registro e inicio de sesión.
- PocketBase Auth: `/api/collections/users/records`, `/auth-with-password`, `/auth-refresh`.
- PocketBase Records: `/api/collections/learning_goals/records` con reglas de propietario.
- `POST /api/trama/goals/{id}/attempts`: evalúa y persiste la respuesta/progreso, solo para la cuenta dueña.
- `GET /healthz`: health check.

## Persistencia y despliegue

El Dockerfile fija `TRAMA_DATA_DIR=/data` y declara `/data` como volumen; no ejecutes contenedores reemplazando ese volumen. Las pruebas locales confirman que la ruta sobrevive a un reinicio del proceso. No hay backup automatizado aún. Antes de datos sensibles hacen falta backups, correo/recovery y hardening adicional.

PocketBase 0.40.4 aún es pre-1.0: su documentación oficial indica que no se recomienda para sistemas críticos sin seguir los cambios de compatibilidad y migraciones.

## Límites del MVP

Las lecciones son scaffolds de demostración. El evaluador es `mock · heuristic-demo-v1`, no una integración de IA. No hay SMTP/verificación por email, recuperación de contraseña, OAuth, funcionalidades de voz ni aplicación de todas las colecciones previstas en `infra/pocketbase-schema.json`.
