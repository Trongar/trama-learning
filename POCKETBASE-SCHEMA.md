# PocketBase schema: piloto Trama

## Implementado hoy

PocketBase 0.40.4 crea la colección Auth `users` del framework. La migración de Trama limita listar/ver/editar/borrar a la propia cuenta y deja abierto el registro público necesario para invitar testers.

`learning_goals` es una colección Base con propietario `user`, dominio, nivel, propósito, minutos/semana, estado, progreso y un campo JSON `path` que guarda los tres pasos de aprendizaje, respuestas y señales de mastery. Las reglas list/view/delete requieren `user = @request.auth.id`; la creación exige `@request.body.user = @request.auth.id`; la edición exige propietario y prohíbe cambiar el campo `user`. El endpoint de intentos exige Auth y vuelve a comprobar el dueño.

El registro inicial crea una ruta privada “Aprender a aprender” por cada cuenta. Los usuarios no comparten la cuenta ni el registro inicial.

## Decisión de alcance

El contrato de datos original proponía 17 colecciones. Para el piloto no se crearon tablas vacías ni relaciones innecesarias: `learning_paths`, `topics`, `concepts`, `concept_dependencies`, `lessons`, `exercises`, `assessments`, `attempts`, `mastery`, `review_schedule`, `sources`, `sessions`, `messages`, `media_assets` y `pronunciation_attempts` quedan planeadas. Hoy el contenido corto y el estado de mastery viven dentro del JSON `path`; una futura iteración puede normalizarlo mediante migraciones cuando exista una necesidad validada.

`infra/pocketbase-schema.json` distingue las colecciones implementadas de las reservadas. Nunca agregar credenciales o tokens a colecciones de producto.
