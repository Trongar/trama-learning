# Trama — arquitectura actual

## Aplicación web

- Backend: Go `net/http` y `html/template`.
- Interfaz: HTML semántico, HTMX como mejora progresiva y CSS propio; no requiere un build frontend.
- Plantillas y CSS se incluyen en el binario mediante `embed.FS`.
- `main.go`: rutas HTTP, validación, store temporal y evaluador de demostración.
- `main_test.go`: recorrido E2E local y validaciones de entrada.
- `infra/pocketbase-schema.json`: contrato preliminar de las colecciones objetivo; todavía no se aplica a un servidor PocketBase.

## Rutas web

- `GET /`: formulario para crear un objetivo.
- `POST /goals`: valida la meta y crea un mapa inicial de tres conceptos.
- `GET /goals/{id}`: muestra el mapa y el progreso.
- `GET /concepts/{id}`: muestra la lección y el ejercicio.
- `POST /exercises/{id}/attempts`: evalúa con el adaptador mock y actualiza la señal de dominio.
- `GET /healthz`: health check del proceso.

## Límites

El almacenamiento actual es memoria de proceso; no hay aislamiento por cuenta ni persistencia entre reinicios. Las lecciones son scaffolds de demostración y el evaluador heurístico no sustituye a OpenCode. Un preview autorizado sirve solo para probar la interfaz: no ingresar datos personales o sensibles. Para un servicio de producción hacen falta autenticación, propiedad de datos y persistencia.

## Próxima arquitectura

Usar PocketBase para autenticación y registros, con relaciones y reglas de acceso por propietario. El backend debe mantener credenciales/proveedores fuera del navegador y hacer la integración de OpenCode solo tras confirmar su contrato. Diseñar voz como adaptadores independientes de STT, TTS y evaluación de pronunciación.
