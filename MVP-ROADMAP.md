# MVP roadmap

## Fase 1 — recorrido demostrativo (completada)
Meta → ruta de tres pasos → lección → respuesta → retroalimentación heurística y progreso.

## Fase 2 — cuentas y aislamiento inicial (completada para piloto)
PocketBase Auth, almacenamiento SQLite, migraciones, regla de propietario en `learning_goals`, ruta inicial independiente por cuenta, endpoint de intentos con comprobación del dueño y pruebas de aislamiento con dos cuentas. Persistencia local probada tras reiniciar PocketBase.

Pendientes antes de producción: verificar volumen persistente en el host Dokploy y restauración desde backup, configurar SMTP/verificación y recuperación de cuenta, políticas de retención/exportación y controles contra abuso del registro.

## Fase 3 — aprendizaje generado y evaluado
Integrar el proveedor OpenCode detrás del contrato confirmado. No inventar endpoints ni modelos; mantener el mock claramente etiquetado hasta configurar una credencial segura. Añadir contenidos específicos y fuentes trazables.

## Fase 4 — práctica adaptativa
Quizzes variados, repetición espaciada, adaptación al desempeño y revisión de calidad del contenido.

## Fuera del MVP
Voz, cursos masivos, comunidad, marketplace y analítica avanzada.

## Gate de invitación y despliegue
Cada tester crea su propia cuenta; el registro genera su propia ruta inicial. No se deben compartir credenciales ni ingresar información sensible. Las reglas de PocketBase bloquean lecturas y actualizaciones entre cuentas y fueron verificadas contra el API real local. El preview es solo para feedback; PocketBase 0.40.4 es pre-1.0 y los backups/recuperación todavía no están configurados.
